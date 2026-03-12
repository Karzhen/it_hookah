package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	MiddleName   *string
	Phone        *string
	Age          *int
	Role         string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UpdateUserProfile struct {
	FirstName  *string
	LastName   *string
	MiddleName *string
	Phone      *string
	Age        *int
}
