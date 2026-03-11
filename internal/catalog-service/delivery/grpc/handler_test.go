package grpc

import (
	"context"
	"fmt"
	"testing"

	catalogv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/catalog/v1"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/entity"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест 1: Создание товара
func TestHandler_CreateProduct(t *testing.T) {
	// Arrange (подготовка)
	mockUC := usecase.NewMockProductUseCase()
	handler := NewHandler(mockUC)

	req := &catalogv1.CreateProductRequest{
		Name:     "iPhone 15 Pro",
		Price:    9990000,
		Stock:    150,
		Category: "Electronics",
	}

	// Act (действие)
	resp, err := handler.CreateProduct(context.Background(), req)

	// Assert (проверка)
	require.NoError(t, err)
	assert.NotNil(t, resp.Product)
	assert.NotEmpty(t, resp.Product.Id)
	assert.Equal(t, "iPhone 15 Pro", resp.Product.Name)
	assert.Equal(t, int64(9990000), resp.Product.Price)
}

// Тест 2: Получение товара
func TestHandler_GetProduct(t *testing.T) {
	// Arrange
	mockUC := usecase.NewMockProductUseCase()
	handler := NewHandler(mockUC)

	// Сначала создаём товар через Mock
	_, _ = mockUC.CreateProduct(context.Background(), &entity.Product{
		Name:  "Test Product",
		Price: 1000,
		Stock: 10,
	})

	// Получаем ID созданного товара
	var productID string
	for id := range mockUC.Products {
		productID = id
		break
	}

	// Act
	resp, err := handler.GetProduct(context.Background(), &catalogv1.GetProductRequest{
		Id: productID,
	})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Test Product", resp.Product.Name)
}

// Тест 3: Товар не найден
func TestHandler_GetProduct_NotFound(t *testing.T) {
	// Arrange
	mockUC := usecase.NewMockProductUseCase()
	handler := NewHandler(mockUC)

	// Act
	resp, err := handler.GetProduct(context.Background(), &catalogv1.GetProductRequest{
		Id: "non-existent-id",
	})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Тест 4: Список товаров
func TestHandler_ListProducts(t *testing.T) {
	// Arrange
	mockUC := usecase.NewMockProductUseCase()
	handler := NewHandler(mockUC)

	// Добавляем несколько товаров
	for i := 0; i < 5; i++ {
		_, _ = mockUC.CreateProduct(context.Background(), &entity.Product{
			Name:  fmt.Sprintf("Product %d", i),
			Price: 1000,
			Stock: 10,
		})
	}

	// Act
	resp, err := handler.ListProducts(context.Background(), &catalogv1.ListProductsRequest{
		Page:     1,
		PageSize: 10,
	})

	// Assert
	require.NoError(t, err)
	assert.Len(t, resp.Products, 5)
	assert.Equal(t, int32(5), resp.Pagination.TotalCount)
}
