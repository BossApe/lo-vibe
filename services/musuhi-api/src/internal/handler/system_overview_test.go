package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"musuhi-api/internal/model"
	"musuhi-api/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockSystemOverviewService はテスト用サービスモック
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

func (m *mockSystemOverviewService) Update(ctx context.Context, rawID string, content string) (*model.SystemOverview, error) {
	args := m.Called(ctx, rawID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func TestSystemOverviewHandler_Create_システム概要に通常のテキストを入力して保存する_正常系(t *testing.T) {
	svc := new(mockSystemOverviewService)
	h := NewSystemOverviewHandler(svc)

	id := uuid.New()
	svc.On("Create", mock.Anything, "テスト概要テキスト").Return(
		&model.SystemOverview{ID: id, Content: "テスト概要テキスト", CreatedAt: time.Now()},
		nil,
	)

	body := `{"content":"テスト概要テキスト"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system-overviews", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, id.String(), data["id"])
	svc.AssertExpectations(t)
}

func TestSystemOverviewHandler_Create_システム概要を空文字で入力して保存する_異常系(t *testing.T) {
	svc := new(mockSystemOverviewService)
	h := NewSystemOverviewHandler(svc)

	svc.On("Create", mock.Anything, "").Return(nil, fmt.Errorf("%w: content is required", service.ErrValidation))

	body := `{"content":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system-overviews", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSystemOverviewHandler_GetByID_有効なUUIDでシステム概要を取得する_正常系(t *testing.T) {
	svc := new(mockSystemOverviewService)
	h := NewSystemOverviewHandler(svc)

	id := uuid.New()
	svc.On("GetByID", mock.Anything, id.String()).Return(
		&model.SystemOverview{ID: id, Content: "概要テキスト", CreatedAt: time.Now()},
		nil,
	)

	// Go 1.22 ServeMux の PathValue を再現するため直接呼び出し
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system-overviews/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, id.String(), data["id"])
	svc.AssertExpectations(t)
}

func TestSystemOverviewHandler_GetByID_存在しないUUIDでシステム概要を取得する_異常系(t *testing.T) {
	svc := new(mockSystemOverviewService)
	h := NewSystemOverviewHandler(svc)

	id := uuid.New()
	svc.On("GetByID", mock.Anything, id.String()).Return(nil, fmt.Errorf("%w", service.ErrNotFound))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system-overviews/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSystemOverviewHandler_Update_有効なUUIDと通常のテキストでシステム概要を更新する_正常系(t *testing.T) {
	svc := new(mockSystemOverviewService)
	h := NewSystemOverviewHandler(svc)

	id := uuid.New()
	content := "更新後のシステム概要"
	updated := &model.SystemOverview{ID: id, Content: content, CreatedAt: time.Now()}

	svc.On("Update", mock.Anything, id.String(), content).Return(updated, nil)

	body, _ := json.Marshal(map[string]string{"content": content})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/system-overviews/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, id.String(), data["id"])
	svc.AssertExpectations(t)
}

func TestSystemOverviewHandler_Update_存在しないUUIDでシステム概要を更新する_異常系(t *testing.T) {
	svc := new(mockSystemOverviewService)
	h := NewSystemOverviewHandler(svc)

	id := uuid.New()
	content := "更新テキスト"
	svc.On("Update", mock.Anything, id.String(), content).Return(nil, fmt.Errorf("%w", service.ErrNotFound))

	body, _ := json.Marshal(map[string]string{"content": content})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/system-overviews/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestSystemOverviewHandler_Update_UUID形式ではないIDでシステム概要を更新する_異常系(t *testing.T) {
	svc := new(mockSystemOverviewService)
	h := NewSystemOverviewHandler(svc)

	content := "更新テキスト"
	svc.On("Update", mock.Anything, "not-a-uuid", content).Return(nil, fmt.Errorf("%w: invalid id format", service.ErrValidation))

	body, _ := json.Marshal(map[string]string{"content": content})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/system-overviews/not-a-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	svc.AssertExpectations(t)
}
