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

func TestSystemOverviewHandler_Create_OK(t *testing.T) {
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

func TestSystemOverviewHandler_Create_ValidationError(t *testing.T) {
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

func TestSystemOverviewHandler_GetByID_OK(t *testing.T) {
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

func TestSystemOverviewHandler_GetByID_NotFound(t *testing.T) {
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
