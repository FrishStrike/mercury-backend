package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/entity"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) repository.ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*entity.Product, error) {
	query := `
        SELECT id, name, description, price, stock, category, is_active, created_at, updated_at
        FROM products
        WHERE id = $1
    `

	var product entity.Product
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.Category,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("product not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) List(ctx context.Context, offset, limit int32, category string, onlyActive bool) ([]*entity.Product, int32, error) {
	query := `
        SELECT id, name, description, price, stock, category, is_active, created_at, updated_at
        FROM products
        WHERE 1=1
    `

	countQuery := `SELECT COUNT(*) FROM products WHERE 1=1`

	args := []any{}
	argIndex := 1

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	if onlyActive {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, true)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	var totalCount int32
	err := r.pool.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		var product entity.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.Category,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, totalCount, nil
}

func (r *ProductRepository) Create(ctx context.Context, product *entity.Product) error {
	query := `
        INSERT INTO products (id, name, description, price, stock, category, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `

	_, err := r.pool.Exec(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.Category,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

func (r *ProductRepository) Update(ctx context.Context, product *entity.Product) error {
	query := `
        UPDATE products
        SET name = $2, description = $3, price = $4, stock = $5, 
            category = $6, is_active = $7, updated_at = $8
        WHERE id = $1
    `

	result, err := r.pool.Exec(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.Category,
		product.IsActive,
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found: %s", product.ID)
	}

	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found: %s", id)
	}

	return nil
}
