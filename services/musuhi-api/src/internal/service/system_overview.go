package service

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"musuhi-api/internal/model"
	"musuhi-api/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const maxContentLength = 4096

// ErrNotFound はリソース未存在エラー
var ErrNotFound = errors.New("not found")

// ErrValidation はバリデーションエラー
var ErrValidation = errors.New("validation error")

// SystemOverviewService はシステム概要のビジネスロジックインターフェース
type SystemOverviewService interface {
	Create(ctx context.Context, content string) (*model.SystemOverview, error)
	GetByID(ctx context.Context, rawID string) (*model.SystemOverview, error)
	Update(ctx context.Context, rawID string, content string) (*model.SystemOverview, error)
}

type systemOverviewService struct {
	repo repository.SystemOverviewRepository
}

// NewSystemOverviewService は SystemOverviewService を生成する
func NewSystemOverviewService(repo repository.SystemOverviewRepository) SystemOverviewService {
	return &systemOverviewService{repo: repo}
}

func (s *systemOverviewService) Create(ctx context.Context, content string) (*model.SystemOverview, error) {
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrValidation)
	}
	if utf8.RuneCountInString(content) > maxContentLength {
		return nil, fmt.Errorf("%w: content must be %d characters or less", ErrValidation, maxContentLength)
	}

	m, err := s.repo.Create(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("systemOverviewService.Create: %w", err)
	}
	return m, nil
}

func (s *systemOverviewService) GetByID(ctx context.Context, rawID string) (*model.SystemOverview, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid id format", ErrValidation)
	}

	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: system_overview id=%s", ErrNotFound, rawID)
		}
		return nil, fmt.Errorf("systemOverviewService.GetByID: %w", err)
	}
	return m, nil
}

func (s *systemOverviewService) Update(ctx context.Context, rawID string, content string) (*model.SystemOverview, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid id format", ErrValidation)
	}
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrValidation)
	}
	if utf8.RuneCountInString(content) > maxContentLength {
		return nil, fmt.Errorf("%w: content must be %d characters or less", ErrValidation, maxContentLength)
	}

	m, err := s.repo.UpdateByID(ctx, id, content)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: system_overview id=%s", ErrNotFound, rawID)
		}
		return nil, fmt.Errorf("systemOverviewService.Update: %w", err)
	}
	return m, nil
}
