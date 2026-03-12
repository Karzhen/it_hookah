package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Status      OrderStatus
	TotalAmount string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Items       []OrderItem
}

type OrderItem struct {
	ID              uuid.UUID
	OrderID         uuid.UUID
	ProductID       uuid.UUID
	ProductNameSnap string
	UnitPriceSnap   string
	Quantity        int
	Subtotal        string
	CreatedAt       time.Time
}
