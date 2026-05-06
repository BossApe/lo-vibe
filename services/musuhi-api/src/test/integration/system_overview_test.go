package integration

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestIntegration_ヘルスチェックAPIにアクセスする_正常系(t *testing.T) {
	svc := new(mockSystemOverviewService)
	srv := newTestServer(svc)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestIntegration_CreateSystemOverview_通常のテキストでシステム概要を保存する_正常系(t *testing.T) {
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

func TestIntegration_CreateSystemOverview_空文字でシステム概要を保存する_異常系(t *testing.T) {
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

func TestIntegration_GetByID_存在しないUUIDでシステム概要を取得する_異常系(t *testing.T) {
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

func TestIntegration_CORS_システム概要作成APIのCORSヘッダーを確認する_連携(t *testing.T) {
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
