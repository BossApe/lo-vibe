package repository

import (
	"context"
	"fmt"

	"musuhi-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresSystemOverviewRepository struct {
	db *pgxpool.Pool
}

// NewPostgresSystemOverviewRepository は PostgreSQL を使う SystemOverviewRepository を返す
func NewPostgresSystemOverviewRepository(db *pgxpool.Pool) SystemOverviewRepository {
	return &postgresSystemOverviewRepository{db: db}
}

func (r *postgresSystemOverviewRepository) Create(ctx context.Context, content string) (*model.SystemOverview, error) {
	const q = `
		INSERT INTO system_overviews (content)
		VALUES ($1)
		RETURNING id, content, created_at
	`
	var m model.SystemOverview
	err := r.db.QueryRow(ctx, q, content).Scan(&m.ID, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("systemOverviewRepository.Create: %w", err)
	}
	return &m, nil
}

func (r *postgresSystemOverviewRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.SystemOverview, error) {
	const q = `
		SELECT id, content, created_at
		FROM system_overviews
		WHERE id = $1
	`
	var m model.SystemOverview
	err := r.db.QueryRow(ctx, q, id).Scan(&m.ID, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("systemOverviewRepository.FindByID: %w", err)
	}
	return &m, nil
}
