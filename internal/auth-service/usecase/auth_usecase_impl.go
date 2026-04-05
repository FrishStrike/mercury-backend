package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/frishstrike/mercury-backend/internal/auth-service/domain"
	"github.com/frishstrike/mercury-backend/internal/auth-service/domain/entity"
	domainrepo "github.com/frishstrike/mercury-backend/internal/auth-service/domain/repository"
	domainuc "github.com/frishstrike/mercury-backend/internal/auth-service/domain/usecase"
	"github.com/frishstrike/mercury-backend/pkg/auth"
	"github.com/google/uuid"
)

type authUseCase struct {
	userRepo   domainrepo.UserRepository
	jwtManager *auth.Manager
	tokenStore *auth.TokenStore
}

func NewAuthUseCase(
	userRepo domainrepo.UserRepository,
	jwtManager *auth.Manager,
	tokenStore *auth.TokenStore,
) domainuc.AuthUseCase {
	return &authUseCase{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		tokenStore: tokenStore,
	}
}

func (uc *authUseCase) Register(ctx context.Context, input domainuc.RegisterInput) (*entity.User, *auth.TokenPair, error) {
	// Check if user already exists
	existing, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &entity.User{
		ID:           uuid.New().String(),
		Email:        input.Email,
		PasswordHash: passwordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Role:         entity.RoleUser,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("create user: %w", err)
	}

	// Generate tokens
	tokens, err := uc.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("generate tokens: %w", err)
	}

	// Save refresh token to Redis
	if err := uc.tokenStore.SaveRefreshToken(ctx, user.ID, tokens.RefreshToken, 7*24*time.Hour); err != nil {
		return nil, nil, fmt.Errorf("save refresh token: %w", err)
	}

	return user, tokens, nil
}

func (uc *authUseCase) Login(ctx context.Context, input domainuc.LoginInput) (*entity.User, *auth.TokenPair, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, domain.ErrUserInactive
	}

	// Check password
	if !auth.CheckPassword(input.Password, user.PasswordHash) {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := uc.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("generate tokens: %w", err)
	}

	// Save refresh token to Redis
	if err := uc.tokenStore.SaveRefreshToken(ctx, user.ID, tokens.RefreshToken, 7*24*time.Hour); err != nil {
		return nil, nil, fmt.Errorf("save refresh token: %w", err)
	}

	return user, tokens, nil
}

func (uc *authUseCase) Logout(ctx context.Context, userID, accessToken, refreshToken string) error {
	// Validate access token to get expiry
	claims, err := uc.jwtManager.ValidateToken(accessToken)
	if err != nil && !errors.Is(err, auth.ErrExpiredToken) {
		return domain.ErrInvalidToken
	}

	// Blacklist access token
	if claims != nil {
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			if err := uc.tokenStore.BlacklistAccessToken(ctx, accessToken, ttl); err != nil {
				return fmt.Errorf("blacklist token: %w", err)
			}
		}
	}

	// Delete refresh token
	if err := uc.tokenStore.DeleteRefreshToken(ctx, userID, refreshToken); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

func (uc *authUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	// Validate refresh token
	claims, err := uc.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if claims.Type != auth.RefreshToken {
		return nil, domain.ErrInvalidToken
	}

	// Check if refresh token exists in Redis
	valid, err := uc.tokenStore.IsRefreshTokenValid(ctx, claims.UserID, refreshToken)
	if err != nil || !valid {
		return nil, domain.ErrInvalidToken
	}

	// Delete old refresh token
	if err := uc.tokenStore.DeleteRefreshToken(ctx, claims.UserID, refreshToken); err != nil {
		return nil, fmt.Errorf("delete old refresh token: %w", err)
	}

	// Generate new token pair
	tokens, err := uc.jwtManager.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	// Save new refresh token
	if err := uc.tokenStore.SaveRefreshToken(ctx, claims.UserID, tokens.RefreshToken, 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return tokens, nil
}

func (uc *authUseCase) GetUser(ctx context.Context, userID string) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}
