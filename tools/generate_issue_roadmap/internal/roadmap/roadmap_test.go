package roadmap

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildOutputFileName(t *testing.T) {
	ts := time.Date(2026, 5, 3, 12, 34, 56, 0, time.FixedZone("JST", 9*60*60))
	got := buildOutputFileName("Musuhi/Project:v2", ts)
	want := "Musuhi_Project_v2進捗_20260503123456.html"
	if got != want {
		t.Fatalf("buildOutputFileName() = %q, want %q", got, want)
	}
}

func TestBuildLanes(t *testing.T) {
	rows := []issueRow{
		{ID: "#10", Number: 10, Title: "SP0-1: Sprint", Type: "Sprint", Phase: "Phase 0", Sprint: "1", Parent: "SP0-1", Difficulty: 0},
		{ID: "#16", Number: 16, Title: "TK-1-1: first", Type: "Ticket", Parent: "SP0-1", Estimate: "L", Difficulty: 5},
		{ID: "#17", Number: 17, Title: "TK-1-2: second", Type: "Ticket", Parent: "SP0-1", Estimate: "M", Difficulty: 3},
	}

	lanes := buildLanes(rows)
	if len(lanes) != 1 {
		t.Fatalf("len(lanes) = %d, want 1", len(lanes))
	}
	if lanes[0].TotalDifficulty != 8 {
		t.Fatalf("TotalDifficulty = %d, want 8", lanes[0].TotalDifficulty)
	}
	if len(lanes[0].Tickets) != 2 {
		t.Fatalf("len(Tickets) = %d, want 2", len(lanes[0].Tickets))
	}
	if lanes[0].Tickets[0].Start != 0 || lanes[0].Tickets[0].End != 5 {
		t.Fatalf("first ticket range = %d-%d, want 0-5", lanes[0].Tickets[0].Start, lanes[0].Tickets[0].End)
	}
	if lanes[0].Tickets[1].Start != 5 || lanes[0].Tickets[1].End != 8 {
		t.Fatalf("second ticket range = %d-%d, want 5-8", lanes[0].Tickets[1].Start, lanes[0].Tickets[1].End)
	}
}

func TestRunWritesHTML(t *testing.T) {
	tmp := t.TempDir()
	fixedNow := time.Date(2026, 5, 3, 1, 2, 3, 0, time.UTC)

	fakeFetcher := func(cfg config) (*projectData, error) {
		return &projectData{
			Title: "Musuhi 開発",
			Items: []projectItemNode{
				makeIssueNode(10, "SP0-1: Sprint", "https://example.com/10", "BossApe/Musuhi", map[string]string{"Type": "Sprint", "Phase": "Phase 0", "Sprint": "1", "Parent": "SP0-1"}),
				makeIssueNode(16, "TK-1-1: first", "https://example.com/16", "BossApe/Musuhi", map[string]string{"Type": "Ticket", "Parent": "SP0-1", "Estimate": "L"}),
				makeIssueNode(17, "TK-1-2: second", "https://example.com/17", "BossApe/Musuhi", map[string]string{"Type": "Ticket", "Parent": "SP0-1", "Estimate": "M"}),
			},
		}, nil
	}

	var stdout bytes.Buffer
	err := Run([]string{
		"-owner", "BossApe",
		"-repo", "Musuhi",
		"-project-number", "2",
		"-output-dir", tmp,
	}, &stdout, fakeFetcher, func() time.Time { return fixedNow })
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	wantFile := filepath.Join(tmp, "Musuhi 開発進捗_20260503010203.html")
	b, err := os.ReadFile(wantFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	text := string(b)
	if !strings.Contains(text, "Musuhi 開発進捗") {
		t.Fatalf("html does not contain title")
	}
	if !strings.Contains(text, "TK-1-1") || !strings.Contains(text, "TK-1-2") {
		t.Fatalf("html does not contain ticket rows")
	}
	if !strings.Contains(stdout.String(), "出力:") {
		t.Fatalf("stdout does not contain output path: %s", stdout.String())
	}
}

func makeIssueNode(number int, title, url, repo string, fields map[string]string) projectItemNode {
	n := projectItemNode{}
	n.Content.Number = number
	n.Content.Title = title
	n.Content.URL = url
	n.Content.Repository.NameWithOwner = repo

	for name, value := range fields {
		fv := fieldValueNode{}
		fv.Field.Name = name
		fv.Name = value
		n.FieldValues.Nodes = append(n.FieldValues.Nodes, fv)
	}
	return n
}

func TestStoryPointDifficulty(t *testing.T) {
	tests := []struct {
		name   string
		spText string
		est    string
		want   int
	}{
		{"SP設定あり(3)", "3", "M", 3},
		{"SP設定あり(8)", "8", "XS", 8},
		{"SP=0は無視→Estimateフォールバック", "0", "S", 2},
		{"SP未設定→Estimate L=5", "", "L", 5},
		{"SP非数値→Estimate XS=1", "abc", "XS", 1},
		{"SP未設定・Estimate未知→デフォルト1", "", "UNKNOWN", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storyPointDifficulty(tt.spText, tt.est)
			if got != tt.want {
				t.Fatalf("storyPointDifficulty(%q, %q) = %d, want %d", tt.spText, tt.est, got, tt.want)
			}
		})
	}
}

func TestNormalizeRowsStoryPoint(t *testing.T) {
	project := &projectData{
		Title: "Test Project",
		Items: []projectItemNode{
			// Story Point 設定あり → SP値を使用
			makeIssueNode(1, "TK-1-1: task with SP", "https://example.com/1", "BossApe/Musuhi",
				map[string]string{"Type": "Ticket", "Story Point": "7", "Estimate": "M"}),
			// Story Point 未設定 → Estimate フォールバック
			makeIssueNode(2, "TK-1-2: task without SP", "https://example.com/2", "BossApe/Musuhi",
				map[string]string{"Type": "Ticket", "Estimate": "L"}),
		},
	}
	cfg := config{Owner: "BossApe", Repo: "Musuhi", ProjectNumber: 2}
	rows := normalizeRows(project, cfg)

	if len(rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(rows))
	}
	if rows[0].Difficulty != 7 {
		t.Fatalf("row[0].Difficulty = %d, want 7 (Story Point優先)", rows[0].Difficulty)
	}
	if rows[1].Difficulty != 5 {
		t.Fatalf("row[1].Difficulty = %d, want 5 (Estimate L フォールバック)", rows[1].Difficulty)
	}
}
