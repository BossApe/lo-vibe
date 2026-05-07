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
	// 書籍キーワードが含まれるので先頭候補は "shoseki"
	assert.Equal(t, "shoseki", got.Candidates[0])
	assert.NotEmpty(t, got.Items)
	assert.Equal(t, "shoseki", got.Items[0].Name)
	assert.False(t, got.Items[0].AISuggested)
	assert.Empty(t, got.Items[0].Reason)
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
	assert.Equal(t, "musuhi-project", got.Candidates[0])
	repo.AssertExpectations(t)
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
