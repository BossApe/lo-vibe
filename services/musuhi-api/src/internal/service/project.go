package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

type systemNameCandidate struct {
	keywords []string
	romaji   string
}

type themedNameCandidate struct {
	keywords []string
	names    []string
	reasons  map[string]string
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

	items := suggestProjectNameCandidates(content)
	candidates := make([]string, 0, len(items))
	for _, item := range items {
		candidates = append(candidates, item.Name)
	}

	return &model.ProjectNameSuggestion{Candidates: candidates, Items: items}, nil
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

	root := filepath.Clean(localPath)
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

var systemNameRomaji = []systemNameCandidate{
	{keywords: []string{"書籍", "本棚", "book", "ブックマン", "bookman"}, romaji: "shoseki"},
	{keywords: []string{"図書", "ライブラリ", "library"}, romaji: "tosho"},
	{keywords: []string{"タスク", "todo", "task"}, romaji: "tasuku"},
	{keywords: []string{"在庫", "inventory", "stock"}, romaji: "zaiko"},
	{keywords: []string{"予約", "booking", "reservation", "リザーブ"}, romaji: "yoyaku"},
	{keywords: []string{"会員", "membership", "メンバー"}, romaji: "kaiin"},
	{keywords: []string{"注文", "order", "カート"}, romaji: "chumon"},
	{keywords: []string{"給与", "payroll", "salary", "給料"}, romaji: "kyuyo"},
	{keywords: []string{"勤怠", "attendance", "出退勤"}, romaji: "kintai"},
	{keywords: []string{"決済", "payment", "課金"}, romaji: "kessai"},
	{keywords: []string{"物流", "logistics", "配送", "shipping"}, romaji: "butsuryuu"},
	{keywords: []string{"人事", "hr", "採用", "recruit"}, romaji: "jinji"},
	{keywords: []string{"musuhi", "むすひ", "ムスヒ"}, romaji: "musuhi"},
}

var themeGodNames = []themedNameCandidate{
	{
		keywords: []string{"書籍", "本", "知識", "学習", "library", "book"},
		names:    []string{"amenominakanushi", "futsunushi", "kuninosazuchi"},
		reasons: map[string]string{
			"amenominakanushi": "知識や構想の起点となる主宰神にちなみました。\n書籍や学習を支える基盤サービスのイメージに近い名前です。",
			"futsunushi":       "判断力と整序の象徴として選んだ候補です。\n本や情報を整理し扱う機能を持つサービス名として収まりが良いです。",
			"kuninosazuchi":    "土台づくりの意味を持たせた候補です。\n知識基盤を育てるシステムの由来として使えます。",
		},
	},
	{
		keywords: []string{"商品", "売買", "販売", "ec", "カタログ"},
		names:    []string{"okuninushi", "ebisu", "daikoku"},
		reasons: map[string]string{
			"okuninushi": "縁結びと国づくりで知られ、商いとの親和性も高い神名です。\n商品や顧客をつなぐサービスの中心名として採用しています。",
			"ebisu":      "福徳と商売繁盛の象徴として自然な候補です。\n販売やカタログ系のサービスに直感的な由来を持たせられます。",
			"daikoku":    "豊穣と財に結びつく神名です。\nECや販売管理の候補名として分かりやすい響きです。",
		},
	},
	{
		keywords: []string{"タスク", "todo", "生産", "プロジェクト", "計画"},
		names:    []string{"futodama", "amenokoyane", "amenofutodama"},
		reasons: map[string]string{
			"futodama":      "祭祀や段取りを司る存在として、進行管理の連想がしやすい名前です。\nタスクや計画を整えるサービスの由来に向いています。",
			"amenokoyane":   "言葉や秩序を支える神格から、計画立案の意味を込めています。\nプロジェクト推進の補助役としての印象を持たせられます。",
			"amenofutodama": "準備と運営を支える役割にちなむ候補です。\n実務を着実に前へ進める管理サービス名として使えます。",
		},
	},
	{
		keywords: []string{"旅行", "地図", "観光", "travel", "map"},
		names:    []string{"sukunahikona", "watatsumi", "urashima"},
		reasons: map[string]string{
			"sukunahikona": "旅や知恵に結びつく存在として知られる名です。\n旅行計画や観光支援サービスに、軽やかで知的な印象を与えます。",
			"watatsumi":    "海路や移動の広がりを想起しやすい神名です。\n移動や旅程を扱うサービスのスケール感を表現できます。",
			"urashima":     "旅の物語性を連想しやすい伝承由来の候補です。\n観光体験や周遊提案を重視するサービスに合います。",
		},
	},
	{
		keywords: []string{"認証", "auth", "セキュリティ", "login", "会員"},
		names:    []string{"takemikazuchi", "futsunushi", "raijin"},
		reasons: map[string]string{
			"takemikazuchi": "強さと守りを象徴する神名から採りました。\n認証やセキュリティの中核サービスに防御的な印象を与えられます。",
			"futsunushi":    "秩序維持と判断のニュアンスを持つ候補です。\n会員認証や権限管理の候補として自然です。",
			"raijin":        "強い防壁や警戒の印象を持たせやすい名前です。\n外部からの脅威を防ぐサービスの候補名として使えます。",
		},
	},
	{
		keywords: []string{"在庫", "物流", "倉庫", "logistics"},
		names:    []string{"okuninushi", "kotoshironushi", "kuebiko"},
		reasons: map[string]string{
			"okuninushi":     "多くの物事を調停し流れを整える象徴として扱えます。\n在庫や物流全体を束ねるシステム名に向きます。",
			"kotoshironushi": "判断と配分のイメージを持たせやすい候補です。\n在庫配置や供給判断を支えるサービスに合います。",
			"kuebiko":        "現場把握の知恵を連想させる候補です。\n倉庫や在庫状況を把握するサービスに相性が良いです。",
		},
	},
	{
		keywords: []string{"医療", "健康", "health", "病院", "薬"},
		names:    []string{"onamuchi", "sukunahikona", "kagamitsukuri"},
		reasons: map[string]string{
			"onamuchi":      "医療や救済の伝承と結びつく候補です。\n健康支援サービスに穏やかな由来を与えられます。",
			"sukunahikona":  "医療知識と旅の知恵を持つ存在として親和性があります。\nヘルスケア領域のAI提案名として説得力があります。",
			"kagamitsukuri": "支援ツールや診療補助を連想させる候補です。\n補助的な医療サービス名として使えます。",
		},
	},
	{
		keywords: []string{"農業", "食料", "食品", "farm", "food"},
		names:    []string{"ukemochi", "inari", "toyo-uke"},
		reasons: map[string]string{
			"ukemochi": "食物を司る神名で、食品や供給管理に直接つながる候補です。\n農業・食料系のサービスに由来が明確です。",
			"inari":    "収穫と繁栄の象徴として広く認知されるため、親しみやすい候補です。\n農業・流通のサービス名として覚えやすさがあります。",
			"toyo-uke": "豊かな食の供給を支える意味を持たせています。\n食関連システムの基盤名として安定感があります。",
		},
	},
	{
		keywords: []string{"金融", "会計", "決済", "finance", "payment"},
		names:    []string{"kanayamahiko", "kanayamahime", "ebisu"},
		reasons: map[string]string{
			"kanayamahiko": "金融基盤らしい堅さを出せる候補です。\n会計や決済を扱うサービス名として収まりが良いです。",
			"kanayamahime": "財や産業の循環を支える印象を持たせる候補です。\n決済や会計の支援サービスに柔らかい響きがあります。",
			"ebisu":        "商売繁盛の連想が強く、金融・決済でも理解されやすい候補です。\n利用者にとって覚えやすい名前です。",
		},
	},
	{
		keywords: []string{"教育", "school", "学校", "学習"},
		names:    []string{"amenominakanushi", "amenokoyane", "kuninosazuchi"},
		reasons: map[string]string{
			"amenominakanushi": "学びの起点や全体設計を連想させる候補です。\n教育サービスの基盤名として広がりを持たせられます。",
			"amenokoyane":      "言葉と知の体系化を想起しやすい名前です。\n学習支援や教育設計のサービスに向いています。",
			"kuninosazuchi":    "基礎づくりや育成のイメージを持たせる候補です。\n教育基盤を支えるシステムの由来として使えます。",
		},
	},
}

func suggestProjectNameCandidates(content string) []model.ProjectNameCandidate {
	lower := strings.ToLower(content)

	for _, entry := range systemNameRomaji {
		if containsAny(lower, entry.keywords) {
			base := entry.romaji
			return []model.ProjectNameCandidate{
				{Name: base, AISuggested: false},
				{Name: base + "-core", AISuggested: false},
				{Name: base + "-app", AISuggested: false},
			}
		}
	}

	for _, entry := range themeGodNames {
		if containsAny(lower, entry.keywords) {
			out := make([]model.ProjectNameCandidate, 0, len(entry.names))
			for _, name := range entry.names {
				out = append(out, model.ProjectNameCandidate{
					Name:        name,
					Reason:      entry.reasons[name],
					AISuggested: true,
				})
			}
			return uniqueProjectNameCandidates(out)
		}
	}

	return []model.ProjectNameCandidate{
		{Name: "musuhi-project", AISuggested: false},
		{Name: "amenominakanushi", Reason: "全体を束ねる起点という意味合いから AI が補助候補として選びました。\n要件がまだ粗い段階でも、基盤的なプロジェクト名として扱いやすい名前です。", AISuggested: true},
		{Name: "okuninushi", Reason: "多様な要素をまとめ上げる象徴として AI が選んだ候補です。\n用途が広いサービスに対して、柔軟で親しみやすい由来を持たせられます。", AISuggested: true},
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

func uniqueProjectNameCandidates(values []model.ProjectNameCandidate) []model.ProjectNameCandidate {
	seen := map[string]struct{}{}
	out := make([]model.ProjectNameCandidate, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v.Name]; ok {
			continue
		}
		seen[v.Name] = struct{}{}
		out = append(out, v)
	}
	return out
}
