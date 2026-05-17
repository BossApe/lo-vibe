package roadmap

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Main はCLIエントリーポイント
func Main() {
	if err := Run(os.Args[1:], os.Stdout, fetchProjectByGH, time.Now); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

// Run はテスト可能なメイン処理
func Run(args []string, stdout io.Writer, fetcher projectFetcher, now func() time.Time) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}

	project, err := fetcher(cfg)
	if err != nil {
		return err
	}

	rows := normalizeRows(project, cfg)
	if len(rows) == 0 {
		return errors.New("指定プロジェクト内に対象リポジトリのIssueが見つかりません")
	}

	lanes := buildLanes(rows)
	if len(lanes) == 0 {
		return errors.New("Type=Ticket のIssueが見つかりません")
	}

	maxDifficulty, totalTickets := applyPercentages(lanes)
	if maxDifficulty <= 0 {
		maxDifficulty = 1
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("出力ディレクトリ作成失敗: %w", err)
	}

	fileName := buildOutputFileName(project.Title, now())
	outputPath := filepath.Join(cfg.OutputDir, fileName)
	if err := writeHTML(outputPath, htmlData{
		GeneratedAt:      now().Format("2006-01-02 15:04:05 -0700"),
		ProjectTitle:     project.Title,
		Owner:            cfg.Owner,
		Repo:             cfg.Repo,
		ProjectNumber:    cfg.ProjectNumber,
		OutputPath:       outputPath,
		TotalItems:       len(rows),
		TotalTickets:     totalTickets,
		MaxDifficulty:    maxDifficulty,
		DifficultyLegend: "Story Point優先 / フォールバック: XS=1, S=2, M=3, L=5, XL=8",
		Lanes:            lanes,
	}); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "プロジェクト: %s\n", project.Title)
	fmt.Fprintf(stdout, "レーン数: %d / チケット数: %d\n", len(lanes), totalTickets)
	fmt.Fprintf(stdout, "出力: %s\n", outputPath)
	return nil
}
