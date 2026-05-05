package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"musuhi-api/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockSystemOverviewRepository はテスト用モック
type mockSystemOverviewRepository struct {
	mock.Mock
}

func (m *mockSystemOverviewRepository) Create(ctx context.Context, content string) (*model.SystemOverview, error) {
	args := m.Called(ctx, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func (m *mockSystemOverviewRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.SystemOverview, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func TestSystemOverviewService_Create_OK(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	id := uuid.New()
	content := "テスト用システム概要テキスト"
	expected := &model.SystemOverview{ID: id, Content: content, CreatedAt: time.Now()}

	repo.On("Create", mock.Anything, content).Return(expected, nil)

	got, err := svc.Create(context.Background(), content)
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, got.ID)
	assert.Equal(t, expected.Content, got.Content)
	repo.AssertExpectations(t)
}

func TestSystemOverviewService_Create_EmptyContent(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	_, err := svc.Create(context.Background(), "")
	assert.ErrorIs(t, err, ErrValidation)
	repo.AssertNotCalled(t, "Create")
}

func TestSystemOverviewService_Create_TooLongContent(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	// 4097文字のテキスト（上限4096を超える）
	longContent := ""
	for range 4097 {
		longContent += "あ"
	}

	_, err := svc.Create(context.Background(), longContent)
	assert.ErrorIs(t, err, ErrValidation)
	repo.AssertNotCalled(t, "Create")
}

func TestSystemOverviewService_GetByID_OK(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	id := uuid.New()
	expected := &model.SystemOverview{ID: id, Content: "概要テキスト", CreatedAt: time.Now()}

	repo.On("FindByID", mock.Anything, id).Return(expected, nil)

	got, err := svc.GetByID(context.Background(), id.String())
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, got.ID)
	repo.AssertExpectations(t)
}

func TestSystemOverviewService_GetByID_InvalidUUID(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	_, err := svc.GetByID(context.Background(), "not-a-uuid")
	assert.ErrorIs(t, err, ErrValidation)
	repo.AssertNotCalled(t, "FindByID")
}

func TestSystemOverviewService_GetByID_NotFound(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	id := uuid.New()
	// pgx.ErrNoRows を模倣
	repo.On("FindByID", mock.Anything, id).Return(nil, errors.New("no rows in result set"))

	_, err := svc.GetByID(context.Background(), id.String())
	assert.Error(t, err)
	repo.AssertExpectations(t)
}
