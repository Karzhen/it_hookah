package service

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/validation"
)

var tagCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type mixReader interface {
	GetMixByID(ctx context.Context, mixID uuid.UUID, adminView bool) (*domain.Mix, error)
}

type TagManager struct {
	store    TagStore
	products ProductReader
	mixes    mixReader
}

func NewTagManager(store TagStore, products ProductReader, mixes mixReader) *TagManager {
	return &TagManager{
		store:    store,
		products: products,
		mixes:    mixes,
	}
}

func (s *TagManager) ListPublicTags(ctx context.Context, limit, offset int) ([]domain.Tag, error) {
	limit = normalizeTagLimit(limit)
	offset = normalizeTagOffset(offset)

	tags, err := s.store.ListTags(ctx, true, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list tags", http.StatusInternalServerError, err)
	}
	return tags, nil
}

func (s *TagManager) ListAdminTags(ctx context.Context, limit, offset int) ([]domain.Tag, error) {
	limit = normalizeTagLimit(limit)
	offset = normalizeTagOffset(offset)

	tags, err := s.store.ListTags(ctx, false, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list tags", http.StatusInternalServerError, err)
	}
	return tags, nil
}

func (s *TagManager) GetPublicTagByID(ctx context.Context, tagID uuid.UUID) (*domain.Tag, error) {
	tag, err := s.store.GetTagByID(ctx, tagID, false)
	if err != nil {
		return nil, normalizeTagStoreError(err, "failed to fetch tag")
	}
	return tag, nil
}

func (s *TagManager) GetAdminTagByID(ctx context.Context, tagID uuid.UUID) (*domain.Tag, error) {
	tag, err := s.store.GetTagByID(ctx, tagID, true)
	if err != nil {
		return nil, normalizeTagStoreError(err, "failed to fetch tag")
	}
	return tag, nil
}

func (s *TagManager) CreateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	if tag == nil {
		return nil, apperror.New(apperror.CodeValidationError, "tag payload is required", http.StatusBadRequest)
	}

	code := strings.ToLower(strings.TrimSpace(tag.Code))
	if !tagCodePattern.MatchString(code) {
		return nil, apperror.New(apperror.CodeValidationError, "invalid tag code", http.StatusBadRequest)
	}

	name := strings.TrimSpace(tag.Name)
	if validation.IsBlank(name) {
		return nil, apperror.New(apperror.CodeValidationError, "name is required", http.StatusBadRequest)
	}

	if _, err := s.store.GetTagByCode(ctx, code, true); err == nil {
		return nil, apperror.New(apperror.CodeDuplicateTagCode, "duplicate tag code", http.StatusConflict)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to validate tag code", http.StatusInternalServerError, err)
	}

	if _, err := s.store.GetTagByName(ctx, name, true); err == nil {
		return nil, apperror.New(apperror.CodeDuplicateTagName, "duplicate tag name", http.StatusConflict)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to validate tag name", http.StatusInternalServerError, err)
	}

	now := time.Now().UTC()
	entity := &domain.Tag{
		ID:          uuid.New(),
		Code:        code,
		Name:        name,
		Description: trimTagOptional(tag.Description),
		IsActive:    tag.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.store.CreateTag(ctx, entity); err != nil {
		return nil, normalizeTagStoreError(err, "failed to create tag")
	}

	return entity, nil
}

func (s *TagManager) UpdateTag(ctx context.Context, tagID uuid.UUID, tag *domain.Tag) (*domain.Tag, error) {
	if tag == nil {
		return nil, apperror.New(apperror.CodeValidationError, "tag payload is required", http.StatusBadRequest)
	}

	current, err := s.store.GetTagByID(ctx, tagID, true)
	if err != nil {
		return nil, normalizeTagStoreError(err, "failed to fetch tag")
	}

	code := strings.ToLower(strings.TrimSpace(tag.Code))
	if !tagCodePattern.MatchString(code) {
		return nil, apperror.New(apperror.CodeValidationError, "invalid tag code", http.StatusBadRequest)
	}

	name := strings.TrimSpace(tag.Name)
	if validation.IsBlank(name) {
		return nil, apperror.New(apperror.CodeValidationError, "name is required", http.StatusBadRequest)
	}

	existingByCode, err := s.store.GetTagByCode(ctx, code, true)
	if err == nil && existingByCode.ID != current.ID {
		return nil, apperror.New(apperror.CodeDuplicateTagCode, "duplicate tag code", http.StatusConflict)
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to validate tag code", http.StatusInternalServerError, err)
	}

	existingByName, err := s.store.GetTagByName(ctx, name, true)
	if err == nil && existingByName.ID != current.ID {
		return nil, apperror.New(apperror.CodeDuplicateTagName, "duplicate tag name", http.StatusConflict)
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to validate tag name", http.StatusInternalServerError, err)
	}

	entity := &domain.Tag{
		ID:          current.ID,
		Code:        code,
		Name:        name,
		Description: trimTagOptional(tag.Description),
		IsActive:    tag.IsActive,
		CreatedAt:   current.CreatedAt,
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.store.UpdateTag(ctx, entity); err != nil {
		return nil, normalizeTagStoreError(err, "failed to update tag")
	}

	updated, err := s.store.GetTagByID(ctx, tagID, true)
	if err != nil {
		return nil, normalizeTagStoreError(err, "failed to fetch updated tag")
	}
	return updated, nil
}

func (s *TagManager) DeactivateTag(ctx context.Context, tagID uuid.UUID) error {
	if err := s.store.DeactivateTag(ctx, tagID); err != nil {
		return normalizeTagStoreError(err, "failed to deactivate tag")
	}
	return nil
}

func (s *TagManager) SetProductTags(ctx context.Context, productID uuid.UUID, tagIDs []uuid.UUID) error {
	if _, err := s.products.GetProductByID(ctx, productID, true); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to fetch product", http.StatusInternalServerError, err)
	}

	uniqueTagIDs := dedupeTagIDs(tagIDs)
	if len(uniqueTagIDs) > 0 {
		tags, err := s.store.GetTagsByIDs(ctx, uniqueTagIDs, true)
		if err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to fetch tags", http.StatusInternalServerError, err)
		}
		if len(tags) != len(uniqueTagIDs) {
			return apperror.New(apperror.CodeTagNotFound, "tag not found", http.StatusNotFound)
		}
	}

	if err := s.store.SetProductTags(ctx, productID, uniqueTagIDs); err != nil {
		return normalizeTagStoreError(err, "failed to set product tags")
	}
	return nil
}

func (s *TagManager) ListProductTags(ctx context.Context, productID uuid.UUID, adminView bool) ([]domain.Tag, error) {
	if _, err := s.products.GetProductByID(ctx, productID, true); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch product", http.StatusInternalServerError, err)
	}

	tags, err := s.store.ListProductTags(ctx, productID, !adminView)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list product tags", http.StatusInternalServerError, err)
	}
	return tags, nil
}

func (s *TagManager) SetMixTags(ctx context.Context, mixID uuid.UUID, tagIDs []uuid.UUID) error {
	if _, err := s.mixes.GetMixByID(ctx, mixID, true); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(apperror.CodeMixNotFound, "mix not found", http.StatusNotFound)
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to fetch mix", http.StatusInternalServerError, err)
	}

	uniqueTagIDs := dedupeTagIDs(tagIDs)
	if len(uniqueTagIDs) > 0 {
		tags, err := s.store.GetTagsByIDs(ctx, uniqueTagIDs, true)
		if err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to fetch tags", http.StatusInternalServerError, err)
		}
		if len(tags) != len(uniqueTagIDs) {
			return apperror.New(apperror.CodeTagNotFound, "tag not found", http.StatusNotFound)
		}
	}

	if err := s.store.SetMixTags(ctx, mixID, uniqueTagIDs); err != nil {
		return normalizeTagStoreError(err, "failed to set mix tags")
	}
	return nil
}

func (s *TagManager) ListMixTags(ctx context.Context, mixID uuid.UUID, adminView bool) ([]domain.Tag, error) {
	if _, err := s.mixes.GetMixByID(ctx, mixID, true); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeMixNotFound, "mix not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch mix", http.StatusInternalServerError, err)
	}

	tags, err := s.store.ListMixTags(ctx, mixID, !adminView)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list mix tags", http.StatusInternalServerError, err)
	}
	return tags, nil
}

func normalizeTagStoreError(err error, message string) error {
	if errors.Is(err, repository.ErrNotFound) {
		return apperror.New(apperror.CodeTagNotFound, "tag not found", http.StatusNotFound)
	}
	if errors.Is(err, repository.ErrDuplicateTagCode) {
		return apperror.New(apperror.CodeDuplicateTagCode, "duplicate tag code", http.StatusConflict)
	}
	if errors.Is(err, repository.ErrDuplicateTagName) {
		return apperror.New(apperror.CodeDuplicateTagName, "duplicate tag name", http.StatusConflict)
	}
	return apperror.Wrap(apperror.CodeInternal, message, http.StatusInternalServerError, err)
}

func trimTagOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeTagLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeTagOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func dedupeTagIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, item := range values {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
