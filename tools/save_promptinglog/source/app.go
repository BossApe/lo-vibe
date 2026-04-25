package source

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var sensitiveValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`github_pat_[A-Za-z0-9_]+`),
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]+`),
}

type entry struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp string          `json:"timestamp"`
}

type sessionStartData struct {
	SessionID      string `json:"sessionId"`
	StartTime      string `json:"startTime"`
	CopilotVersion string `json:"copilotVersion"`
	VSCodeVersion  string `json:"vscodeVersion"`
}

type messageData struct {
	Content string `json:"content"`
}

type role string

const (
	roleUser      role = "user"
	roleAssistant role = "assistant"
)

type message struct {
	Role      role
	Content   string
	Timestamp time.Time
}

type sessionInfo struct {
	SessionID      string
	StartTime      time.Time
	CopilotVersion string
	VSCodeVersion  string
	Messages       []message
}

type config struct {
	SessionID  string
	DaysAgo    int
	OutputDir  string
	StorageDir string
}

type transcriptCandidate struct {
	path    string
	modTime time.Time
}

const (
	outputFilePrefix    = "promptinglog_"
	outputFileExtension = ".md"
	outputTimeLayout    = "20060102150405"
)

func Main() {
	if err := Run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func Run(args []string, stdout io.Writer) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("出力ディレクトリ作成失敗: %w", err)
	}

	transcriptPath, err := resolveTranscriptPath(cfg)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "トランスクリプト: %s\n", transcriptPath)

	info, err := loadSessionInfo(transcriptPath)
	if err != nil {
		return err
	}

	baseTime, skipMessage, err := applyMessageSelection(cfg, info)
	if err != nil {
		return err
	}
	if skipMessage != "" {
		fmt.Fprintln(stdout, skipMessage)
		return nil
	}

	outputPath := buildOutputPath(cfg.OutputDir, info, baseTime)

	if err := os.WriteFile(outputPath, []byte(renderMarkdown(info)), 0o644); err != nil {
		return fmt.Errorf("ファイル書き込み失敗: %w", err)
	}

	fmt.Fprintf(stdout, "保存完了: %s\n", outputPath)
	fmt.Fprintf(stdout, "ターン数: %d\n", countTurns(info.Messages))
	return nil
}

func applyMessageSelection(cfg config, info *sessionInfo) (time.Time, string, error) {
	if cfg.DaysAgo >= 0 {
		return selectMessagesByDaysAgo(cfg.DaysAgo, info)
	}

	if cfg.SessionID != "" {
		return sessionBaseTime(info), "", nil
	}

	return selectMessagesIncremental(cfg.OutputDir, info)
}

func selectMessagesByDaysAgo(daysAgo int, info *sessionInfo) (time.Time, string, error) {
	filtered, basedOn, targetDay, err := filterMessagesByDaysAgo(info.Messages, daysAgo, nowFunc())
	if err != nil {
		return time.Time{}, "", err
	}
	if len(filtered) == 0 {
		return time.Time{}, fmt.Sprintf("指定日(%s)のログはありません", targetDay.Format("2006-01-02")), nil
	}

	info.Messages = filtered
	return basedOn.In(jstZone), "", nil
}


func selectMessagesIncremental(outputDir string, info *sessionInfo) (time.Time, string, error) {
	baseTime := sessionBaseTime(info)
	filtered, checkpoint, found, err := filterMessagesSinceLastSave(outputDir, info.Messages)
	if err != nil {
		return time.Time{}, "", err
	}
	if found {
		if len(filtered) == 0 {
			return time.Time{}, "新規ログはありません", nil
		}
		info.Messages = filtered
		return checkpoint.In(jstZone), "", nil
	}

	return baseTime, "", nil
}

func filterMessagesSinceLastSave(outputDir string, messages []message) ([]message, time.Time, bool, error) {
	boundary, found, err := latestSavedLogTime(outputDir)
	if err != nil {
		return nil, time.Time{}, false, fmt.Errorf("前回保存ログの取得失敗: %w", err)
	}
	if !found {
		return messages, time.Time{}, false, nil
	}

	filtered := filterMessagesAfter(messages, boundary)

	if len(filtered) == 0 {
		return filtered, boundary, true, nil
	}

	return filtered, lastMessageTimestamp(filtered, boundary), true, nil
}

func filterMessagesAfter(messages []message, boundary time.Time) []message {
	var filtered []message
	for _, msg := range messages {
		if msg.Timestamp.IsZero() {
			continue
		}
		if msg.Timestamp.After(boundary) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

func filterMessagesByDaysAgo(messages []message, daysAgo int, now time.Time) ([]message, time.Time, time.Time, error) {
	if daysAgo < 0 {
		return nil, time.Time{}, time.Time{}, errors.New("-d は 0 以上の整数で指定してください")
	}

	nowJST := now.In(jstZone)
	targetDay := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, jstZone).AddDate(0, 0, -daysAgo)
	nextDay := targetDay.AddDate(0, 0, 1)

	var filtered []message
	for _, msg := range messages {
		if msg.Timestamp.IsZero() {
			continue
		}
		ts := msg.Timestamp.In(jstZone)
		if !ts.Before(targetDay) && ts.Before(nextDay) {
			filtered = append(filtered, msg)
		}
	}

	if len(filtered) == 0 {
		return filtered, targetDay, targetDay, nil
	}

	return filtered, lastMessageTimestamp(filtered, targetDay), targetDay, nil
}

func lastMessageTimestamp(messages []message, fallback time.Time) time.Time {
	last := messages[len(messages)-1].Timestamp
	if last.IsZero() {
		return fallback
	}
	return last
}

func latestSavedLogTime(outputDir string) (time.Time, bool, error) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return time.Time{}, false, err
	}

	latest := time.Time{}
	found := false
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		timestamp, ok := savedLogTimeFromFilename(e.Name())
		if !ok {
			info, err := e.Info()
			if err != nil {
				continue
			}
			timestamp = info.ModTime()
		}

		if !found || timestamp.After(latest) {
			latest = timestamp
			found = true
		}
	}

	return latest, found, nil
}

func savedLogTimeFromFilename(name string) (time.Time, bool) {
	if !strings.HasPrefix(name, outputFilePrefix) || !strings.HasSuffix(name, outputFileExtension) {
		return time.Time{}, false
	}

	base := strings.TrimSuffix(strings.TrimPrefix(name, outputFilePrefix), outputFileExtension)
	if idx := strings.Index(base, "_"); idx >= 0 {
		base = base[:idx]
	}

	if len(base) != len(outputTimeLayout) {
		return time.Time{}, false
	}

	t, err := time.ParseInLocation(outputTimeLayout, base, jstZone)
	if err != nil {
		return time.Time{}, false
	}

	return t, true
}

func parseConfig(args []string) (config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config{}, fmt.Errorf("設定の初期化失敗: ホームディレクトリの取得失敗: %w", err)
	}

	defaultStorageDir := filepath.Join(
		homeDir,
		"Library", "Application Support", "Code", "User", "workspaceStorage",
	)

	fs := flag.NewFlagSet("save_promptinglog", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var cfg config
	fs.StringVar(&cfg.SessionID, "session-id", "", "保存するセッション ID (省略時は最新)")
	fs.IntVar(&cfg.DaysAgo, "d", -1, "何日前のログを保存するか。例: -d2 は2日前")
	fs.StringVar(&cfg.OutputDir, "output-dir", resolveDefaultOutputDir(), "出力先ディレクトリ")
	fs.StringVar(&cfg.StorageDir, "storage-dir", defaultStorageDir, "workspaceStorage のルートパス")
	normalized := normalizeArgs(args)
	if err := fs.Parse(normalized); err != nil {
		return config{}, fmt.Errorf("設定の初期化失敗: %w", err)
	}
	if err := validateDaysAgo(fs, cfg.DaysAgo); err != nil {
		return config{}, err
	}

	return cfg, nil
}

func validateDaysAgo(fs *flag.FlagSet, daysAgo int) error {
	daysAgoSpecified := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "d" {
			daysAgoSpecified = true
		}
	})
	if daysAgoSpecified && daysAgo < 0 {
		return errors.New("設定の初期化失敗: -d は 0 以上の整数で指定してください")
	}
	if daysAgo < -1 {
		return errors.New("設定の初期化失敗: -d は 0 以上の整数で指定してください")
	}
	return nil
}

func normalizeArgs(args []string) []string {
	normalized := make([]string, 0, len(args)+1)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-d") && len(arg) > 2 {
			suffix := arg[2:]
			if _, err := strconv.Atoi(suffix); err == nil {
				normalized = append(normalized, "-d", suffix)
				continue
			}
		}
		normalized = append(normalized, arg)
	}
	return normalized
}

func loadSessionInfo(transcriptPath string) (*sessionInfo, error) {
	info, err := parseTranscript(transcriptPath)
	if err != nil {
		return nil, fmt.Errorf("パース失敗: %w", err)
	}
	if info.SessionID == "" {
		info.SessionID = strings.TrimSuffix(filepath.Base(transcriptPath), ".jsonl")
	}
	return info, nil
}

func sessionBaseTime(info *sessionInfo) time.Time {
	if !info.StartTime.IsZero() {
		return info.StartTime.In(jstZone)
	}
	return nowFunc().In(jstZone)
}

func parseTranscript(path string) (*sessionInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	info := &sessionInfo{}
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 1<<20)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var e entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, e.Timestamp)

		switch e.Type {
		case "session.start":
			applySessionStart(info, e.Data)
		case "user.message":
			appendMessage(info, roleUser, e.Data, ts)
		case "assistant.message":
			appendMessage(info, roleAssistant, e.Data, ts)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan transcript: %w", err)
	}
	return info, nil
}

func applySessionStart(info *sessionInfo, raw json.RawMessage) {
	var d sessionStartData
	if err := json.Unmarshal(raw, &d); err != nil {
		return
	}
	info.SessionID = d.SessionID
	info.CopilotVersion = d.CopilotVersion
	info.VSCodeVersion = d.VSCodeVersion
	if t, err := time.Parse(time.RFC3339Nano, d.StartTime); err == nil {
		info.StartTime = t
	}
}

func appendMessage(info *sessionInfo, msgRole role, raw json.RawMessage, ts time.Time) {
	var d messageData
	if err := json.Unmarshal(raw, &d); err != nil {
		return
	}
	if strings.TrimSpace(d.Content) == "" {
		return
	}
	info.Messages = append(info.Messages, message{
		Role:      msgRole,
		Content:   d.Content,
		Timestamp: ts,
	})
}

var jstZone = time.FixedZone("JST", 9*60*60)

var nowFunc = time.Now

func renderMarkdown(info *sessionInfo) string {
	var sb strings.Builder

	sb.WriteString("# Copilot Prompting Log\n\n")
	writeSessionInfoSection(&sb, info)
	writeConversationSection(&sb, info.Messages)

	return sb.String()
}

func writeSessionInfoSection(sb *strings.Builder, info *sessionInfo) {
	sb.WriteString("## セッション情報\n\n")
	sb.WriteString("| 項目 | 値 |\n")
	sb.WriteString("| --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| セッション ID | %s |\n", escapeTableCell(info.SessionID)))
	if !info.StartTime.IsZero() {
		startJST := info.StartTime.In(jstZone)
		sb.WriteString(fmt.Sprintf("| 開始日時 (JST) | %s |\n", escapeTableCell(startJST.Format("2006-01-02 15:04:05"))))
	}
	sb.WriteString(fmt.Sprintf("| Copilot バージョン | %s |\n", escapeTableCell(info.CopilotVersion)))
	sb.WriteString(fmt.Sprintf("| VS Code バージョン | %s |\n\n", escapeTableCell(info.VSCodeVersion)))
}

func writeConversationSection(sb *strings.Builder, messages []message) {
	sb.WriteString("## 会話ログ\n\n")

	turnNum := 0
	for _, msg := range messages {
		if msg.Role == roleUser {
			turnNum++
			sb.WriteString(fmt.Sprintf("### ターン %d\n\n", turnNum))
		}

		writeMessageSection(sb, msg)
	}
}

func writeMessageSection(sb *strings.Builder, msg message) {
	sb.WriteString(messageHeading(msg))
	sb.WriteString("\n\n")
	sb.WriteString(redactSensitiveText(strings.TrimRight(msg.Content, "\n")))
	sb.WriteString("\n\n")
}

func messageHeading(msg message) string {
	heading := "#### アシスタント"
	if msg.Role == roleUser {
		heading = "#### ユーザー"
	}

	timestamp := formatMessageTimestamp(msg.Timestamp)
	if timestamp == "" {
		return heading
	}

	return fmt.Sprintf("%s (%s)", heading, timestamp)
}

func formatMessageTimestamp(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.In(jstZone).Format("2006-01-02 15:04:05 JST")
}

func escapeTableCell(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", "<br>")
	return value
}

func redactSensitiveText(value string) string {
	redacted := value
	for _, pattern := range sensitiveValuePatterns {
		redacted = pattern.ReplaceAllString(redacted, "[REDACTED]")
	}
	return redacted
}

func findTranscriptDirs(storageRoot string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(storageRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "transcripts" && strings.Contains(path, "GitHub.copilot-chat") {
			dirs = append(dirs, path)
			return fs.SkipDir
		}
		return nil
	})
	return dirs, err
}

func latestTranscript(transcriptDir string) (string, error) {
	files, err := listTranscriptCandidates(transcriptDir)
	if err != nil {
		return "", err
	}
	return latestCandidatePath(files, "no transcript file found")
}

func listTranscriptCandidates(transcriptDir string) ([]transcriptCandidate, error) {
	entries, err := os.ReadDir(transcriptDir)
	if err != nil {
		return nil, err
	}
	var files []transcriptCandidate
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, transcriptCandidate{
			path:    filepath.Join(transcriptDir, e.Name()),
			modTime: info.ModTime(),
		})
	}
	return files, nil
}

func resolveTranscriptPath(cfg config) (string, error) {
	dirs, err := findTranscriptDirs(cfg.StorageDir)
	if err != nil {
		return "", fmt.Errorf("transcripts ディレクトリの検索失敗: %w", err)
	}
	if len(dirs) == 0 {
		return "", errors.New("transcripts ディレクトリが見つかりません")
	}

	if cfg.SessionID != "" {
		return findTranscriptBySessionID(dirs, cfg.SessionID)
	}

	return findLatestTranscript(dirs)
}

func findTranscriptBySessionID(dirs []string, sessionID string) (string, error) {
	for _, dir := range dirs {
		candidate := filepath.Join(dir, sessionID+".jsonl")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("セッション %s のトランスクリプトが見つかりません", sessionID)
}

func findLatestTranscript(dirs []string) (string, error) {
	var candidates []transcriptCandidate
	for _, dir := range dirs {
		files, err := listTranscriptCandidates(dir)
		if err != nil {
			continue
		}
		candidates = append(candidates, files...)
	}
	return latestCandidatePath(candidates, "トランスクリプトファイルが見つかりません")
}

func latestCandidatePath(candidates []transcriptCandidate, emptyMessage string) (string, error) {
	if len(candidates) == 0 {
		return "", errors.New(emptyMessage)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].modTime.After(candidates[j].modTime)
	})
	return candidates[0].path, nil
}

func buildOutputPath(outputDir string, info *sessionInfo, baseTime time.Time) string {
	shortID := ""
	if len(info.SessionID) >= 8 {
		shortID = "_" + info.SessionID[:8]
	}
	filename := fmt.Sprintf("%s%s%s%s", outputFilePrefix, baseTime.Format(outputTimeLayout), shortID, outputFileExtension)
	return filepath.Join(outputDir, filename)
}

func resolveDefaultOutputDir() string {
	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, "..", "..", "_document", "promptinglog")
		if info, err := os.Stat(filepath.Clean(candidate)); err == nil && info.IsDir() {
			return candidate
		}
	}

	execPath, err := os.Executable()
	if err == nil {
		return filepath.Join(filepath.Dir(execPath), "..", "..", "_document", "promptinglog")
	}

	return filepath.Join("..", "..", "_document", "promptinglog")
}

func countTurns(msgs []message) int {
	n := 0
	for _, m := range msgs {
		if m.Role == roleUser {
			n++
		}
	}
	return n
}
