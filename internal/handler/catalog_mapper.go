package handler

import (
	"time"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/dto"
)

func toCategoryResponse(item domain.ProductCategory) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:          item.ID.String(),
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		IsActive:    item.IsActive,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toFlavorResponse(item domain.TobaccoFlavor) dto.FlavorResponse {
	return dto.FlavorResponse{
		ID:          item.ID.String(),
		Name:        item.Name,
		Description: item.Description,
		IsActive:    item.IsActive,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toStrengthResponse(item domain.TobaccoStrength) dto.StrengthResponse {
	return dto.StrengthResponse{
		ID:          item.ID.String(),
		Name:        item.Name,
		Level:       item.Level,
		Description: item.Description,
		IsActive:    item.IsActive,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toProductResponse(item *domain.Product) dto.ProductResponse {
	flavors := make([]dto.ProductFlavorInfo, 0, len(item.Flavors))
	for _, flavor := range item.Flavors {
		flavors = append(flavors, dto.ProductFlavorInfo{
			ID:   flavor.ID.String(),
			Name: flavor.Name,
		})
	}
	tags := make([]dto.ProductTagInfo, 0, len(item.Tags))
	for _, tag := range item.Tags {
		tags = append(tags, dto.ProductTagInfo{
			ID:   tag.ID.String(),
			Code: tag.Code,
			Name: tag.Name,
		})
	}

	var strength *dto.ProductStrengthInfo
	if item.Strength != nil {
		strength = &dto.ProductStrengthInfo{
			ID:    item.Strength.ID.String(),
			Name:  item.Strength.Name,
			Level: item.Strength.Level,
		}
	}

	return dto.ProductResponse{
		ID:            item.ID.String(),
		Name:          item.Name,
		Description:   item.Description,
		Price:         item.Price,
		StockQuantity: item.StockQuantity,
		Unit:          item.Unit,
		IsActive:      item.IsActive,
		Category: dto.ProductCategoryInfo{
			ID:   item.Category.ID.String(),
			Code: item.Category.Code,
			Name: item.Category.Name,
		},
		Strength:  strength,
		Flavors:   flavors,
		Tags:      tags,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
