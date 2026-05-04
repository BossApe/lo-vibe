package source

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
		{ID: "#10", Number: 10, Title: "IT0-1: Iteration", Type: "Iteration", Phase: "Phase 0", Iteration: "1", Parent: "IT0-1", Difficulty: 0},
		{ID: "#16", Number: 16, Title: "TK-1-1: first", Type: "Ticket", Parent: "IT0-1", Estimate: "L", Difficulty: 5},
		{ID: "#17", Number: 17, Title: "TK-1-2: second", Type: "Ticket", Parent: "IT0-1", Estimate: "M", Difficulty: 3},
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
				makeIssueNode(10, "IT0-1: Iteration", "https://example.com/10", "BossApe/Musuhi", map[string]string{"Type": "Iteration", "Phase": "Phase 0", "Iteration": "1", "Parent": "IT0-1"}),
				makeIssueNode(16, "TK-1-1: first", "https://example.com/16", "BossApe/Musuhi", map[string]string{"Type": "Ticket", "Parent": "IT0-1", "Estimate": "L"}),
				makeIssueNode(17, "TK-1-2: second", "https://example.com/17", "BossApe/Musuhi", map[string]string{"Type": "Ticket", "Parent": "IT0-1", "Estimate": "M"}),
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
	if !strings.Contains(text, "Issue難易度ロードマップ") {
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
