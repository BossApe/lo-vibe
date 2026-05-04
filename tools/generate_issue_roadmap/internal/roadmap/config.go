package roadmap

import (
	"errors"
	"flag"
	"io"
	"path/filepath"
)

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("generate_issue_roadmap", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := config{}
	fs.StringVar(&cfg.Owner, "owner", defaultOwner, "GitHub owner (user or organization)")
	fs.StringVar(&cfg.Repo, "repo", defaultRepo, "GitHub repository name")
	fs.IntVar(&cfg.ProjectNumber, "project-number", defaultProjectNo, "GitHub Project (v2) number")
	fs.StringVar(&cfg.OutputDir, "output-dir", defaultOutputDir(), "出力先ディレクトリ")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if cfg.Owner == "" || cfg.Repo == "" {
		return config{}, errors.New("設定の初期化失敗: owner/repo は必須です")
	}
	if cfg.ProjectNumber <= 0 {
		return config{}, errors.New("設定の初期化失敗: project-number は1以上で指定してください")
	}
	return cfg, nil
}

func defaultOutputDir() string {
	return filepath.Join("..", "..", "_document", "003.設計・開発・テストフェーズ", "002.開発進捗")
}
