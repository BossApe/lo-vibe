package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"musuhi-api/internal/model"
	"musuhi-api/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProjectService struct {
	mock.Mock
}

func (m *mockProjectService) ExtractFeatures(ctx context.Context, overviewID string) (*model.ProjectExtraction, error) {
	args := m.Called(ctx, overviewID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectExtraction), args.Error(1)
}

func (m *mockProjectService) SuggestName(ctx context.Context, overviewID string) (*model.ProjectNameSuggestion, error) {
	args := m.Called(ctx, overviewID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectNameSuggestion), args.Error(1)
}

func (m *mockProjectService) InitDirectory(ctx context.Context, projectName, localPath, template string) (*model.ProjectInitResult, error) {
	args := m.Called(ctx, projectName, localPath, template)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectInitResult), args.Error(1)
}

func (m *mockProjectService) CreateRepositoryWithExternal(ctx context.Context, owner, repoName, visibility, localPath, commitMessage string) (*model.ProjectWithExternalResult, error) {
	args := m.Called(ctx, owner, repoName, visibility, localPath, commitMessage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectWithExternalResult), args.Error(1)
}

func TestProjectHandler_ExtractFeatures_有効な概要IDから機能一覧と構成要素を抽出する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	h := NewProjectHandler(svc)
	overviewID := uuid.New().String()
	svc.On("ExtractFeatures", mock.Anything, overviewID).Return(
		&model.ProjectExtraction{Features: []string{"ユーザ管理"}, Components: []string{"Backend API"}},
		nil,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/extract-features", bytes.NewBufferString(`{"overviewId":"`+overviewID+`"}`))
	rec := httptest.NewRecorder()

	h.ExtractFeatures(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.NotEmpty(t, data["features"])
	svc.AssertExpectations(t)
}

func TestProjectHandler_SuggestName_概要IDを空で指定してプロジェクト名候補を取得する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	h := NewProjectHandler(svc)
	svc.On("SuggestName", mock.Anything, "").Return(nil, fmt.Errorf("%w: overviewId is required", service.ErrValidation))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/suggest-name", bytes.NewBufferString(`{"overviewId":""}`))
	rec := httptest.NewRecorder()

	h.SuggestName(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestProjectHandler_InitDirectory_有効な入力で初期ディレクトリを作成する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	h := NewProjectHandler(svc)
	id := uuid.New()
	svc.On("InitDirectory", mock.Anything, "demo-project", "/tmp", "default").Return(
		&model.ProjectInitResult{ID: id, DirectoryStatus: "success"}, nil,
	)

	body := `{"projectName":"demo-project","localPath":"/tmp","template":"default"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/init-directory", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.InitDirectory(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "success", data["directoryStatus"])
	svc.AssertExpectations(t)
}

func TestProjectHandler_SuggestName_有効な概要IDからプロジェクト名候補を取得する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	h := NewProjectHandler(svc)
	overviewID := uuid.New().String()
	svc.On("SuggestName", mock.Anything, overviewID).Return(
		&model.ProjectNameSuggestion{
			Candidates: []string{"sukunahikona", "watatsumi"},
			Items: []model.ProjectNameCandidate{
				{Name: "sukunahikona", Reason: "旅行支援に合う神名", AISuggested: true},
				{Name: "watatsumi", Reason: "移動の広がりを表す神名", AISuggested: true},
			},
		},
		nil,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/suggest-name", bytes.NewBufferString(`{"overviewId":"`+overviewID+`"}`))
	rec := httptest.NewRecorder()

	h.SuggestName(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.NotEmpty(t, data["candidates"])
	items := data["items"].([]any)
	assert.NotEmpty(t, items)
	assert.Equal(t, "sukunahikona", items[0].(map[string]any)["name"])
	svc.AssertExpectations(t)
}

func TestProjectHandler_InitDirectory_不正なプロジェクト名で初期ディレクトリを作成する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	h := NewProjectHandler(svc)
	svc.On("InitDirectory", mock.Anything, "bad name!", "/tmp", "default").Return(
		nil, fmt.Errorf("%w: projectName must match pattern", service.ErrValidation),
	)

	body := `{"projectName":"bad name!","localPath":"/tmp","template":"default"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/init-directory", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.InitDirectory(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestProjectHandler_ExtractFeatures_存在しない概要IDから機能一覧を抽出する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	h := NewProjectHandler(svc)
	overviewID := uuid.New().String()
	svc.On("ExtractFeatures", mock.Anything, overviewID).Return(
		nil, fmt.Errorf("%w: system_overview not found", service.ErrNotFound),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/extract-features", bytes.NewBufferString(`{"overviewId":"`+overviewID+`"}`))
	rec := httptest.NewRecorder()

	h.ExtractFeatures(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestProjectHandler_WithExternal_有効な入力でGitHubリポジトリ作成と初回pushを実行する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	h := NewProjectHandler(svc)

	svc.On("CreateRepositoryWithExternal", mock.Anything, "BossApe", "demo-project", "private", "/tmp/demo-project", "initial commit").Return(
		&model.ProjectWithExternalResult{
			RepositoryURL:     "https://github.com/BossApe/demo-project",
			ExternalProjectID: "123456",
			PushStatus:        "success",
		}, nil,
	)

	body := `{"owner":"BossApe","repoName":"demo-project","visibility":"private","localPath":"/tmp/demo-project","commitMessage":"initial commit"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/with-external", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.WithExternal(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "success", data["pushStatus"])
	assert.Equal(t, "https://github.com/BossApe/demo-project", data["repositoryUrl"])
	svc.AssertExpectations(t)
}

func TestProjectHandler_WithExternal_ownerを空で指定してGitHubリポジトリ作成を実行する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	h := NewProjectHandler(svc)

	svc.On("CreateRepositoryWithExternal", mock.Anything, "", "demo-project", "private", "/tmp/demo-project", "initial commit").Return(
		nil, fmt.Errorf("%w: owner is required", service.ErrValidation),
	)

	body := `{"owner":"","repoName":"demo-project","visibility":"private","localPath":"/tmp/demo-project","commitMessage":"initial commit"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/with-external", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.WithExternal(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	svc.AssertExpectations(t)
}
