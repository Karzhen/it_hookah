package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/validation"
)

type UserService struct {
	users UserRepository
}

func NewUserService(users UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeUserNotFound, "user not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch user", http.StatusInternalServerError, err)
	}
	return user, nil
}

func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, apperror.New(apperror.CodeForbidden, "user is inactive", http.StatusForbidden)
	}

	return user, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID uuid.UUID, req dto.UpdateMeRequest) (*domain.User, error) {
	input := domain.UpdateUserProfile{}
	hasUpdates := false

	if req.FirstName != nil {
		trimmed := strings.TrimSpace(*req.FirstName)
		if validation.IsBlank(trimmed) {
			return nil, apperror.New(apperror.CodeValidationError, "first_name cannot be empty", http.StatusBadRequest)
		}
		input.FirstName = &trimmed
		hasUpdates = true
	}
	if req.LastName != nil {
		trimmed := strings.TrimSpace(*req.LastName)
		if validation.IsBlank(trimmed) {
			return nil, apperror.New(apperror.CodeValidationError, "last_name cannot be empty", http.StatusBadRequest)
		}
		input.LastName = &trimmed
		hasUpdates = true
	}
	if req.MiddleName != nil {
		trimmed := strings.TrimSpace(*req.MiddleName)
		input.MiddleName = &trimmed
		hasUpdates = true
	}
	if req.Phone != nil {
		trimmed := strings.TrimSpace(*req.Phone)
		input.Phone = &trimmed
		hasUpdates = true
	}
	if req.Age != nil {
		if *req.Age <= 0 {
			return nil, apperror.New(apperror.CodeValidationError, "age must be positive", http.StatusBadRequest)
		}
		age := *req.Age
		input.Age = &age
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, apperror.New(apperror.CodeValidationError, "no fields provided for update", http.StatusBadRequest)
	}

	updated, err := s.users.UpdateProfile(ctx, userID, input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeUserNotFound, "user not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to update profile", http.StatusInternalServerError, err)
	}

	return updated, nil
}
