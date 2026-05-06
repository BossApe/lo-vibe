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

func newProjectTestServer(svc service.ProjectService) *httptest.Server {
	ph := handler.NewProjectHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/projects/extract-features", ph.ExtractFeatures)
	mux.HandleFunc("POST /api/v1/projects/suggest-name", ph.SuggestName)
	mux.HandleFunc("POST /api/v1/projects/init-directory", ph.InitDirectory)
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
		&model.ProjectNameSuggestion{Candidates: []string{"book-app", "book-core", "book-manager"}},
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
	svc.On("InitDirectory", mock.Anything, "demo-project", "/tmp/musuhi", "default").Return(
		&model.ProjectInitResult{ID: id, DirectoryStatus: "success"}, nil,
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := `{"projectName":"demo-project","localPath":"/tmp/musuhi","template":"default"}`
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
	svc.On("InitDirectory", mock.Anything, "bad name!", "/tmp/musuhi", "default").Return(
		nil, fmt.Errorf("%w: projectName must match pattern", service.ErrValidation),
	)
	srv := newProjectTestServer(svc)
	defer srv.Close()

	body := `{"projectName":"bad name!","localPath":"/tmp/musuhi","template":"default"}`
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
