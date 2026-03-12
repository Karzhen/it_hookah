package domain

import (
	"time"

	"github.com/google/uuid"
)

const CategoryCodeHookahTobacco = "hookah_tobacco"

type ProductCategory struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TobaccoFlavor struct {
	ID          uuid.UUID
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TobaccoStrength struct {
	ID          uuid.UUID
	Name        string
	Level       int16
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Product struct {
	ID            uuid.UUID
	CategoryID    uuid.UUID
	Name          string
	Description   *string
	Price         string
	StockQuantity int
	Unit          string
	IsActive      bool
	StrengthID    *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Category ProductCategory
	Strength *TobaccoStrength
	Flavors  []TobaccoFlavor
	Tags     []Tag
}

type ProductFilter struct {
	CategoryCode *string
	Search       *string
	MinPrice     *string
	MaxPrice     *string
	InStock      *bool
	FlavorID     *uuid.UUID
	StrengthID   *uuid.UUID
	IsActive     *bool
	Limit        int
	Offset       int
	AdminView    bool
}

type ProductUpsert struct {
	ID            uuid.UUID
	CategoryID    uuid.UUID
	Name          string
	Description   *string
	Price         string
	StockQuantity int
	Unit          string
	IsActive      bool
	StrengthID    *uuid.UUID
	FlavorIDs     []uuid.UUID
}
