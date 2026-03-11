package usecase

import (
	"context"

	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/entity"
)

// ProductUseCase — интерфейс бизнес-логики для товаров
//
//go:generate mockgen -source=product_usecase.go -destination=mock_product_usecase_gen.go -package=usecase
type ProductUseCase interface {
	// Получить товар по ID
	GetProduct(ctx context.Context, id string) (*entity.Product, error)
	// Получить список товаров (offset, limit для пагинации)
	ListProducts(ctx context.Context, offset, limit int32, category string, onlyActive bool) ([]*entity.Product, int32, error)
	// Создать новый товар
	CreateProduct(ctx context.Context, product *entity.Product) (*entity.Product, error)
	// Обновить товар
	UpdateProduct(ctx context.Context, product *entity.Product) (*entity.Product, error)
	// Удалить товар
	DeleteProduct(ctx context.Context, id string) error
}
