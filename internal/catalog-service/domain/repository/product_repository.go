package repository

import (
	"context"

	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/entity"
)

// ProductRepository — интерфейс репозитория для товаров
type ProductRepository interface {
	// Получить товар по ID
	GetByID(ctx context.Context, id string) (*entity.Product, error)
	// Получить список товаров с пагинацией
	List(ctx context.Context, limit, offset int32, category string, onlyActive bool) ([]*entity.Product, int32, error)
	// Создать новый товар
	Create(ctx context.Context, product *entity.Product) error
	// Обновить товар
	Update(ctx context.Context, product *entity.Product) error
	// Удалить товар
	Delete(ctx context.Context, id string) error
}
