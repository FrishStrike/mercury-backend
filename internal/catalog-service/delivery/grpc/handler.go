package grpc

import (
	"context"
	"errors"

	catalogv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/catalog/v1"
	commonv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/common/v1"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/domain/entity"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	catalogv1.UnimplementedCatalogServiceServer
	uc usecase.ProductUseCase
}

func NewHandler(uc usecase.ProductUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	product, err := h.uc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &catalogv1.GetProductResponse{Product: product.ToProto()}, nil
}

func (h *Handler) ListProducts(ctx context.Context, req *catalogv1.ListProductsRequest) (*catalogv1.ListProductsResponse, error) {
	offset := (req.Page - 1) * req.PageSize

	products, totalCount, err := h.uc.ListProducts(ctx, offset, req.PageSize, req.Category, req.OnlyActive)
	if err != nil {
		return nil, toGRPCError(err)
	}

	protoProducts := make([]*catalogv1.Product, 0, len(products))
	for _, p := range products {
		protoProducts = append(protoProducts, p.ToProto())
	}

	return &catalogv1.ListProductsResponse{
		Products: protoProducts,
		Pagination: &commonv1.Pagination{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalCount: totalCount,
		},
	}, nil
}

func (h *Handler) CreateProduct(ctx context.Context, req *catalogv1.CreateProductRequest) (*catalogv1.CreateProductResponse, error) {
	product := &entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		IsActive:    true,
	}

	created, err := h.uc.CreateProduct(ctx, product)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &catalogv1.CreateProductResponse{Product: created.ToProto()}, nil
}

func (h *Handler) UpdateProduct(ctx context.Context, req *catalogv1.UpdateProductRequest) (*catalogv1.UpdateProductResponse, error) {
	product := &entity.Product{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		IsActive:    req.IsActive,
	}

	updated, err := h.uc.UpdateProduct(ctx, product)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &catalogv1.UpdateProductResponse{Product: updated.ToProto()}, nil
}

func (h *Handler) DeleteProduct(ctx context.Context, req *catalogv1.DeleteProductRequest) (*commonv1.Empty, error) {
	if err := h.uc.DeleteProduct(ctx, req.Id); err != nil {
		return nil, toGRPCError(err)
	}
	return &commonv1.Empty{}, nil
}

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case "NOT_FOUND":
			return status.Error(codes.NotFound, domainErr.Message)
		case "VALIDATION_ERROR":
			return status.Error(codes.InvalidArgument, domainErr.Message)
		}
	}

	return status.Error(codes.Internal, err.Error())
}
