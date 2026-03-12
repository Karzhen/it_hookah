package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/ctxkeys"
	"github.com/karzhen/restaurant-lk/internal/utils/validation"
)

var categoryCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type CatalogService struct {
	repo CatalogRepository
}

func NewCatalogService(repo CatalogRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) ListPublicCategories(ctx context.Context) ([]domain.ProductCategory, error) {
	items, err := s.repo.ListCategories(ctx, true)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list categories", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *CatalogService) ListAdminCategories(ctx context.Context) ([]domain.ProductCategory, error) {
	items, err := s.repo.ListCategories(ctx, false)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list categories", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *CatalogService) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (*domain.ProductCategory, error) {
	code := normalizeCode(req.Code)
	name := strings.TrimSpace(req.Name)
	if !categoryCodePattern.MatchString(code) {
		return nil, apperror.New(apperror.CodeValidationError, "invalid category code", http.StatusBadRequest)
	}
	if validation.IsBlank(name) {
		return nil, apperror.New(apperror.CodeValidationError, "name is required", http.StatusBadRequest)
	}

	description := trimOptionalString(req.Description)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := time.Now().UTC()
	category := &domain.ProductCategory{
		ID:          uuid.New(),
		Code:        code,
		Name:        name,
		Description: description,
		IsActive:    isActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.New(apperror.CodeValidationError, "category code already exists", http.StatusConflict)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create category", http.StatusInternalServerError, err)
	}

	return category, nil
}

func (s *CatalogService) UpdateCategory(ctx context.Context, categoryID uuid.UUID, req dto.UpdateCategoryRequest) (*domain.ProductCategory, error) {
	category, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeCategoryNotFound, "category not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch category", http.StatusInternalServerError, err)
	}

	if req.Code != nil {
		code := normalizeCode(*req.Code)
		if !categoryCodePattern.MatchString(code) {
			return nil, apperror.New(apperror.CodeValidationError, "invalid category code", http.StatusBadRequest)
		}
		category.Code = code
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if validation.IsBlank(name) {
			return nil, apperror.New(apperror.CodeValidationError, "name cannot be empty", http.StatusBadRequest)
		}
		category.Name = name
	}
	if req.Description != nil {
		category.Description = trimOptionalString(req.Description)
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateCategory(ctx, category); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeCategoryNotFound, "category not found", http.StatusNotFound)
		}
		if isUniqueViolation(err) {
			return nil, apperror.New(apperror.CodeValidationError, "category code already exists", http.StatusConflict)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to update category", http.StatusInternalServerError, err)
	}

	updated, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch updated category", http.StatusInternalServerError, err)
	}

	return updated, nil
}

func (s *CatalogService) DeactivateCategory(ctx context.Context, categoryID uuid.UUID) error {
	if err := s.repo.DeactivateCategory(ctx, categoryID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(apperror.CodeCategoryNotFound, "category not found", http.StatusNotFound)
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to deactivate category", http.StatusInternalServerError, err)
	}
	return nil
}

func (s *CatalogService) ListPublicFlavors(ctx context.Context) ([]domain.TobaccoFlavor, error) {
	items, err := s.repo.ListFlavors(ctx, true)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list flavors", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *CatalogService) ListAdminFlavors(ctx context.Context) ([]domain.TobaccoFlavor, error) {
	items, err := s.repo.ListFlavors(ctx, false)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list flavors", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *CatalogService) CreateFlavor(ctx context.Context, req dto.CreateFlavorRequest) (*domain.TobaccoFlavor, error) {
	name := strings.TrimSpace(req.Name)
	if validation.IsBlank(name) {
		return nil, apperror.New(apperror.CodeValidationError, "name is required", http.StatusBadRequest)
	}

	description := trimOptionalString(req.Description)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := time.Now().UTC()
	flavor := &domain.TobaccoFlavor{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		IsActive:    isActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateFlavor(ctx, flavor); err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.New(apperror.CodeValidationError, "flavor already exists", http.StatusConflict)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create flavor", http.StatusInternalServerError, err)
	}

	return flavor, nil
}

func (s *CatalogService) UpdateFlavor(ctx context.Context, flavorID uuid.UUID, req dto.UpdateFlavorRequest) (*domain.TobaccoFlavor, error) {
	flavor, err := s.repo.GetFlavorByID(ctx, flavorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeFlavorNotFound, "flavor not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch flavor", http.StatusInternalServerError, err)
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if validation.IsBlank(name) {
			return nil, apperror.New(apperror.CodeValidationError, "name cannot be empty", http.StatusBadRequest)
		}
		flavor.Name = name
	}
	if req.Description != nil {
		flavor.Description = trimOptionalString(req.Description)
	}
	if req.IsActive != nil {
		flavor.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateFlavor(ctx, flavor); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeFlavorNotFound, "flavor not found", http.StatusNotFound)
		}
		if isUniqueViolation(err) {
			return nil, apperror.New(apperror.CodeValidationError, "flavor already exists", http.StatusConflict)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to update flavor", http.StatusInternalServerError, err)
	}

	updated, err := s.repo.GetFlavorByID(ctx, flavorID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch updated flavor", http.StatusInternalServerError, err)
	}
	return updated, nil
}

func (s *CatalogService) DeactivateFlavor(ctx context.Context, flavorID uuid.UUID) error {
	if err := s.repo.DeactivateFlavor(ctx, flavorID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(apperror.CodeFlavorNotFound, "flavor not found", http.StatusNotFound)
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to deactivate flavor", http.StatusInternalServerError, err)
	}
	return nil
}

func (s *CatalogService) ListPublicStrengths(ctx context.Context) ([]domain.TobaccoStrength, error) {
	items, err := s.repo.ListStrengths(ctx, true)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list strengths", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *CatalogService) ListAdminStrengths(ctx context.Context) ([]domain.TobaccoStrength, error) {
	items, err := s.repo.ListStrengths(ctx, false)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list strengths", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *CatalogService) CreateStrength(ctx context.Context, req dto.CreateStrengthRequest) (*domain.TobaccoStrength, error) {
	name := strings.TrimSpace(req.Name)
	if validation.IsBlank(name) {
		return nil, apperror.New(apperror.CodeValidationError, "name is required", http.StatusBadRequest)
	}
	if req.Level <= 0 {
		return nil, apperror.New(apperror.CodeValidationError, "level must be positive", http.StatusBadRequest)
	}

	description := trimOptionalString(req.Description)
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	now := time.Now().UTC()
	strength := &domain.TobaccoStrength{
		ID:          uuid.New(),
		Name:        name,
		Level:       req.Level,
		Description: description,
		IsActive:    isActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateStrength(ctx, strength); err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.New(apperror.CodeValidationError, "strength already exists", http.StatusConflict)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create strength", http.StatusInternalServerError, err)
	}

	return strength, nil
}

func (s *CatalogService) UpdateStrength(ctx context.Context, strengthID uuid.UUID, req dto.UpdateStrengthRequest) (*domain.TobaccoStrength, error) {
	strength, err := s.repo.GetStrengthByID(ctx, strengthID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeStrengthNotFound, "strength not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch strength", http.StatusInternalServerError, err)
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if validation.IsBlank(name) {
			return nil, apperror.New(apperror.CodeValidationError, "name cannot be empty", http.StatusBadRequest)
		}
		strength.Name = name
	}
	if req.Level != nil {
		if *req.Level <= 0 {
			return nil, apperror.New(apperror.CodeValidationError, "level must be positive", http.StatusBadRequest)
		}
		strength.Level = *req.Level
	}
	if req.Description != nil {
		strength.Description = trimOptionalString(req.Description)
	}
	if req.IsActive != nil {
		strength.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateStrength(ctx, strength); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeStrengthNotFound, "strength not found", http.StatusNotFound)
		}
		if isUniqueViolation(err) {
			return nil, apperror.New(apperror.CodeValidationError, "strength already exists", http.StatusConflict)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to update strength", http.StatusInternalServerError, err)
	}

	updated, err := s.repo.GetStrengthByID(ctx, strengthID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch updated strength", http.StatusInternalServerError, err)
	}
	return updated, nil
}

func (s *CatalogService) DeactivateStrength(ctx context.Context, strengthID uuid.UUID) error {
	if err := s.repo.DeactivateStrength(ctx, strengthID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(apperror.CodeStrengthNotFound, "strength not found", http.StatusNotFound)
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to deactivate strength", http.StatusInternalServerError, err)
	}
	return nil
}

func (s *CatalogService) ListPublicProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	filter.AdminView = false
	items, err := s.repo.ListProducts(ctx, filter)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list products", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *CatalogService) ListAdminProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	filter.AdminView = true
	items, err := s.repo.ListProducts(ctx, filter)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list products", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *CatalogService) GetPublicProductByID(ctx context.Context, productID uuid.UUID) (*domain.Product, error) {
	item, err := s.repo.GetProductByID(ctx, productID, false)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch product", http.StatusInternalServerError, err)
	}
	return item, nil
}

func (s *CatalogService) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*domain.Product, error) {
	categoryID, err := parseUUID(req.CategoryID)
	if err != nil {
		return nil, apperror.New(apperror.CodeValidationError, "invalid category_id", http.StatusBadRequest)
	}

	category, err := s.repo.GetCategoryByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeCategoryNotFound, "category not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch category", http.StatusInternalServerError, err)
	}

	name := strings.TrimSpace(req.Name)
	if validation.IsBlank(name) {
		return nil, apperror.New(apperror.CodeValidationError, "name is required", http.StatusBadRequest)
	}

	price := strings.TrimSpace(req.Price.String())
	if err := validatePrice(price); err != nil {
		return nil, err
	}

	if req.StockQuantity < 0 {
		return nil, apperror.New(apperror.CodeValidationError, "stock_quantity must be >= 0", http.StatusBadRequest)
	}

	unit := strings.TrimSpace(req.Unit)
	if validation.IsBlank(unit) {
		return nil, apperror.New(apperror.CodeValidationError, "unit is required", http.StatusBadRequest)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	strengthID, err := parseOptionalUUID(req.StrengthID)
	if err != nil {
		return nil, apperror.New(apperror.CodeValidationError, "invalid strength_id", http.StatusBadRequest)
	}

	flavorIDs, err := parseUUIDList(req.FlavorIDs)
	if err != nil {
		return nil, apperror.New(apperror.CodeValidationError, "invalid flavor_ids", http.StatusBadRequest)
	}
	flavorIDs = deduplicateUUIDs(flavorIDs)

	if err := s.validateTobaccoAttributes(ctx, category, isActive, strengthID, flavorIDs); err != nil {
		return nil, err
	}

	productID := uuid.New()
	upsert := domain.ProductUpsert{
		ID:            productID,
		CategoryID:    categoryID,
		Name:          name,
		Description:   trimOptionalString(req.Description),
		Price:         price,
		StockQuantity: req.StockQuantity,
		Unit:          unit,
		IsActive:      isActive,
		StrengthID:    strengthID,
		FlavorIDs:     flavorIDs,
	}

	if err := s.repo.CreateProduct(ctx, upsert); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create product", http.StatusInternalServerError, err)
	}

	created, err := s.repo.GetProductByID(ctx, productID, true)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch created product", http.StatusInternalServerError, err)
	}
	return created, nil
}

func (s *CatalogService) UpdateProduct(ctx context.Context, productID uuid.UUID, req dto.UpdateProductRequest) (*domain.Product, error) {
	current, err := s.repo.GetProductByID(ctx, productID, true)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch product", http.StatusInternalServerError, err)
	}

	categoryID := current.CategoryID
	category := &current.Category
	if req.CategoryID != nil {
		parsedCategoryID, err := parseUUID(*req.CategoryID)
		if err != nil {
			return nil, apperror.New(apperror.CodeValidationError, "invalid category_id", http.StatusBadRequest)
		}
		cat, err := s.repo.GetCategoryByID(ctx, parsedCategoryID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, apperror.New(apperror.CodeCategoryNotFound, "category not found", http.StatusNotFound)
			}
			return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch category", http.StatusInternalServerError, err)
		}
		categoryID = parsedCategoryID
		category = cat
	}

	name := current.Name
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
		if validation.IsBlank(name) {
			return nil, apperror.New(apperror.CodeValidationError, "name cannot be empty", http.StatusBadRequest)
		}
	}

	description := current.Description
	if req.Description != nil {
		description = trimOptionalString(req.Description)
	}

	price := current.Price
	if req.Price != nil {
		price = strings.TrimSpace(req.Price.String())
		if err := validatePrice(price); err != nil {
			return nil, err
		}
	}

	stockQuantity := current.StockQuantity
	if req.StockQuantity != nil {
		if *req.StockQuantity < 0 {
			return nil, apperror.New(apperror.CodeValidationError, "stock_quantity must be >= 0", http.StatusBadRequest)
		}
		stockQuantity = *req.StockQuantity
	}

	unit := current.Unit
	if req.Unit != nil {
		unit = strings.TrimSpace(*req.Unit)
		if validation.IsBlank(unit) {
			return nil, apperror.New(apperror.CodeValidationError, "unit cannot be empty", http.StatusBadRequest)
		}
	}

	isActive := current.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	strengthID := current.StrengthID
	strengthProvided := false
	if req.StrengthID != nil {
		parsedStrengthID, err := parseOptionalUUID(req.StrengthID)
		if err != nil {
			return nil, apperror.New(apperror.CodeValidationError, "invalid strength_id", http.StatusBadRequest)
		}
		strengthID = parsedStrengthID
		strengthProvided = true
	}

	flavorIDs := make([]uuid.UUID, 0, len(current.Flavors))
	for _, flavor := range current.Flavors {
		flavorIDs = append(flavorIDs, flavor.ID)
	}
	flavorProvided := false
	if req.FlavorIDs != nil {
		parsedFlavorIDs, err := parseUUIDList(*req.FlavorIDs)
		if err != nil {
			return nil, apperror.New(apperror.CodeValidationError, "invalid flavor_ids", http.StatusBadRequest)
		}
		flavorIDs = deduplicateUUIDs(parsedFlavorIDs)
		flavorProvided = true
	}

	if category.Code != domain.CategoryCodeHookahTobacco {
		if strengthProvided && strengthID != nil {
			return nil, apperror.New(apperror.CodeValidationError, "strength_id is allowed only for hookah_tobacco", http.StatusBadRequest)
		}
		if flavorProvided && len(flavorIDs) > 0 {
			return nil, apperror.New(apperror.CodeValidationError, "flavor_ids are allowed only for hookah_tobacco", http.StatusBadRequest)
		}
		strengthID = nil
		flavorIDs = []uuid.UUID{}
	} else {
		if err := s.validateTobaccoAttributes(ctx, category, isActive, strengthID, flavorIDs); err != nil {
			return nil, err
		}
	}

	if isActive && !category.IsActive {
		return nil, apperror.New(apperror.CodeValidationError, "active product cannot belong to inactive category", http.StatusBadRequest)
	}

	upsert := domain.ProductUpsert{
		ID:            productID,
		CategoryID:    categoryID,
		Name:          name,
		Description:   description,
		Price:         price,
		StockQuantity: stockQuantity,
		Unit:          unit,
		IsActive:      isActive,
		StrengthID:    strengthID,
		FlavorIDs:     deduplicateUUIDs(flavorIDs),
	}

	if err := s.repo.UpdateProduct(ctx, upsert); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to update product", http.StatusInternalServerError, err)
	}

	updated, err := s.repo.GetProductByID(ctx, productID, true)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch updated product", http.StatusInternalServerError, err)
	}
	return updated, nil
}

func (s *CatalogService) DeactivateProduct(ctx context.Context, productID uuid.UUID) error {
	if err := s.repo.DeactivateProduct(ctx, productID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to deactivate product", http.StatusInternalServerError, err)
	}
	return nil
}

func (s *CatalogService) UpdateProductStock(ctx context.Context, productID uuid.UUID, req dto.UpdateStockRequest) (*domain.Product, error) {
	if req.Quantity <= 0 {
		return nil, apperror.New(apperror.CodeValidationError, "quantity must be positive", http.StatusBadRequest)
	}

	operation := domain.StockMovementOperation(strings.ToLower(strings.TrimSpace(req.Operation)))
	switch operation {
	case domain.StockMovementOperationSet,
		domain.StockMovementOperationIncrement,
		domain.StockMovementOperationDecrement:
		// valid operation
	default:
		return nil, apperror.New(apperror.CodeInvalidStockOp, "invalid stock operation", http.StatusBadRequest)
	}

	reason := trimOptionalString(req.Reason)

	var createdByUserID *uuid.UUID
	if userID, ok := ctxkeys.UserID(ctx); ok {
		createdByUserID = &userID
	}

	if err := s.repo.ApplyProductStockOperation(ctx, productID, operation, req.Quantity, reason, createdByUserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, apperror.New(apperror.CodeInsufficientStock, "insufficient stock", http.StatusBadRequest)
		}
		if errors.Is(err, repository.ErrInvalidStockOperation) {
			return nil, apperror.New(apperror.CodeInvalidStockOp, "invalid stock operation", http.StatusBadRequest)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to update stock", http.StatusInternalServerError, err)
	}

	updated, err := s.repo.GetProductByID(ctx, productID, true)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch updated product", http.StatusInternalServerError, err)
	}
	return updated, nil
}

func (s *CatalogService) validateTobaccoAttributes(
	ctx context.Context,
	category *domain.ProductCategory,
	isActive bool,
	strengthID *uuid.UUID,
	flavorIDs []uuid.UUID,
) error {
	if category.Code != domain.CategoryCodeHookahTobacco {
		if strengthID != nil {
			return apperror.New(apperror.CodeValidationError, "strength_id is allowed only for hookah_tobacco", http.StatusBadRequest)
		}
		if len(flavorIDs) > 0 {
			return apperror.New(apperror.CodeValidationError, "flavor_ids are allowed only for hookah_tobacco", http.StatusBadRequest)
		}
		if isActive && !category.IsActive {
			return apperror.New(apperror.CodeValidationError, "active product cannot belong to inactive category", http.StatusBadRequest)
		}
		return nil
	}

	if isActive && !category.IsActive {
		return apperror.New(apperror.CodeValidationError, "active product cannot belong to inactive category", http.StatusBadRequest)
	}

	if strengthID != nil {
		strength, err := s.repo.GetStrengthByID(ctx, *strengthID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return apperror.New(apperror.CodeStrengthNotFound, "strength not found", http.StatusNotFound)
			}
			return apperror.Wrap(apperror.CodeInternal, "failed to fetch strength", http.StatusInternalServerError, err)
		}
		if isActive && !strength.IsActive {
			return apperror.New(apperror.CodeValidationError, "active product requires active strength", http.StatusBadRequest)
		}
	}

	if len(flavorIDs) > 0 {
		flavors, err := s.repo.GetFlavorsByIDs(ctx, flavorIDs)
		if err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to fetch flavors", http.StatusInternalServerError, err)
		}
		if len(flavors) != len(flavorIDs) {
			return apperror.New(apperror.CodeFlavorNotFound, "flavor not found", http.StatusNotFound)
		}
		if isActive {
			for _, flavor := range flavors {
				if !flavor.IsActive {
					return apperror.New(apperror.CodeValidationError, "active product requires active flavors", http.StatusBadRequest)
				}
			}
		}
	}

	return nil
}

func normalizeCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validatePrice(price string) error {
	if price == "" {
		return apperror.New(apperror.CodeValidationError, "price is required", http.StatusBadRequest)
	}

	val, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return apperror.New(apperror.CodeValidationError, "price must be a valid number", http.StatusBadRequest)
	}
	if val <= 0 {
		return apperror.New(apperror.CodeValidationError, "price must be greater than 0", http.StatusBadRequest)
	}

	parts := strings.Split(price, ".")
	if len(parts) > 2 {
		return apperror.New(apperror.CodeValidationError, "price must be a decimal with up to 2 digits", http.StatusBadRequest)
	}
	if len(parts) == 2 && len(parts[1]) > 2 {
		return apperror.New(apperror.CodeValidationError, "price must be a decimal with up to 2 digits", http.StatusBadRequest)
	}

	return nil
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(value))
}

func parseOptionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseUUIDList(values []string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return []uuid.UUID{}, nil
	}
	result := make([]uuid.UUID, 0, len(values))
	for _, raw := range values {
		parsed, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid uuid %q: %w", raw, err)
		}
		result = append(result, parsed)
	}
	return result, nil
}

func deduplicateUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, item := range values {
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
