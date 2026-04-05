package repository

import (
	"context"

	"github.com/frishstrike/mercury-backend/internal/auth-service/domain/entity"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}
