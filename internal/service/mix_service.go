package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/validation"
)

type MixManager struct {
	store    MixStore
	products ProductReader
}

func NewMixManager(store MixStore, products ProductReader) *MixManager {
	return &MixManager{
		store:    store,
		products: products,
	}
}

func (s *MixManager) GetPublicMixByID(ctx context.Context, mixID uuid.UUID) (*domain.Mix, error) {
	mix, err := s.store.GetMixByID(ctx, mixID, false)
	if err != nil {
		return nil, normalizeMixStoreError(err, "failed to fetch mix")
	}

	items, err := s.store.ListMixItemsByMixID(ctx, mix.ID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch mix items", http.StatusInternalServerError, err)
	}
	mix.Items = items

	return mix, nil
}

func (s *MixManager) ListPublicMixes(ctx context.Context, limit, offset int) ([]domain.Mix, error) {
	limit = normalizeMixLimit(limit)
	offset = normalizeMixOffset(offset)

	mixes, err := s.store.ListMixes(ctx, true, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list mixes", http.StatusInternalServerError, err)
	}

	if err := s.attachMixItems(ctx, mixes); err != nil {
		return nil, err
	}

	return mixes, nil
}

func (s *MixManager) GetAdminMixByID(ctx context.Context, mixID uuid.UUID) (*domain.Mix, error) {
	mix, err := s.store.GetMixByID(ctx, mixID, true)
	if err != nil {
		return nil, normalizeMixStoreError(err, "failed to fetch mix")
	}

	items, err := s.store.ListMixItemsByMixID(ctx, mix.ID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch mix items", http.StatusInternalServerError, err)
	}
	mix.Items = items

	return mix, nil
}

func (s *MixManager) ListAdminMixes(ctx context.Context, limit, offset int) ([]domain.Mix, error) {
	limit = normalizeMixLimit(limit)
	offset = normalizeMixOffset(offset)

	mixes, err := s.store.ListMixes(ctx, false, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list mixes", http.StatusInternalServerError, err)
	}

	if err := s.attachMixItems(ctx, mixes); err != nil {
		return nil, err
	}

	return mixes, nil
}

func (s *MixManager) CreateMix(ctx context.Context, mix *domain.Mix, items []domain.MixItem) (*domain.Mix, error) {
	if mix == nil {
		return nil, apperror.New(apperror.CodeValidationError, "mix payload is required", http.StatusBadRequest)
	}

	name := strings.TrimSpace(mix.Name)
	if validation.IsBlank(name) {
		return nil, apperror.New(apperror.CodeValidationError, "name is required", http.StatusBadRequest)
	}

	isActive := mix.IsActive
	description := trimOptionalText(mix.Description)
	finalStrength := trimOptionalText(mix.FinalStrengthLabel)

	validatedItems, err := s.validateAndPrepareMixItems(ctx, uuid.Nil, items, isActive)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	entity := &domain.Mix{
		ID:                 uuid.New(),
		Name:               name,
		Description:        description,
		FinalStrengthLabel: finalStrength,
		IsActive:           isActive,
		CreatedAt:          now,
		UpdatedAt:          now,
		Items:              validatedItems,
	}
	for i := range entity.Items {
		entity.Items[i].MixID = entity.ID
	}

	if err := s.store.CreateMix(ctx, entity); err != nil {
		return nil, normalizeMixStoreError(err, "failed to create mix")
	}

	return entity, nil
}

func (s *MixManager) UpdateMix(ctx context.Context, mixID uuid.UUID, mix *domain.Mix, items []domain.MixItem) (*domain.Mix, error) {
	if mix == nil {
		return nil, apperror.New(apperror.CodeValidationError, "mix payload is required", http.StatusBadRequest)
	}

	current, err := s.store.GetMixByID(ctx, mixID, true)
	if err != nil {
		return nil, normalizeMixStoreError(err, "failed to fetch mix")
	}

	name := strings.TrimSpace(mix.Name)
	if validation.IsBlank(name) {
		return nil, apperror.New(apperror.CodeValidationError, "name is required", http.StatusBadRequest)
	}

	isActive := mix.IsActive
	description := trimOptionalText(mix.Description)
	finalStrength := trimOptionalText(mix.FinalStrengthLabel)

	validatedItems, err := s.validateAndPrepareMixItems(ctx, mixID, items, isActive)
	if err != nil {
		return nil, err
	}

	updated := &domain.Mix{
		ID:                 current.ID,
		Name:               name,
		Description:        description,
		FinalStrengthLabel: finalStrength,
		IsActive:           isActive,
		CreatedAt:          current.CreatedAt,
		UpdatedAt:          time.Now().UTC(),
		Items:              validatedItems,
	}

	if err := s.store.UpdateMix(ctx, updated); err != nil {
		return nil, normalizeMixStoreError(err, "failed to update mix")
	}

	fresh, err := s.store.GetMixByID(ctx, mixID, true)
	if err != nil {
		return nil, normalizeMixStoreError(err, "failed to fetch updated mix")
	}
	freshItems, err := s.store.ListMixItemsByMixID(ctx, mixID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch mix items", http.StatusInternalServerError, err)
	}
	fresh.Items = freshItems

	return fresh, nil
}

func (s *MixManager) DeactivateMix(ctx context.Context, mixID uuid.UUID) error {
	if err := s.store.DeactivateMix(ctx, mixID); err != nil {
		return normalizeMixStoreError(err, "failed to deactivate mix")
	}
	return nil
}

func (s *MixManager) validateAndPrepareMixItems(
	ctx context.Context,
	mixID uuid.UUID,
	items []domain.MixItem,
	mixIsActive bool,
) ([]domain.MixItem, error) {
	if len(items) == 0 {
		return nil, apperror.New(apperror.CodeInvalidMixPercentTotal, "mix must contain at least one item", http.StatusBadRequest)
	}

	seenProducts := make(map[uuid.UUID]struct{}, len(items))
	result := make([]domain.MixItem, 0, len(items))
	percentTotal := 0
	now := time.Now().UTC()

	for _, item := range items {
		if item.ProductID == uuid.Nil {
			return nil, apperror.New(apperror.CodeInvalidMixItemProduct, "invalid mix item product", http.StatusBadRequest)
		}
		if item.Percent <= 0 || item.Percent > 100 {
			return nil, apperror.New(apperror.CodeInvalidMixPercentTotal, "invalid mix percent total", http.StatusBadRequest)
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return nil, apperror.New(apperror.CodeInvalidMixItemProduct, "duplicate tobacco product in mix", http.StatusBadRequest)
		}
		seenProducts[item.ProductID] = struct{}{}

		product, err := s.products.GetProductByID(ctx, item.ProductID, true)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, apperror.New(apperror.CodeTobaccoProductNotFound, "tobacco product not found", http.StatusNotFound)
			}
			return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch tobacco product", http.StatusInternalServerError, err)
		}
		if product.Category.Code != domain.CategoryCodeHookahTobacco {
			return nil, apperror.New(apperror.CodeInvalidMixItemProduct, "mix item product must be hookah tobacco", http.StatusBadRequest)
		}
		if mixIsActive && !product.IsActive {
			return nil, apperror.New(apperror.CodeInvalidMixItemProduct, "active mix requires active tobacco products", http.StatusBadRequest)
		}

		result = append(result, domain.MixItem{
			ID:        uuid.New(),
			MixID:     mixID,
			ProductID: item.ProductID,
			Percent:   item.Percent,
			CreatedAt: now,
		})
		percentTotal += int(item.Percent)
	}

	if percentTotal != 100 {
		return nil, apperror.New(apperror.CodeInvalidMixPercentTotal, "invalid mix percent total", http.StatusBadRequest)
	}

	return result, nil
}

func (s *MixManager) attachMixItems(ctx context.Context, mixes []domain.Mix) error {
	for i := range mixes {
		items, err := s.store.ListMixItemsByMixID(ctx, mixes[i].ID)
		if err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to fetch mix items", http.StatusInternalServerError, err)
		}
		mixes[i].Items = items
	}
	return nil
}

func normalizeMixStoreError(err error, message string) error {
	if errors.Is(err, repository.ErrNotFound) {
		return apperror.New(apperror.CodeMixNotFound, "mix not found", http.StatusNotFound)
	}
	if errors.Is(err, repository.ErrInvalidMixPercentTotal) {
		return apperror.New(apperror.CodeInvalidMixPercentTotal, "invalid mix percent total", http.StatusBadRequest)
	}
	if errors.Is(err, repository.ErrInvalidMixItemProduct) {
		return apperror.New(apperror.CodeInvalidMixItemProduct, "invalid mix item product", http.StatusBadRequest)
	}
	if errors.Is(err, repository.ErrTobaccoProductNotFound) {
		return apperror.New(apperror.CodeTobaccoProductNotFound, "tobacco product not found", http.StatusNotFound)
	}
	return apperror.Wrap(apperror.CodeInternal, message, http.StatusInternalServerError, err)
}

func trimOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeMixLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeMixOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
