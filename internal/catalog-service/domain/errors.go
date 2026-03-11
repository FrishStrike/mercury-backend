package domain

import (
	"errors"
	"fmt"
)

// Ошибки доменного уровня
var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidPrice    = errors.New("invalid price: must be >= 0")
	ErrInvalidStock    = errors.New("invalid stock: must be >= 0")
	ErrEmptyName       = errors.New("product name cannot be empty")
	ErrInvalidQuantity = errors.New("invalid quantity: must be > 0")
)

// DomainError — обёртка для доменных ошибок
type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewValidationError — создать ошибку валидации
func NewValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:    "VALIDATION_ERROR",
		Message: fmt.Sprintf("%s: %s", field, message),
	}
}

// NewNotFoundError — создать ошибку не найдено
func NewNotFoundError(entity, id string) *DomainError {
	return &DomainError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s with id %s not found", entity, id),
	}
}
