package domain

import (
	"time"

	"github.com/google/uuid"
)

type Mix struct {
	ID                 uuid.UUID
	Name               string
	Description        *string
	FinalStrengthLabel *string
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Items              []MixItem
	Tags               []Tag
}

type MixItem struct {
	ID        uuid.UUID
	MixID     uuid.UUID
	ProductID uuid.UUID
	Percent   int16
	CreatedAt time.Time
}
