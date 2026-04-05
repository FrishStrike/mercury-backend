package usecase

import (
	"context"

	"github.com/frishstrike/mercury-backend/internal/auth-service/domain/entity"
	"github.com/frishstrike/mercury-backend/pkg/auth"
)

type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*entity.User, *auth.TokenPair, error)
	Login(ctx context.Context, input LoginInput) (*entity.User, *auth.TokenPair, error)
	Logout(ctx context.Context, userID, accessToken, refreshToken string) error
	RefreshTokens(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	GetUser(ctx context.Context, userID string) (*entity.User, error)
}
