package domain

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductTag struct {
	ProductID uuid.UUID
	TagID     uuid.UUID
	CreatedAt time.Time
}

type MixTag struct {
	MixID     uuid.UUID
	TagID     uuid.UUID
	CreatedAt time.Time
}
