package domain

import (
	"time"

	"github.com/google/uuid"
)

type StockMovementOperation string

const (
	StockMovementOperationSet               StockMovementOperation = "set"
	StockMovementOperationIncrement         StockMovementOperation = "increment"
	StockMovementOperationDecrement         StockMovementOperation = "decrement"
	StockMovementOperationCheckoutDecrement StockMovementOperation = "checkout_decrement"
)

type StockMovement struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	ProductName     *string
	Operation       StockMovementOperation
	Quantity        int
	BeforeQuantity  int
	AfterQuantity   int
	Reason          *string
	CreatedByUserID *uuid.UUID
	CreatedAt       time.Time
}

type StockMovementFilter struct {
	ProductID *uuid.UUID
	Operation *StockMovementOperation
	Limit     int
	Offset    int
}
