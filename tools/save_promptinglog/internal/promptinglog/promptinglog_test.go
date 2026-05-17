package promptinglog

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupRunDirs(t *testing.T) (string, string, string) {
	t.Helper()

	root := t.TempDir()
	transcriptsDir := filepath.Join(root, "workspace", "GitHub.copilot-chat", "transcripts")
	if err := os.MkdirAll(transcriptsDir, 0o755); err != nil {
		t.Fatalf("mkdir transcripts dir: %v", err)
	}

	outputDir := filepath.Join(root, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output dir: %v", err)
	}

	return root, transcriptsDir, outputDir
}

func writeTranscript(t *testing.T, transcriptsDir, sessionID string, lines []string) string {
	t.Helper()

	transcriptPath := filepath.Join(transcriptsDir, sessionID+".jsonl")
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(transcriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	return transcriptPath
}

func withFixedNow(t *testing.T, fixed time.Time) {
	t.Helper()

	originalNowFunc := nowFunc
	nowFunc = func() time.Time {
		return fixed
	}
	t.Cleanup(func() {
		nowFunc = originalNowFunc
	})
}

func TestParseTranscript(t *testing.T) {
	transcriptPath := filepath.Join(t.TempDir(), "session.jsonl")
	content := strings.Join([]string{
		`{"type":"session.start","timestamp":"2026-04-25T00:33:50Z","data":{"sessionId":"abc12345-session","startTime":"2026-04-25T00:33:50Z","copilotVersion":"0.45.1","vscodeVersion":"1.117.0"}}`,
		`{"type":"user.message","timestamp":"2026-04-25T00:34:10Z","data":{"content":"hello"}}`,
		`{"type":"assistant.message","timestamp":"2026-04-25T00:34:20Z","data":{"content":"world"}}`,
	}, "\n")
	if err := os.WriteFile(transcriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	info, err := parseTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("parseTranscript returned error: %v", err)
	}

	if info.SessionID != "abc12345-session" {
		t.Fatalf("SessionID = %q, want %q", info.SessionID, "abc12345-session")
	}
	if len(info.Messages) != 2 {
		t.Fatalf("message count = %d, want 2", len(info.Messages))
	}
	if info.Messages[0].Role != roleUser || info.Messages[1].Role != roleAssistant {
		t.Fatalf("unexpected message roles: %#v", info.Messages)
	}
}

func TestRenderMarkdownRedactsSensitiveValues(t *testing.T) {
	info := &sessionInfo{
		SessionID:      "session-12345678",
		StartTime:      time.Date(2026, 4, 25, 0, 33, 50, 0, time.UTC),
		CopilotVersion: "0.45.1",
		VSCodeVersion:  "1.117.0",
		Messages: []message{{
			Role:      roleAssistant,
			Timestamp: time.Date(2026, 4, 25, 0, 34, 20, 0, time.UTC),
			Content:   "token=ghp_verysecret and github_pat_supersecret",
		}},
	}

	rendered := renderMarkdown(info)
	if strings.Contains(rendered, "ghp_verysecret") || strings.Contains(rendered, "github_pat_supersecret") {
		t.Fatalf("rendered markdown still contains sensitive values: %s", rendered)
	}
	if !strings.Contains(rendered, "[REDACTED]") {
		t.Fatalf("rendered markdown did not contain redaction marker: %s", rendered)
	}
}

func TestFindLatestTranscript(t *testing.T) {
	dir := t.TempDir()
	older := filepath.Join(dir, "older.jsonl")
	newer := filepath.Join(dir, "newer.jsonl")
	if err := os.WriteFile(older, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write older transcript: %v", err)
	}
	if err := os.WriteFile(newer, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write newer transcript: %v", err)
	}

	olderTime := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	newerTime := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	if err := os.Chtimes(older, olderTime, olderTime); err != nil {
		t.Fatalf("set older modtime: %v", err)
	}
	if err := os.Chtimes(newer, newerTime, newerTime); err != nil {
		t.Fatalf("set newer modtime: %v", err)
	}

	got, err := latestTranscript(dir)
	if err != nil {
		t.Fatalf("latestTranscript returned error: %v", err)
	}
	if got != newer {
		t.Fatalf("latestTranscript = %q, want %q", got, newer)
	}
}

func TestRunWritesMarkdownFile(t *testing.T) {
	root := t.TempDir()
	transcriptsDir := filepath.Join(root, "workspace", "GitHub.copilot-chat", "transcripts")
	if err := os.MkdirAll(transcriptsDir, 0o755); err != nil {
		t.Fatalf("mkdir transcripts dir: %v", err)
	}

	transcriptPath := filepath.Join(transcriptsDir, "abc12345-session.jsonl")
	content := strings.Join([]string{
		`{"type":"session.start","timestamp":"2026-04-25T00:33:50Z","data":{"sessionId":"abc12345-session","startTime":"2026-04-25T00:33:50Z","copilotVersion":"0.45.1","vscodeVersion":"1.117.0"}}`,
		`{"type":"user.message","timestamp":"2026-04-25T00:34:10Z","data":{"content":"hello"}}`,
		`{"type":"assistant.message","timestamp":"2026-04-25T00:34:20Z","data":{"content":"reply with ghp_secretvalue"}}`,
	}, "\n")
	if err := os.WriteFile(transcriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	outputDir := filepath.Join(root, "output")
	var stdout bytes.Buffer
	err := Run([]string{
		"-storage-dir", root,
		"-output-dir", outputDir,
		"-session-id", "abc12345-session",
	}, &stdout)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	outputPath := filepath.Join(outputDir, "promptinglog_20260425093350_abc12345.md")
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "# Copilot Prompting Log") {
		t.Fatalf("output missing header: %s", text)
	}
	if strings.Contains(text, "ghp_secretvalue") {
		t.Fatalf("output still contains secret: %s", text)
	}
	if !strings.Contains(stdout.String(), "保存完了") {
		t.Fatalf("stdout missing completion message: %s", stdout.String())
	}
}

func TestRunWithoutArgsSavesOnlyAfterPreviousSave(t *testing.T) {
	root, transcriptsDir, outputDir := setupRunDirs(t)

	writeTranscript(t, transcriptsDir, "abc12345-session", []string{
		`{"type":"session.start","timestamp":"2026-04-25T00:33:50Z","data":{"sessionId":"abc12345-session","startTime":"2026-04-25T00:33:50Z","copilotVersion":"0.45.1","vscodeVersion":"1.117.0"}}`,
		`{"type":"user.message","timestamp":"2026-04-25T00:34:10Z","data":{"content":"old user message"}}`,
		`{"type":"assistant.message","timestamp":"2026-04-25T00:34:20Z","data":{"content":"new assistant message"}}`,
		`{"type":"user.message","timestamp":"2026-04-25T00:34:30Z","data":{"content":"new user message"}}`,
	})

	previous := filepath.Join(outputDir, "promptinglog_20260425093415_abc12345.md")
	if err := os.WriteFile(previous, []byte("previous"), 0o644); err != nil {
		t.Fatalf("write previous output: %v", err)
	}
	checkpoint := time.Date(2026, 4, 25, 0, 34, 15, 0, time.UTC)
	if err := os.Chtimes(previous, checkpoint, checkpoint); err != nil {
		t.Fatalf("set previous modtime: %v", err)
	}

	var stdout bytes.Buffer
	err := Run([]string{
		"-storage-dir", root,
		"-output-dir", outputDir,
	}, &stdout)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	outputPath := filepath.Join(outputDir, "promptinglog_20260425093430_abc12345.md")
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	text := string(data)
	if strings.Contains(text, "old user message") {
		t.Fatalf("output still contains old message: %s", text)
	}
	if !strings.Contains(text, "new assistant message") || !strings.Contains(text, "new user message") {
		t.Fatalf("output missing new messages: %s", text)
	}
}

func TestRunWithoutArgsSkipsWhenNoNewMessages(t *testing.T) {
	root, transcriptsDir, outputDir := setupRunDirs(t)

	writeTranscript(t, transcriptsDir, "abc12345-session", []string{
		`{"type":"session.start","timestamp":"2026-04-25T00:33:50Z","data":{"sessionId":"abc12345-session","startTime":"2026-04-25T00:33:50Z","copilotVersion":"0.45.1","vscodeVersion":"1.117.0"}}`,
		`{"type":"user.message","timestamp":"2026-04-25T00:34:10Z","data":{"content":"old user message"}}`,
		`{"type":"assistant.message","timestamp":"2026-04-25T00:34:20Z","data":{"content":"old assistant message"}}`,
	})

	previous := filepath.Join(outputDir, "promptinglog_20260425093500_abc12345.md")
	if err := os.WriteFile(previous, []byte("previous"), 0o644); err != nil {
		t.Fatalf("write previous output: %v", err)
	}

	var stdout bytes.Buffer
	err := Run([]string{
		"-storage-dir", root,
		"-output-dir", outputDir,
	}, &stdout)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "promptinglog_*.md"))
	if err != nil {
		t.Fatalf("glob output files: %v", err)
	}
	if len(files) != 1 || files[0] != previous {
		t.Fatalf("unexpected output files: %#v", files)
	}
	if !strings.Contains(stdout.String(), "新規ログはありません") {
		t.Fatalf("stdout missing no-new-log message: %s", stdout.String())
	}
}

func TestSavedLogTimeFromFilename(t *testing.T) {
	timestamp, ok := savedLogTimeFromFilename("promptinglog_20260425183000_abc12345.md")
	if !ok {
		t.Fatalf("expected valid filename")
	}
	want := time.Date(2026, 4, 25, 18, 30, 0, 0, jstZone)
	if !timestamp.Equal(want) {
		t.Fatalf("timestamp = %v, want %v", timestamp, want)
	}

	if _, ok := savedLogTimeFromFilename("promptinglog_invalid.md"); ok {
		t.Fatalf("expected invalid filename to fail")
	}
}

func TestLatestSavedLogTimeFallsBackToModTime(t *testing.T) {
	outputDir := t.TempDir()

	valid := filepath.Join(outputDir, "promptinglog_20260425120000_abc12345.md")
	if err := os.WriteFile(valid, []byte("valid"), 0o644); err != nil {
		t.Fatalf("write valid file: %v", err)
	}

	legacy := filepath.Join(outputDir, "promptinglog_legacy.md")
	if err := os.WriteFile(legacy, []byte("legacy"), 0o644); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}
	legacyTime := time.Date(2026, 4, 25, 13, 0, 0, 0, jstZone)
	if err := os.Chtimes(legacy, legacyTime, legacyTime); err != nil {
		t.Fatalf("set legacy modtime: %v", err)
	}

	got, found, err := latestSavedLogTime(outputDir)
	if err != nil {
		t.Fatalf("latestSavedLogTime returned error: %v", err)
	}
	if !found {
		t.Fatalf("latestSavedLogTime did not find files")
	}
	if !got.Equal(legacyTime) {
		t.Fatalf("latestSavedLogTime = %v, want %v", got, legacyTime)
	}
}

func TestParseConfigSupportsCompactDaysAgoFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"-d2"})
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}
	if cfg.DaysAgo != 2 {
		t.Fatalf("DaysAgo = %d, want 2", cfg.DaysAgo)
	}
}

func TestParseConfigSupportsEqualStyleDaysAgoFlag(t *testing.T) {
	cfg, err := parseConfig([]string{"-d=2"})
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}
	if cfg.DaysAgo != 2 {
		t.Fatalf("DaysAgo = %d, want 2", cfg.DaysAgo)
	}
}

func TestRunWithDaysAgoSavesOnlyTargetDayMessages(t *testing.T) {
	root, transcriptsDir, outputDir := setupRunDirs(t)

	writeTranscript(t, transcriptsDir, "abc12345-session", []string{
		`{"type":"session.start","timestamp":"2026-04-23T23:00:00Z","data":{"sessionId":"abc12345-session","startTime":"2026-04-23T23:00:00Z","copilotVersion":"0.45.1","vscodeVersion":"1.117.0"}}`,
		`{"type":"user.message","timestamp":"2026-04-23T14:00:00Z","data":{"content":"old day message"}}`,
		`{"type":"assistant.message","timestamp":"2026-04-24T01:00:00Z","data":{"content":"target day message 1"}}`,
		`{"type":"user.message","timestamp":"2026-04-24T02:00:00Z","data":{"content":"target day message 2"}}`,
		`{"type":"assistant.message","timestamp":"2026-04-25T01:00:00Z","data":{"content":"new day message"}}`,
	})

	withFixedNow(t, time.Date(2026, 4, 26, 10, 0, 0, 0, jstZone))

	var stdout bytes.Buffer
	err := Run([]string{
		"-storage-dir", root,
		"-output-dir", outputDir,
		"-session-id", "abc12345-session",
		"-d2",
	}, &stdout)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	outputPath := filepath.Join(outputDir, "promptinglog_20260424110000_abc12345.md")
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	text := string(data)
	if strings.Contains(text, "old day message") || strings.Contains(text, "new day message") {
		t.Fatalf("output contains non-target day messages: %s", text)
	}
	if !strings.Contains(text, "target day message 1") || !strings.Contains(text, "target day message 2") {
		t.Fatalf("output missing target day messages: %s", text)
	}
}

func TestRunWithDaysAgoSkipsWhenNoTargetDayMessages(t *testing.T) {
	root, transcriptsDir, outputDir := setupRunDirs(t)

	writeTranscript(t, transcriptsDir, "abc12345-session", []string{
		`{"type":"session.start","timestamp":"2026-04-23T23:00:00Z","data":{"sessionId":"abc12345-session","startTime":"2026-04-23T23:00:00Z","copilotVersion":"0.45.1","vscodeVersion":"1.117.0"}}`,
		`{"type":"user.message","timestamp":"2026-04-23T14:00:00Z","data":{"content":"old day message"}}`,
	})

	withFixedNow(t, time.Date(2026, 4, 26, 10, 0, 0, 0, jstZone))

	var stdout bytes.Buffer
	err := Run([]string{
		"-storage-dir", root,
		"-output-dir", outputDir,
		"-session-id", "abc12345-session",
		"-d=2",
	}, &stdout)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "promptinglog_*.md"))
	if err != nil {
		t.Fatalf("glob output files: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("unexpected output files: %#v", files)
	}
	if !strings.Contains(stdout.String(), "指定日(2026-04-24)のログはありません") {
		t.Fatalf("stdout missing no-target-day message: %s", stdout.String())
	}
}

func TestApplyMessageSelectionDaysAgoTakesPrecedence(t *testing.T) {
	info := &sessionInfo{
		SessionID: "abc12345-session",
		Messages: []message{
			{Role: roleUser, Timestamp: time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC), Content: "old"},
			{Role: roleAssistant, Timestamp: time.Date(2026, 4, 24, 1, 0, 0, 0, time.UTC), Content: "target"},
		},
	}

	withFixedNow(t, time.Date(2026, 4, 26, 10, 0, 0, 0, jstZone))

	baseTime, skipMessage, err := applyMessageSelection(config{SessionID: "abc12345-session", DaysAgo: 2}, info)
	if err != nil {
		t.Fatalf("applyMessageSelection returned error: %v", err)
	}
	if skipMessage != "" {
		t.Fatalf("skipMessage = %q, want empty", skipMessage)
	}
	if len(info.Messages) != 1 || info.Messages[0].Content != "target" {
		t.Fatalf("unexpected filtered messages: %#v", info.Messages)
	}
	wantBaseTime := time.Date(2026, 4, 24, 10, 0, 0, 0, jstZone)
	if !baseTime.Equal(wantBaseTime) {
		t.Fatalf("baseTime = %v, want %v", baseTime, wantBaseTime)
	}
}

func TestParseConfigRejectsNegativeDaysAgo(t *testing.T) {
	_, err := parseConfig([]string{"-d=-2"})
	if err == nil {
		t.Fatalf("expected error for negative daysAgo")
	}
}

func TestParseConfigRejectsExplicitMinusOneDaysAgo(t *testing.T) {
	_, err := parseConfig([]string{"-d=-1"})
	if err == nil {
		t.Fatalf("expected error for explicit -d=-1")
	}
}
