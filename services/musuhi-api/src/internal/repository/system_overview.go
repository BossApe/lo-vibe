package repository

import (
	"context"

	"musuhi-api/internal/model"

	"github.com/google/uuid"
)

// SystemOverviewRepository はシステム概要の永続化インターフェース
type SystemOverviewRepository interface {
	Create(ctx context.Context, content string) (*model.SystemOverview, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.SystemOverview, error)
}
