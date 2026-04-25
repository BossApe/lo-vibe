package source

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
