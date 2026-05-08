package service

import (
	"context"
	"testing"
	"time"

	"musuhi-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (m *mockSystemOverviewRepository) UpdateByID(ctx context.Context, id uuid.UUID, content string) (*model.SystemOverview, error) {
	args := m.Called(ctx, id, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemOverview), args.Error(1)
}

func TestSystemOverviewService_Create_システム概要に通常のテキストを入力して保存する_正常系(t *testing.T) {
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

func TestSystemOverviewService_Create_システム概要を空文字で入力して保存する_異常系(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	_, err := svc.Create(context.Background(), "")
	assert.ErrorIs(t, err, ErrValidation)
	repo.AssertNotCalled(t, "Create")
}

func TestSystemOverviewService_Create_システム概要に4097文字を入力して保存する_境界値(t *testing.T) {
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

func TestSystemOverviewService_GetByID_有効なUUIDでシステム概要を取得する_正常系(t *testing.T) {
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

func TestSystemOverviewService_GetByID_UUID形式ではないIDでシステム概要を取得する_異常系(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	_, err := svc.GetByID(context.Background(), "not-a-uuid")
	assert.ErrorIs(t, err, ErrValidation)
	repo.AssertNotCalled(t, "FindByID")
}

func TestSystemOverviewService_GetByID_存在しないUUIDでシステム概要を取得する_異常系(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, pgx.ErrNoRows)

	_, err := svc.GetByID(context.Background(), id.String())
	assert.ErrorIs(t, err, ErrNotFound)
	repo.AssertExpectations(t)
}

func TestSystemOverviewService_Update_有効なUUIDと通常のテキストでシステム概要を更新する_正常系(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	id := uuid.New()
	content := "更新後のシステム概要"
	updated := &model.SystemOverview{ID: id, Content: content, CreatedAt: time.Now()}

	repo.On("UpdateByID", mock.Anything, id, content).Return(updated, nil)

	got, err := svc.Update(context.Background(), id.String(), content)
	assert.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, content, got.Content)
	repo.AssertExpectations(t)
}

func TestSystemOverviewService_Update_UUID形式ではないIDでシステム概要を更新する_異常系(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	_, err := svc.Update(context.Background(), "not-a-uuid", "てすと")
	assert.ErrorIs(t, err, ErrValidation)
	repo.AssertNotCalled(t, "UpdateByID")
}

func TestSystemOverviewService_Update_空文字でシステム概要を更新する_異常系(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	_, err := svc.Update(context.Background(), uuid.New().String(), "")
	assert.ErrorIs(t, err, ErrValidation)
	repo.AssertNotCalled(t, "UpdateByID")
}

func TestSystemOverviewService_Update_4097文字でシステム概要を更新する_境界値(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	longContent := ""
	for range 4097 {
		longContent += "あ"
	}

	_, err := svc.Update(context.Background(), uuid.New().String(), longContent)
	assert.ErrorIs(t, err, ErrValidation)
	repo.AssertNotCalled(t, "UpdateByID")
}

func TestSystemOverviewService_Update_存在しないUUIDでシステム概要を更新する_異常系(t *testing.T) {
	repo := new(mockSystemOverviewRepository)
	svc := NewSystemOverviewService(repo)

	id := uuid.New()
	repo.On("UpdateByID", mock.Anything, id, "てすと").Return(nil, pgx.ErrNoRows)

	_, err := svc.Update(context.Background(), id.String(), "てすと")
	assert.ErrorIs(t, err, ErrNotFound)
	repo.AssertExpectations(t)
}
