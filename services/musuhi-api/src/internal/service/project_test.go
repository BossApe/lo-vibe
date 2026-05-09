package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"musuhi-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProjectOverviewRepository struct {
	mock.Mock
}

type mockGitHubClient struct {
	mock.Mock
}

type mockGitHubProjectsClient struct {
	mock.Mock
}

type mockProjectNameSuggester struct {
	mock.Mock
}

func (m *mockGitHubClient) CreateRepositoryAndInitialPush(ctx context.Context, owner, repoName, visibility, localPath, commitMessage string) (*model.ProjectWithExternalResult, error) {
	args := m.Called(ctx, owner, repoName, visibility, localPath, commitMessage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectWithExternalResult), args.Error(1)
}

func (m *mockGitHubProjectsClient) CreateProject(ctx context.Context, owner, title string) (*model.GitHubProjectsResult, error) {
	args := m.Called(ctx, owner, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GitHubProjectsResult), args.Error(1)
}

func (m *mockGitHubProjectsClient) AddPhase0Tasks(ctx context.Context, owner, projectID string) ([]*model.Phase0Task, error) {
	args := m.Called(ctx, owner, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Phase0Task), args.Error(1)
}

func (m *mockProjectNameSuggester) SuggestGodNames(ctx context.Context, overviewContent string) ([]model.ProjectNameCandidate, error) {
	args := m.Called(ctx, overviewContent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ProjectNameCandidate), args.Error(1)
}

func (m *mockProjectOverviewRepository) Create(ctx context.Context, content string) (*model.SystemOverview, error) {
	args := m.Called(ctx, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func (m *mockProjectOverviewRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.SystemOverview, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func (m *mockProjectOverviewRepository) UpdateByID(ctx context.Context, id uuid.UUID, content string) (*model.SystemOverview, error) {
	args := m.Called(ctx, id, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func TestProjectService_ExtractFeatures_有効な概要IDから機能一覧と構成要素を抽出する_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: "- ユーザ管理\n- 商品表示\n- PostgreSQL", CreatedAt: time.Now()},
		nil,
	)

	got, err := svc.ExtractFeatures(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.Contains(t, got.Features, "ユーザ管理")
	assert.NotEmpty(t, got.Components)
	repo.AssertExpectations(t)
}

func TestProjectService_SuggestName_存在しない概要IDからプロジェクト名候補を取得する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	repo.On("FindByID", mock.Anything, overviewID).Return(nil, pgx.ErrNoRows)

	_, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestProjectService_InitDirectory_有効なプロジェクト名と絶対パスで初期ディレクトリを作成する_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)

	tmp := t.TempDir()
	target := filepath.Join(tmp, "demo_project")
	result, err := svc.InitDirectory(context.Background(), "demo_project", target, "default")
	assert.NoError(t, err)
	assert.Equal(t, "success", result.DirectoryStatus)

	_, statErr := os.Stat(filepath.Join(target, "_document/000.進捗状況/.keep"))
	assert.NoError(t, statErr)

	// README.md が生成されていること
	readmeContent, readErr := os.ReadFile(filepath.Join(target, "README.md"))
	assert.NoError(t, readErr)
	assert.Contains(t, string(readmeContent), "demo_project")
}

func TestProjectService_SuggestName_有効な概要IDからプロジェクト名候補を取得する_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: "書籍管理システム\n- ユーザ登録\n- 本棚表示", CreatedAt: time.Now()},
		nil,
	)

	got, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, got.Candidates)
	// 書籍キーワードでも神様候補が先頭に来る
	assert.Equal(t, "amenominakanushi", got.Candidates[0])
	assert.Contains(t, got.Candidates, "shoseki")
	assert.NotEmpty(t, got.Items)
	assert.Equal(t, "amenominakanushi", got.Items[0].Name)
	assert.True(t, got.Items[0].AISuggested)
	assert.NotEmpty(t, got.Items[0].Reason)
	// 候補名は英数字・ハイフン・アンダースコア形式であること
	for _, c := range got.Candidates {
		assert.Regexp(t, `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`, c)
	}
	repo.AssertExpectations(t)
}

func TestProjectService_SuggestName_システム名なしのテーマ概要から神名候補が返る_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	// 旅行キーワード → sukunahikona が先頭に来ること
	// （旅行・観光はsystemNameRomajiに登録されていないためテーマ候補が選ばれる）
	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: "- 旅行プラン提案\n- 観光スポット検索\n- 地図表示", CreatedAt: time.Now()},
		nil,
	)

	got, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, got.Candidates)
	assert.Equal(t, "sukunahikona", got.Candidates[0])
	assert.NotEmpty(t, got.Items)
	assert.Equal(t, "sukunahikona", got.Items[0].Name)
	assert.True(t, got.Items[0].AISuggested)
	assert.NotEmpty(t, got.Items[0].Reason)
	repo.AssertExpectations(t)
}

func TestProjectService_SuggestName_音楽テーマから作曲支援系の神名候補が返る_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: "楽しい曲・アップテンポ・ワルツ調機能等の作曲したい情報を入力するとそれに合わせてAIが作曲し、楽譜を出力するシステム機能", CreatedAt: time.Now()},
		nil,
	)

	got, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, got.Candidates)
	assert.Equal(t, "benzaiten", got.Candidates[0])
	assert.Contains(t, got.Candidates, "amenouzume")
	assert.Contains(t, got.Candidates, "ichikishimahime")
	assert.Equal(t, "benzaiten", got.Items[0].Name)
	assert.True(t, got.Items[0].AISuggested)
	assert.NotEmpty(t, got.Items[0].Reason)
	repo.AssertExpectations(t)
}

func TestProjectService_SuggestName_医療テーマから診断支援系の神名候補が返る_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: "医療機関の診断支援と健康管理を行うシステム", CreatedAt: time.Now()},
		nil,
	)

	got, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, got.Candidates)
	assert.Equal(t, "onamuchi", got.Candidates[0])
	assert.Contains(t, got.Candidates, "sukunahikona")
	assert.Contains(t, got.Candidates, "kuebiko")
	assert.Equal(t, "onamuchi", got.Items[0].Name)
	assert.True(t, got.Items[0].AISuggested)
	assert.NotEmpty(t, got.Items[0].Reason)
	repo.AssertExpectations(t)
}

func TestProjectService_SuggestName_該当キーワードなし時にフォールバック候補が返る_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: "概要が未定義のシステム", CreatedAt: time.Now()},
		nil,
	)

	got, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, got.Candidates)
	assert.Equal(t, "amenominakanushi", got.Candidates[0])
	assert.Equal(t, "musuhi-project", got.Candidates[len(got.Candidates)-1])
	assert.True(t, got.Items[0].AISuggested)
	assert.NotEmpty(t, got.Items[0].Reason)
	repo.AssertExpectations(t)
}

func TestProjectService_SuggestName_在庫管理テーマから資産保全系の神名候補が返る_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: "在庫管理と倉庫最適化を行うシステム", CreatedAt: time.Now()},
		nil,
	)

	got, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, got.Candidates)
	assert.Equal(t, "daikoku", got.Candidates[0])
	assert.Contains(t, got.Candidates, "bishamonten")
	assert.Contains(t, got.Candidates, "kotoshironushi")
	assert.Equal(t, "daikoku", got.Items[0].Name)
	assert.True(t, got.Items[0].AISuggested)
	assert.NotEmpty(t, got.Items[0].Reason)
	repo.AssertExpectations(t)
}

func TestProjectService_SuggestName_テーマ未一致時にAI候補が返る_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	ai := new(mockProjectNameSuggester)
	svc := NewProjectServiceWithNameSuggester(repo, ai)
	overviewID := uuid.New()
	content := "宇宙線を可視化して異常傾向を予測する研究支援システム"

	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: content, CreatedAt: time.Now()},
		nil,
	)
	ai.On("SuggestGodNames", mock.Anything, content).Return(
		[]model.ProjectNameCandidate{
			{Name: "amenohoakari", Reason: "光と顕現の象徴として、観測データの可視化に意味が近い候補です。", AISuggested: true},
			{Name: "takamimusubi", Reason: "複雑な要素を統合し生成するイメージから、予測支援に結びつけやすい候補です。", AISuggested: true},
		},
		nil,
	)

	got, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, got.Candidates)
	assert.Equal(t, "amenohoakari", got.Candidates[0])
	assert.Equal(t, "amenohoakari", got.Items[0].Name)
	assert.True(t, got.Items[0].AISuggested)
	assert.NotEmpty(t, got.Items[0].Reason)
	repo.AssertExpectations(t)
	ai.AssertExpectations(t)
}

func TestProjectService_SuggestName_AI候補が不正形式の時にフォールバック候補が返る_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	ai := new(mockProjectNameSuggester)
	svc := NewProjectServiceWithNameSuggester(repo, ai)
	overviewID := uuid.New()
	content := "量子相関を扱う未知領域の分析支援"

	repo.On("FindByID", mock.Anything, overviewID).Return(
		&model.SystemOverview{ID: overviewID, Content: content, CreatedAt: time.Now()},
		nil,
	)
	ai.On("SuggestGodNames", mock.Anything, content).Return(
		[]model.ProjectNameCandidate{
			{Name: "不正な名前", Reason: "", AISuggested: true},
		},
		nil,
	)

	got, err := svc.SuggestName(context.Background(), overviewID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, got.Candidates)
	assert.Equal(t, "amenominakanushi", got.Candidates[0])
	assert.Equal(t, "amenominakanushi", got.Items[0].Name)
	repo.AssertExpectations(t)
	ai.AssertExpectations(t)
}

func TestProjectService_InitDirectory_相対パスで初期ディレクトリを作成する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)

	_, err := svc.InitDirectory(context.Background(), "demo_project", "relative/path", "default")
	assert.ErrorIs(t, err, ErrValidation)
}

func TestProjectService_ExtractFeatures_UUID形式ではない概要IDから機能一覧を抽出する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)

	_, err := svc.ExtractFeatures(context.Background(), "bad-id")
	assert.ErrorIs(t, err, ErrValidation)
}

func TestProjectService_ExtractFeatures_概要取得時にリポジトリエラーが発生した状態で機能一覧を抽出する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	svc := NewProjectService(repo)
	overviewID := uuid.New()

	repo.On("FindByID", mock.Anything, overviewID).Return(nil, errors.New("db down"))

	_, err := svc.ExtractFeatures(context.Background(), overviewID.String())
	assert.Error(t, err)
}

func TestProjectService_CreateRepositoryWithExternal_有効な入力でGitHubリポジトリ作成と初回pushを実行する_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	gh := new(mockGitHubClient)
	svc := &projectService{overviewRepo: repo, githubClient: gh}

	tmpDir := t.TempDir()
	gh.On("CreateRepositoryAndInitialPush", mock.Anything, "BossApe", "demo-project", "private", tmpDir, "initial commit").Return(
		&model.ProjectWithExternalResult{
			RepositoryURL:     "https://github.com/BossApe/demo-project",
			ExternalProjectID: "123456",
			PushStatus:        "success",
		}, nil,
	)

	got, err := svc.CreateRepositoryWithExternal(context.Background(), "BossApe", "demo-project", "private", tmpDir, "initial commit")
	assert.NoError(t, err)
	assert.Equal(t, "success", got.PushStatus)
	assert.Equal(t, "https://github.com/BossApe/demo-project", got.RepositoryURL)
	gh.AssertExpectations(t)
}

func TestProjectService_CreateRepositoryWithExternal_ownerを空で指定してGitHubリポジトリ作成を実行する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	gh := new(mockGitHubClient)
	svc := &projectService{overviewRepo: repo, githubClient: gh}

	tmpDir := t.TempDir()
	_, err := svc.CreateRepositoryWithExternal(context.Background(), "", "demo-project", "private", tmpDir, "initial commit")
	assert.ErrorIs(t, err, ErrValidation)
}

func TestProjectService_CreateRepositoryWithExternal_localPathが相対パスの状態でGitHubリポジトリ作成を実行する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	gh := new(mockGitHubClient)
	svc := &projectService{overviewRepo: repo, githubClient: gh}

	_, err := svc.CreateRepositoryWithExternal(context.Background(), "BossApe", "demo-project", "private", "relative/path", "initial commit")
	assert.ErrorIs(t, err, ErrValidation)
}

func TestProjectService_CreateRepositoryWithExternal_GitHub側でリポジトリ名重複エラーが発生する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	gh := new(mockGitHubClient)
	svc := &projectService{overviewRepo: repo, githubClient: gh}

	tmpDir := t.TempDir()
	gh.On("CreateRepositoryAndInitialPush", mock.Anything, "BossApe", "demo-project", "private", tmpDir, "initial commit").Return(
		nil, errors.New("repository name already exists"),
	)

	_, err := svc.CreateRepositoryWithExternal(context.Background(), "BossApe", "demo-project", "private", tmpDir, "initial commit")
	assert.ErrorIs(t, err, ErrValidation)
	gh.AssertExpectations(t)
}

func TestProjectService_CreateGitHubProjects_有効な入力でProjectsボードを作成する_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	ghProjects := new(mockGitHubProjectsClient)
	svc := &projectService{overviewRepo: repo, githubProjectsClient: ghProjects}

	ghProjects.On("CreateProject", mock.Anything, "BossApe", "Musuhi Board").Return(
		&model.GitHubProjectsResult{
			ProjectsURL: "https://github.com/orgs/BossApe/projects/77",
			ProjectsID:  "PVT_test_001",
			Status:      "success",
		}, nil,
	)

	got, err := svc.CreateGitHubProjects(context.Background(), "project-1", "BossApe", "Musuhi Board")
	assert.NoError(t, err)
	assert.Equal(t, "success", got.Status)
	assert.Equal(t, "PVT_test_001", got.ProjectsID)
	ghProjects.AssertExpectations(t)
}

func TestProjectService_CreateGitHubProjects_ownerを空で指定してProjectsボードを作成する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	ghProjects := new(mockGitHubProjectsClient)
	svc := &projectService{overviewRepo: repo, githubProjectsClient: ghProjects}

	_, err := svc.CreateGitHubProjects(context.Background(), "project-1", "", "Musuhi Board")
	assert.ErrorIs(t, err, ErrValidation)
}

func TestProjectService_CreatePhase0Tasks_有効な入力でPhase0タスクを登録する_正常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	ghProjects := new(mockGitHubProjectsClient)
	svc := &projectService{overviewRepo: repo, githubProjectsClient: ghProjects}

	ghProjects.On("AddPhase0Tasks", mock.Anything, "BossApe", "PVT_test_001").Return(
		[]*model.Phase0Task{
			{ID: "PVTI_1", Title: "PH0: 提案・要求仕様・要件定義", Type: "Phase"},
			{ID: "PVTI_2", Title: "SP0-1: 提案・要求仕様作成", Type: "Sprint"},
		}, nil,
	)

	got, err := svc.CreatePhase0Tasks(context.Background(), "project-1", "BossApe", "PVT_test_001")
	assert.NoError(t, err)
	assert.Equal(t, "success", got.Status)
	assert.Len(t, got.Tasks, 2)
	ghProjects.AssertExpectations(t)
}

func TestProjectService_CreatePhase0Tasks_projectsIdを空で指定してPhase0タスクを登録する_異常系(t *testing.T) {
	repo := new(mockProjectOverviewRepository)
	ghProjects := new(mockGitHubProjectsClient)
	svc := &projectService{overviewRepo: repo, githubProjectsClient: ghProjects}

	_, err := svc.CreatePhase0Tasks(context.Background(), "project-1", "BossApe", "")
	assert.ErrorIs(t, err, ErrValidation)
}
