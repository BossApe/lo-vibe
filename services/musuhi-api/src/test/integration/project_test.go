package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"musuhi-api/internal/handler"
	"musuhi-api/internal/middleware"
	"musuhi-api/internal/model"
	"musuhi-api/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockProjectService はインテグレーションテスト用プロジェクトサービスモック。
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

func (m *mockProjectService) CreateGitHubProjects(ctx context.Context, id, owner, title string) (*model.GitHubProjectsResult, error) {
	args := m.Called(ctx, id, owner, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GitHubProjectsResult), args.Error(1)
}

func (m *mockProjectService) CreatePhase0Tasks(ctx context.Context, id, owner, projectsID string) (*model.Phase0TasksResult, error) {
	args := m.Called(ctx, id, owner, projectsID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Phase0TasksResult), args.Error(1)
}

func newProjectTestServer(svc service.ProjectService) *httptest.Server {
	ph := handler.NewProjectHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/projects/extract-features", ph.ExtractFeatures)
	mux.HandleFunc("POST /api/v1/projects/suggest-name", ph.SuggestName)
	mux.HandleFunc("POST /api/v1/projects/init-directory", ph.InitDirectory)
	mux.HandleFunc("POST /api/v1/projects/with-external", ph.WithExternal)
	mux.HandleFunc("POST /api/v1/projects/{id}/github-projects", ph.GitHubProjects)
	mux.HandleFunc("POST /api/v1/projects/{id}/phase0-tasks", ph.Phase0Tasks)
	h := middleware.Logger(middleware.CORS(mux))
	return httptest.NewServer(h)
}

// --- ExtractFeatures ---

func TestIntegration_ExtractFeatures_有効な概要IDから機能一覧と構成要素を抽出する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	overviewID := uuid.New().String()
	svc.On("ExtractFeatures", mock.Anything, overviewID).Return(
		&model.ProjectExtraction{
			Features:   []string{"ユーザ管理", "書籍登録"},
			Components: []string{"Backend API", "PostgreSQL"},
		}, nil,
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := fmt.Sprintf(`{"overviewId":"%s"}`, overviewID)
	res, err := http.Post(srv.URL+"/api/v1/projects/extract-features", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	features := data["features"].([]any)
	assert.NotEmpty(t, features)
	components := data["components"].([]any)
	assert.NotEmpty(t, components)
	svc.AssertExpectations(t)
}

func TestIntegration_ExtractFeatures_概要IDを空で指定して機能一覧を抽出する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	svc.On("ExtractFeatures", mock.Anything, "").Return(
		nil, fmt.Errorf("%w: overviewId is required", service.ErrValidation),
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	res, err := http.Post(srv.URL+"/api/v1/projects/extract-features", "application/json", bytes.NewBufferString(`{"overviewId":""}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "VALIDATION_ERROR", errBody["code"])
	svc.AssertExpectations(t)
}

func TestIntegration_ExtractFeatures_存在しない概要IDから機能一覧を抽出する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	overviewID := uuid.New().String()
	svc.On("ExtractFeatures", mock.Anything, overviewID).Return(
		nil, fmt.Errorf("%w: system_overview not found", service.ErrNotFound),
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := fmt.Sprintf(`{"overviewId":"%s"}`, overviewID)
	res, err := http.Post(srv.URL+"/api/v1/projects/extract-features", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "NOT_FOUND", errBody["code"])
	svc.AssertExpectations(t)
}

func TestIntegration_ExtractFeatures_不正なJSONで機能一覧を抽出する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	res, err := http.Post(srv.URL+"/api/v1/projects/extract-features", "application/json", bytes.NewBufferString(`invalid-json`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "BAD_REQUEST", errBody["code"])
}

// --- SuggestName ---

func TestIntegration_SuggestName_有効な概要IDからプロジェクト名候補を取得する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	overviewID := uuid.New().String()
	svc.On("SuggestName", mock.Anything, overviewID).Return(
		&model.ProjectNameSuggestion{
			Candidates: []string{"sukunahikona", "watatsumi", "urashima"},
			Items: []model.ProjectNameCandidate{
				{Name: "sukunahikona", Reason: "旅行支援に合う神名", AISuggested: true},
				{Name: "watatsumi", Reason: "移動の広がりを表す神名", AISuggested: true},
				{Name: "urashima", Reason: "旅の物語性を持つ伝承名", AISuggested: true},
			},
		},
		nil,
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := fmt.Sprintf(`{"overviewId":"%s"}`, overviewID)
	res, err := http.Post(srv.URL+"/api/v1/projects/suggest-name", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	candidates := data["candidates"].([]any)
	assert.NotEmpty(t, candidates)
	items := data["items"].([]any)
	assert.NotEmpty(t, items)
	assert.Equal(t, "sukunahikona", items[0].(map[string]any)["name"])
	svc.AssertExpectations(t)
}

func TestIntegration_SuggestName_概要IDを空で指定してプロジェクト名候補を取得する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	svc.On("SuggestName", mock.Anything, "").Return(
		nil, fmt.Errorf("%w: overviewId is required", service.ErrValidation),
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	res, err := http.Post(srv.URL+"/api/v1/projects/suggest-name", "application/json", bytes.NewBufferString(`{"overviewId":""}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "VALIDATION_ERROR", errBody["code"])
	svc.AssertExpectations(t)
}

func TestIntegration_SuggestName_存在しない概要IDからプロジェクト名候補を取得する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	overviewID := uuid.New().String()
	svc.On("SuggestName", mock.Anything, overviewID).Return(
		nil, fmt.Errorf("%w: system_overview not found", service.ErrNotFound),
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := fmt.Sprintf(`{"overviewId":"%s"}`, overviewID)
	res, err := http.Post(srv.URL+"/api/v1/projects/suggest-name", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "NOT_FOUND", errBody["code"])
	svc.AssertExpectations(t)
}

func TestIntegration_SuggestName_不正なJSONでプロジェクト名候補を取得する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	res, err := http.Post(srv.URL+"/api/v1/projects/suggest-name", "application/json", bytes.NewBufferString(`invalid-json`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "BAD_REQUEST", errBody["code"])
}

// --- InitDirectory ---

func TestIntegration_InitDirectory_有効な入力で初期ディレクトリを作成する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	id := uuid.New()
	svc.On("InitDirectory", mock.Anything, "demo-project", "/tmp/musuhi/demo-project", "default").Return(
		&model.ProjectInitResult{ID: id, DirectoryStatus: "success"}, nil,
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := `{"projectName":"demo-project","localPath":"/tmp/musuhi/demo-project","template":"default"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/init-directory", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, id.String(), data["id"])
	assert.Equal(t, "success", data["directoryStatus"])
	svc.AssertExpectations(t)
}

func TestIntegration_InitDirectory_不正なプロジェクト名で初期ディレクトリを作成する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	svc.On("InitDirectory", mock.Anything, "bad name!", "/tmp/musuhi/bad-name", "default").Return(
		nil, fmt.Errorf("%w: projectName must match pattern", service.ErrValidation),
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := `{"projectName":"bad name!","localPath":"/tmp/musuhi/bad-name","template":"default"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/init-directory", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "VALIDATION_ERROR", errBody["code"])
	svc.AssertExpectations(t)
}

func TestIntegration_InitDirectory_不正なJSONで初期ディレクトリを作成する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	res, err := http.Post(srv.URL+"/api/v1/projects/init-directory", "application/json", bytes.NewBufferString(`invalid-json`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "BAD_REQUEST", errBody["code"])
}

func TestIntegration_WithExternal_GitHubリポジトリ作成と初回pushを実行する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	svc.On("CreateRepositoryWithExternal", mock.Anything, "BossApe", "demo-project", "private", "/tmp/demo-project", "initial commit").Return(
		&model.ProjectWithExternalResult{
			RepositoryURL:     "https://github.com/BossApe/demo-project",
			ExternalProjectID: "123456",
			PushStatus:        "success",
		}, nil,
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := `{"owner":"BossApe","repoName":"demo-project","visibility":"private","localPath":"/tmp/demo-project","commitMessage":"initial commit"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/with-external", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "https://github.com/BossApe/demo-project", data["repositoryUrl"])
	assert.Equal(t, "success", data["pushStatus"])
	svc.AssertExpectations(t)
}

func TestIntegration_WithExternal_ownerを空で指定してGitHubリポジトリ作成を実行する_異常系(t *testing.T) {
	svc := new(mockProjectService)
	svc.On("CreateRepositoryWithExternal", mock.Anything, "", "demo-project", "private", "/tmp/demo-project", "initial commit").Return(
		nil, fmt.Errorf("%w: owner is required", service.ErrValidation),
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := `{"owner":"","repoName":"demo-project","visibility":"private","localPath":"/tmp/demo-project","commitMessage":"initial commit"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/with-external", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "VALIDATION_ERROR", errBody["code"])
	svc.AssertExpectations(t)
}

func TestIntegration_GitHubProjects_有効な入力でProjectsボードを作成する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	svc.On("CreateGitHubProjects", mock.Anything, "project-1", "BossApe", "Musuhi Board").Return(
		&model.GitHubProjectsResult{
			ProjectsURL: "https://github.com/orgs/BossApe/projects/77",
			ProjectsID:  "PVT_test_001",
			Status:      "success",
		}, nil,
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := `{"owner":"BossApe","title":"Musuhi Board"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/project-1/github-projects", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "success", data["status"])
	assert.Equal(t, "PVT_test_001", data["projectsId"])
	svc.AssertExpectations(t)
}

func TestIntegration_Phase0Tasks_有効な入力でPhase0タスクを登録する_正常系(t *testing.T) {
	svc := new(mockProjectService)
	svc.On("CreatePhase0Tasks", mock.Anything, "project-1", "BossApe", "PVT_test_001").Return(
		&model.Phase0TasksResult{
			Tasks: []*model.Phase0Task{
				{ID: "PVTI_1", Title: "PH0: 提案・要求仕様・要件定義", Type: "Phase"},
				{ID: "PVTI_2", Title: "SP0-1: 提案・要求仕様作成", Type: "Sprint"},
			},
			Status: "success",
		}, nil,
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := `{"owner":"BossApe","projectsId":"PVT_test_001"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/project-1/phase0-tasks", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "success", data["status"])
	tasks := data["tasks"].([]any)
	assert.Len(t, tasks, 2)
	svc.AssertExpectations(t)
}
