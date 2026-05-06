package service

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"musuhi-api/internal/model"
	"musuhi-api/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var projectNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ProjectService は FR-002 のビジネスロジックインターフェース。
type ProjectService interface {
	ExtractFeatures(ctx context.Context, overviewID string) (*model.ProjectExtraction, error)
	SuggestName(ctx context.Context, overviewID string) (*model.ProjectNameSuggestion, error)
	InitDirectory(ctx context.Context, projectName, localPath, template string) (*model.ProjectInitResult, error)
}

type projectService struct {
	overviewRepo repository.SystemOverviewRepository
}

// NewProjectService は ProjectService を生成する。
func NewProjectService(overviewRepo repository.SystemOverviewRepository) ProjectService {
	return &projectService{overviewRepo: overviewRepo}
}

func (s *projectService) ExtractFeatures(ctx context.Context, overviewID string) (*model.ProjectExtraction, error) {
	content, err := s.loadOverviewContent(ctx, overviewID)
	if err != nil {
		return nil, err
	}

	features := extractFeatureCandidates(content)
	components := extractComponentCandidates(content)

	return &model.ProjectExtraction{Features: features, Components: components}, nil
}

func (s *projectService) SuggestName(ctx context.Context, overviewID string) (*model.ProjectNameSuggestion, error) {
	content, err := s.loadOverviewContent(ctx, overviewID)
	if err != nil {
		return nil, err
	}

	candidates := suggestProjectNameCandidates(content)
	return &model.ProjectNameSuggestion{Candidates: candidates}, nil
}

func (s *projectService) InitDirectory(_ context.Context, projectName, localPath, template string) (*model.ProjectInitResult, error) {
	if strings.TrimSpace(projectName) == "" {
		return nil, fmt.Errorf("%w: projectName is required", ErrValidation)
	}
	if !projectNamePattern.MatchString(projectName) {
		return nil, fmt.Errorf("%w: projectName must match %s", ErrValidation, projectNamePattern.String())
	}
	if strings.TrimSpace(localPath) == "" {
		return nil, fmt.Errorf("%w: localPath is required", ErrValidation)
	}
	if !filepath.IsAbs(localPath) {
		return nil, fmt.Errorf("%w: localPath must be absolute path", ErrValidation)
	}
	if template == "" {
		template = "default"
	}
	if template != "default" {
		return nil, fmt.Errorf("%w: unsupported template", ErrValidation)
	}

	root := filepath.Join(localPath, projectName)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("projectService.InitDirectory: %w", err)
	}

	dirs := []string{
		"_document/000.進捗状況",
		"_document/001.提案・要求仕様フェーズ",
		"_document/002.要件定義フェーズ",
		"_document/003.設計・開発・テストフェーズ",
		"_document/004.リリース・運用フェーズ",
		"services",
		"tools",
	}

	for _, d := range dirs {
		fullDir := filepath.Join(root, d)
		if err := os.MkdirAll(fullDir, 0o755); err != nil {
			return nil, fmt.Errorf("projectService.InitDirectory: %w", err)
		}
		keepPath := filepath.Join(fullDir, ".keep")
		if err := os.WriteFile(keepPath, []byte(""), 0o644); err != nil {
			return nil, fmt.Errorf("projectService.InitDirectory: %w", err)
		}
	}

	readmePath := filepath.Join(root, "README.md")
	readme := "# " + projectName + "\n\nMusuhi FR-002 によって生成されたプロジェクトです。\n"
	if err := os.WriteFile(readmePath, []byte(readme), 0o644); err != nil {
		return nil, fmt.Errorf("projectService.InitDirectory: %w", err)
	}

	return &model.ProjectInitResult{
		ID:              uuid.New(),
		DirectoryStatus: "success",
	}, nil
}

func (s *projectService) loadOverviewContent(ctx context.Context, overviewID string) (string, error) {
	id, err := uuid.Parse(overviewID)
	if err != nil {
		return "", fmt.Errorf("%w: invalid overviewId format", ErrValidation)
	}

	overview, err := s.overviewRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%w: system_overview id=%s", ErrNotFound, overviewID)
		}
		return "", fmt.Errorf("projectService.loadOverviewContent: %w", err)
	}

	return overview.Content, nil
}

func extractFeatureCandidates(content string) []string {
	items := tokenizeLines(content)
	if len(items) == 0 {
		return []string{"主要機能の定義"}
	}

	features := make([]string, 0, len(items))
	for _, item := range items {
		if strings.Contains(item, "機能") || strings.Contains(item, "管理") || strings.Contains(item, "表示") {
			features = append(features, item)
			continue
		}
		features = append(features, item+"機能")
	}

	return uniqueInOrder(features)
}

func extractComponentCandidates(content string) []string {
	c := strings.ToLower(content)
	components := make([]string, 0, 5)

	if containsAny(c, []string{"ui", "画面", "frontend", "svelte"}) {
		components = append(components, "Frontend UI")
	}
	if containsAny(c, []string{"api", "backend", "go", "サーバ"}) {
		components = append(components, "Backend API")
	}
	if containsAny(c, []string{"db", "database", "postgres", "データベース"}) {
		components = append(components, "RDB")
	}
	if containsAny(c, []string{"queue", "worker", "ジョブ", "batch"}) {
		components = append(components, "Worker")
	}
	if containsAny(c, []string{"auth", "認証", "login"}) {
		components = append(components, "Auth")
	}

	if len(components) == 0 {
		components = []string{"Frontend UI", "Backend API", "RDB"}
	}

	return uniqueInOrder(components)
}

func suggestProjectNameCandidates(content string) []string {
	base := detectProjectBaseName(content)
	hash := crc32.ChecksumIEEE([]byte(content))
	suffix := fmt.Sprintf("%04x", hash%0x10000)

	candidates := []string{
		base + "-" + suffix, // コンテンツ固有のサフィックスを持つ候補を先頭に置く
		base + "-core",
		base + "-app",
	}

	out := uniqueInOrder(candidates)
	sort.Strings(out[1:]) // 先頭の固有候補はそのまま、残りをアルファベット順に整列
	return out
}

func detectProjectBaseName(content string) string {
	lower := strings.ToLower(content)
	switch {
	case containsAny(lower, []string{"book", "書籍", "本"}):
		return "book-hub"
	case containsAny(lower, []string{"task", "todo", "タスク"}):
		return "task-flow"
	case containsAny(lower, []string{"在庫", "inventory"}):
		return "inventory-core"
	case containsAny(lower, []string{"予約", "booking"}):
		return "booking-app"
	default:
		return "musuhi-project"
	}
}

func tokenizeLines(content string) []string {
	lines := strings.Split(content, "\n")
	items := make([]string, 0, len(lines))
	for _, line := range lines {
		v := strings.TrimSpace(line)
		v = strings.TrimPrefix(v, "- ")
		v = strings.TrimPrefix(v, "*")
		v = strings.TrimPrefix(v, "・")
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		items = append(items, v)
	}
	return items
}

func containsAny(base string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(base, k) {
			return true
		}
	}
	return false
}

func uniqueInOrder(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
