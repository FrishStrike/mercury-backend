package entity

import (
	"time"

	catalogv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/catalog/v1"
)

// Product — сущность товара
type Product struct {
	ID          string
	Name        string
	Description string
	Price       int64 // в копейках
	Stock       int32 // остаток на складе
	Category    string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ToProto — конвертация в proto-сообщение
func (p *Product) ToProto() *catalogv1.Product {
	return &catalogv1.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Category:    p.Category,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}
