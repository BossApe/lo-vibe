// save_promptinglog は GitHub Copilot Chat のトランスクリプト (JSONL) を読み込み、
// AsciiDoc 形式のプロンプトログとして指定ディレクトリに保存するバッチツールです。
//
// 使い方:
//
//	go run main.go [flags]
//
// フラグ:
//
//	-session-id   保存するセッション ID (省略時は最新セッション)
//	-output-dir   出力先ディレクトリ (省略時は ../../_document/promptinglog)
//	-storage-dir  workspaceStorage のルートパス (省略時は自動検出)
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ---------- JSONL エントリ ----------

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

// ---------- 会話メッセージ ----------

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

// ---------- セッション情報 ----------

type sessionInfo struct {
	SessionID      string
	StartTime      time.Time
	CopilotVersion string
	VSCodeVersion  string
	Messages       []message
}

// ---------- パース ----------

func parseTranscript(path string) (*sessionInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()

	info := &sessionInfo{}
	scanner := bufio.NewScanner(f)
	// 大きなメッセージに対応するためバッファを拡張
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
			var d sessionStartData
			if err := json.Unmarshal(e.Data, &d); err == nil {
				info.SessionID = d.SessionID
				info.CopilotVersion = d.CopilotVersion
				info.VSCodeVersion = d.VSCodeVersion
				if t, err := time.Parse(time.RFC3339Nano, d.StartTime); err == nil {
					info.StartTime = t
				}
			}
		case "user.message":
			var d messageData
			if err := json.Unmarshal(e.Data, &d); err == nil && strings.TrimSpace(d.Content) != "" {
				info.Messages = append(info.Messages, message{
					Role:      roleUser,
					Content:   d.Content,
					Timestamp: ts,
				})
			}
		case "assistant.message":
			var d messageData
			if err := json.Unmarshal(e.Data, &d); err == nil && strings.TrimSpace(d.Content) != "" {
				info.Messages = append(info.Messages, message{
					Role:      roleAssistant,
					Content:   d.Content,
					Timestamp: ts,
				})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan transcript: %w", err)
	}
	return info, nil
}

// ---------- AsciiDoc 生成 ----------

func renderAsciidoc(info *sessionInfo) string {
	var sb strings.Builder

	jst := time.FixedZone("JST", 9*60*60)
	startJST := info.StartTime.In(jst)

	sb.WriteString("= Copilot Prompting Log\n")
	sb.WriteString(":toc:\n")
	sb.WriteString(":toc-title: 目次\n")
	sb.WriteString(":sectnums:\n")
	sb.WriteString("\n")

	// セッション情報
	sb.WriteString("== セッション情報\n\n")
	sb.WriteString("[cols=\"1,3\"]\n")
	sb.WriteString("|===\n")
	sb.WriteString("|項目|値\n\n")
	sb.WriteString(fmt.Sprintf("|セッション ID|%s\n", info.SessionID))
	sb.WriteString(fmt.Sprintf("|開始日時 (JST)|%s\n", startJST.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("|Copilot バージョン|%s\n", info.CopilotVersion))
	sb.WriteString(fmt.Sprintf("|VS Code バージョン|%s\n", info.VSCodeVersion))
	sb.WriteString("|===\n\n")

	// 会話ログ
	sb.WriteString("== 会話ログ\n\n")

	turnNum := 0
	for i, msg := range info.Messages {
		if msg.Role == roleUser {
			turnNum++
			sb.WriteString(fmt.Sprintf("=== ターン %d\n\n", turnNum))
		}

		tsStr := ""
		if !msg.Timestamp.IsZero() {
			tsStr = msg.Timestamp.In(jst).Format("2006-01-02 15:04:05 JST")
		}

		switch msg.Role {
		case roleUser:
			sb.WriteString("==== ユーザー")
			if tsStr != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", tsStr))
			}
			sb.WriteString("\n\n")
		case roleAssistant:
			sb.WriteString("==== アシスタント")
			if tsStr != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", tsStr))
			}
			sb.WriteString("\n\n")
		}

		// コンテンツをそのまま記述ブロックに入れてエスケープ問題を回避
		sb.WriteString("....\n")
		sb.WriteString(msg.Content)
		if !strings.HasSuffix(msg.Content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("....\n\n")
		_ = i
	}

	return sb.String()
}

// ---------- workspaceStorage 検索 ----------

// findTranscriptDir は ~/Library/Application Support/Code/User/workspaceStorage 以下から
// GitHub.copilot-chat/transcripts ディレクトリを持つパスをすべて返す。
func findTranscriptDirs(storageRoot string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(storageRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 読めないディレクトリはスキップ
		}
		if d.IsDir() && d.Name() == "transcripts" &&
			strings.Contains(path, "GitHub.copilot-chat") {
			dirs = append(dirs, path)
			return fs.SkipDir
		}
		return nil
	})
	return dirs, err
}

// latestTranscript は transcriptDir の中で最も更新時刻が新しい .jsonl ファイルのパスを返す。
func latestTranscript(transcriptDir string) (string, error) {
	entries, err := os.ReadDir(transcriptDir)
	if err != nil {
		return "", err
	}
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			path:    filepath.Join(transcriptDir, e.Name()),
			modTime: info.ModTime(),
		})
	}
	if len(files) == 0 {
		return "", errors.New("no transcript file found")
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})
	return files[0].path, nil
}

// ---------- main ----------

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: ホームディレクトリの取得失敗: %v\n", err)
		os.Exit(1)
	}

	defaultStorageDir := filepath.Join(
		homeDir,
		"Library", "Application Support", "Code", "User", "workspaceStorage",
	)

	// ツール自身の場所から lo-vibe/_document/promptinglog を推定
	execDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	defaultOutputDir := filepath.Join(execDir, "..", "..", "_document", "promptinglog")

	var (
		flagSessionID  = flag.String("session-id", "", "保存するセッション ID (省略時は最新)")
		flagOutputDir  = flag.String("output-dir", defaultOutputDir, "出力先ディレクトリ")
		flagStorageDir = flag.String("storage-dir", defaultStorageDir, "workspaceStorage のルートパス")
	)
	flag.Parse()

	// 出力先ディレクトリを作成
	if err := os.MkdirAll(*flagOutputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: 出力ディレクトリ作成失敗: %v\n", err)
		os.Exit(1)
	}

	// トランスクリプトファイルを特定
	var transcriptPath string

	if *flagSessionID != "" {
		// セッション ID 指定の場合
		dirs, err := findTranscriptDirs(*flagStorageDir)
		if err != nil || len(dirs) == 0 {
			fmt.Fprintf(os.Stderr, "ERROR: transcripts ディレクトリが見つかりません: %v\n", err)
			os.Exit(1)
		}
		found := false
		for _, dir := range dirs {
			candidate := filepath.Join(dir, *flagSessionID+".jsonl")
			if _, err := os.Stat(candidate); err == nil {
				transcriptPath = candidate
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "ERROR: セッション %s のトランスクリプトが見つかりません\n", *flagSessionID)
			os.Exit(1)
		}
	} else {
		// 最新のトランスクリプトを自動検出
		dirs, err := findTranscriptDirs(*flagStorageDir)
		if err != nil || len(dirs) == 0 {
			fmt.Fprintf(os.Stderr, "ERROR: transcripts ディレクトリが見つかりません: %v\n", err)
			os.Exit(1)
		}
		// 各 transcriptsDir の最新ファイルを候補として収集
		type candidate struct {
			path    string
			modTime time.Time
		}
		var candidates []candidate
		for _, dir := range dirs {
			p, err := latestTranscript(dir)
			if err != nil {
				continue
			}
			info, err := os.Stat(p)
			if err != nil {
				continue
			}
			candidates = append(candidates, candidate{path: p, modTime: info.ModTime()})
		}
		if len(candidates) == 0 {
			fmt.Fprintf(os.Stderr, "ERROR: トランスクリプトファイルが見つかりません\n")
			os.Exit(1)
		}
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].modTime.After(candidates[j].modTime)
		})
		transcriptPath = candidates[0].path
	}

	fmt.Printf("トランスクリプト: %s\n", transcriptPath)

	// パース
	info, err := parseTranscript(transcriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: パース失敗: %v\n", err)
		os.Exit(1)
	}
	if info.SessionID == "" {
		// ファイル名からセッション ID を補完
		base := filepath.Base(transcriptPath)
		info.SessionID = strings.TrimSuffix(base, ".jsonl")
	}

	// AsciiDoc 生成
	content := renderAsciidoc(info)

	// 出力ファイル名
	jst := time.FixedZone("JST", 9*60*60)
	now := time.Now().In(jst)
	filename := fmt.Sprintf("promptinglog_%s.adoc", now.Format("20060102150405"))
	outputPath := filepath.Join(*flagOutputDir, filename)

	// 書き込み
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: ファイル書き込み失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("保存完了: %s\n", outputPath)
	fmt.Printf("ターン数: %d\n", countTurns(info.Messages))
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
