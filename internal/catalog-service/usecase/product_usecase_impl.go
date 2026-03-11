package usecase

import (
	"context"
	"time"

	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/entity"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/repository"
	"github.com/google/uuid"
)

type productUseCase struct {
	repo repository.ProductRepository
}

func NewProductUseCase(repo repository.ProductRepository) ProductUseCase {
	return &productUseCase{repo: repo}
}

func (uc *productUseCase) GetProduct(ctx context.Context, id string) (*entity.Product, error) {
	if err := domain.ValidateProductID(id); err != nil {
		return nil, err
	}

	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewNotFoundError("Product", id)
	}

	return product, nil
}

func (uc *productUseCase) ListProducts(ctx context.Context, offset, limit int32, category string, onlyActive bool) ([]*entity.Product, int32, error) {
	if err := domain.ValidatePagination(offset, limit); err != nil {
		return nil, 0, err
	}

	return uc.repo.List(ctx, offset, limit, category, onlyActive)
}

func (uc *productUseCase) CreateProduct(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	if err := domain.ValidateProduct(product); err != nil {
		return nil, err
	}

	product.ID = uuid.New().String()
	product.CreatedAt = time.Now().UTC()
	product.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *productUseCase) UpdateProduct(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	if err := domain.ValidateProductID(product.ID); err != nil {
		return nil, err
	}

	existing, err := uc.repo.GetByID(ctx, product.ID)
	if err != nil {
		return nil, domain.NewNotFoundError("Product", product.ID)
	}

	if err := domain.ValidateProduct(product); err != nil {
		return nil, err
	}

	product.CreatedAt = existing.CreatedAt
	product.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *productUseCase) DeleteProduct(ctx context.Context, id string) error {
	if err := domain.ValidateProductID(id); err != nil {
		return err
	}

	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return domain.NewNotFoundError("Product", id)
	}

	return uc.repo.Delete(ctx, id)
}
