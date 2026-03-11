package domain

import (
	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/entity"
)

// validateProduct — валидация товара
func ValidateProduct(product *entity.Product) error {
	if product == nil {
		return NewValidationError("product", "cannot be nil")
	}

	if product.Name == "" {
		return NewValidationError("name", "cannot be empty")
	}

	if len(product.Name) > 255 {
		return NewValidationError("name", "must be less than 255 characters")
	}

	if product.Price < 0 {
		return NewValidationError("price", "must be >= 0")
	}

	if product.Stock < 0 {
		return NewValidationError("stock", "must be >= 0")
	}

	return nil
}

// validateProductID — валидация ID товара
func ValidateProductID(id string) error {
	if id == "" {
		return NewValidationError("id", "cannot be empty")
	}

	// Проверка формата UUID (опционально)
	if len(id) != 36 {
		return NewValidationError("id", "must be a valid UUID")
	}

	return nil
}

// validatePagination — валидация пагинации
func ValidatePagination(offset, limit int32) error {
	if offset < 0 {
		return NewValidationError("offset", "must be >= 0")
	}

	if limit <= 0 {
		return NewValidationError("limit", "must be > 0")
	}

	if limit > 100 {
		return NewValidationError("limit", "must be <= 100")
	}

	return nil
}
