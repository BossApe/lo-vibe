package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"musuhi-api/internal/handler"
	"musuhi-api/internal/middleware"
	"musuhi-api/internal/model"
	"musuhi-api/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockSystemOverviewService はインテグレーションテスト用サービスモック
type mockSystemOverviewService struct {
	mock.Mock
}

func (m *mockSystemOverviewService) Create(ctx context.Context, content string) (*model.SystemOverview, error) {
	args := m.Called(ctx, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func (m *mockSystemOverviewService) GetByID(ctx context.Context, rawID string) (*model.SystemOverview, error) {
	args := m.Called(ctx, rawID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func newTestServer(svc service.SystemOverviewService) *httptest.Server {
	soHandler := handler.NewSystemOverviewHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("POST /api/v1/system-overviews", soHandler.Create)
	mux.HandleFunc("GET /api/v1/system-overviews/{id}", soHandler.GetByID)
	h := middleware.Logger(middleware.CORS(mux))
	return httptest.NewServer(h)
}

func TestIntegration_HealthCheck(t *testing.T) {
	svc := new(mockSystemOverviewService)
	srv := newTestServer(svc)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestIntegration_CreateSystemOverview_OK(t *testing.T) {
	svc := new(mockSystemOverviewService)
	id := uuid.New()
	svc.On("Create", mock.Anything, "結合テスト用システム概要").Return(
		&model.SystemOverview{ID: id, Content: "結合テスト用システム概要", CreatedAt: time.Now()}, nil,
	)
	srv := newTestServer(svc)
	defer srv.Close()

	body := `{"content":"結合テスト用システム概要"}`
	res, err := http.Post(srv.URL+"/api/v1/system-overviews", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, id.String(), data["id"])
	assert.Equal(t, "結合テスト用システム概要", data["content"])
	svc.AssertExpectations(t)
}

func TestIntegration_CreateSystemOverview_ValidationError(t *testing.T) {
	svc := new(mockSystemOverviewService)
	svc.On("Create", mock.Anything, "").Return(nil, service.ErrValidation)
	srv := newTestServer(svc)
	defer srv.Close()

	body := `{"content":""}`
	res, err := http.Post(srv.URL+"/api/v1/system-overviews", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "VALIDATION_ERROR", errBody["code"])
}

func TestIntegration_GetByID_NotFound(t *testing.T) {
	svc := new(mockSystemOverviewService)
	id := uuid.New()
	svc.On("GetByID", mock.Anything, id.String()).Return(nil, service.ErrNotFound)
	srv := newTestServer(svc)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/v1/system-overviews/" + id.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	errBody := resp["error"].(map[string]any)
	assert.Equal(t, "NOT_FOUND", errBody["code"])
	svc.AssertExpectations(t)
}

func TestIntegration_CORS_Headers(t *testing.T) {
	svc := new(mockSystemOverviewService)
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodOptions, srv.URL+"/api/v1/system-overviews", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, res.Header.Get("Access-Control-Allow-Origin"))
}

// --- FR-002 結合テスト ---

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

func newProjectTestServer(soSvc service.SystemOverviewService, projSvc service.ProjectService) *httptest.Server {
	soHandler := handler.NewSystemOverviewHandler(soSvc)
	projHandler := handler.NewProjectHandler(projSvc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("POST /api/v1/system-overviews", soHandler.Create)
	mux.HandleFunc("GET /api/v1/system-overviews/{id}", soHandler.GetByID)
	mux.HandleFunc("POST /api/v1/projects/extract-features", projHandler.ExtractFeatures)
	mux.HandleFunc("POST /api/v1/projects/suggest-name", projHandler.SuggestName)
	mux.HandleFunc("POST /api/v1/projects/init-directory", projHandler.InitDirectory)
	h := middleware.Logger(middleware.CORS(mux))
	return httptest.NewServer(h)
}

func TestIntegration_ExtractFeatures_OK(t *testing.T) {
	soSvc := new(mockSystemOverviewService)
	projSvc := new(mockProjectService)
	overviewID := uuid.New()
	projSvc.On("ExtractFeatures", mock.Anything, overviewID.String()).Return(
		&model.ProjectExtraction{Features: []string{"ユーザ管理機能"}, Components: []string{"Backend API"}}, nil,
	)
	srv := newProjectTestServer(soSvc, projSvc)
	defer srv.Close()

	body := `{"overviewId":"` + overviewID.String() + `"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/extract-features", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	assert.NotEmpty(t, data["features"])
	projSvc.AssertExpectations(t)
}

func TestIntegration_SuggestName_OK(t *testing.T) {
	soSvc := new(mockSystemOverviewService)
	projSvc := new(mockProjectService)
	overviewID := uuid.New()
	projSvc.On("SuggestName", mock.Anything, overviewID.String()).Return(
		&model.ProjectNameSuggestion{Candidates: []string{"book-app", "book-core"}}, nil,
	)
	srv := newProjectTestServer(soSvc, projSvc)
	defer srv.Close()

	body := `{"overviewId":"` + overviewID.String() + `"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/suggest-name", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	assert.NotEmpty(t, data["candidates"])
	projSvc.AssertExpectations(t)
}

func TestIntegration_InitDirectory_OK(t *testing.T) {
	soSvc := new(mockSystemOverviewService)
	projSvc := new(mockProjectService)
	id := uuid.New()
	projSvc.On("InitDirectory", mock.Anything, "demo-project", "/tmp/test", "default").Return(
		&model.ProjectInitResult{ID: id, DirectoryStatus: "success"}, nil,
	)
	srv := newProjectTestServer(soSvc, projSvc)
	defer srv.Close()

	body := `{"projectName":"demo-project","localPath":"/tmp/test","template":"default"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/init-directory", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var resp map[string]any
	assert.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "success", data["directoryStatus"])
	projSvc.AssertExpectations(t)
}

func TestIntegration_InitDirectory_ValidationError(t *testing.T) {
	soSvc := new(mockSystemOverviewService)
	projSvc := new(mockProjectService)
	projSvc.On("InitDirectory", mock.Anything, "bad name!", "/tmp", "default").Return(
		nil, fmt.Errorf("%w: projectName must match pattern", service.ErrValidation),
	)
	srv := newProjectTestServer(soSvc, projSvc)
	defer srv.Close()

	body := `{"projectName":"bad name!","localPath":"/tmp","template":"default"}`
	res, err := http.Post(srv.URL+"/api/v1/projects/init-directory", "application/json", bytes.NewBufferString(body))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
	projSvc.AssertExpectations(t)
}
