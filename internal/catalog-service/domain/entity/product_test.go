package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProduct_ToProto(t *testing.T) {
	tests := []struct {
		name      string
		product   *Product
		wantName  string
		wantPrice int64
	}{
		{
			name: "valid product",
			product: &Product{
				ID:          "uuid-123",
				Name:        "iPhone 15 Pro",
				Description: "256GB Natural Titanium",
				Price:       9990000,
				Stock:       150,
				Category:    "Electronics",
				IsActive:    true,
				CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantName:  "iPhone 15 Pro",
			wantPrice: 9990000,
		},
		{
			name: "empty product",
			product: &Product{
				ID:        "uuid-456",
				Name:      "",
				Price:     0,
				Stock:     0,
				IsActive:  false,
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
			wantName:  "",
			wantPrice: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto := tt.product.ToProto()

			assert.NotNil(t, proto)
			assert.Equal(t, tt.product.ID, proto.Id)
			assert.Equal(t, tt.wantName, proto.Name)
			assert.Equal(t, tt.wantPrice, proto.Price)
			assert.Equal(t, tt.product.Stock, proto.Stock)
			assert.Equal(t, tt.product.Category, proto.Category)
			assert.Equal(t, tt.product.IsActive, proto.IsActive)
		})
	}
}
