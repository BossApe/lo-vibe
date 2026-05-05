package model

import (
	"time"

	"github.com/google/uuid"
)

// SystemOverview はシステム概要入力・保存機能（FR-001）のドメインモデル
type SystemOverview struct {
	ID        uuid.UUID `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
