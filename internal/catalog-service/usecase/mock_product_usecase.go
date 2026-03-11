package usecase

import (
	"context"
	"fmt"
	"sync"

	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/entity"
	"github.com/google/uuid"
)

type MockProductUseCase struct {
	mu       sync.RWMutex
	Products map[string]*entity.Product
}

func NewMockProductUseCase() *MockProductUseCase {
	return &MockProductUseCase{
		Products: make(map[string]*entity.Product),
	}
}

func (m *MockProductUseCase) GetProduct(ctx context.Context, id string) (*entity.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	product, ok := m.Products[id]
	if !ok {
		return nil, fmt.Errorf("product not found: %s", id)
	}
	return product, nil
}

func (m *MockProductUseCase) ListProducts(ctx context.Context, offset, limit int32, category string, onlyActive bool) ([]*entity.Product, int32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*entity.Product
	for _, p := range m.Products {
		if onlyActive && !p.IsActive {
			continue
		}
		if category != "" && p.Category != category {
			continue
		}
		result = append(result, p)
	}

	totalCount := int32(len(result))
	start := offset
	end := offset + limit

	if start >= totalCount {
		return []*entity.Product{}, totalCount, nil
	}
	if end > totalCount {
		end = totalCount
	}

	return result[start:end], totalCount, nil
}

func (m *MockProductUseCase) CreateProduct(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	product.ID = uuid.New().String()
	m.Products[product.ID] = product
	return product, nil
}

func (m *MockProductUseCase) UpdateProduct(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.Products[product.ID]; !ok {
		return nil, fmt.Errorf("product not found: %s", product.ID)
	}
	m.Products[product.ID] = product
	return product, nil
}

func (m *MockProductUseCase) DeleteProduct(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.Products[id]; !ok {
		return fmt.Errorf("product not found: %s", id)
	}
	delete(m.Products, id)
	return nil
}
