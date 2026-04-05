package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenStore struct {
	client *redis.Client
}

func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

// SaveRefreshToken saves refresh token to Redis with TTL
func (s *TokenStore) SaveRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:%s:%s", userID, token)
	return s.client.Set(ctx, key, userID, ttl).Err()
}

// IsRefreshTokenValid checks if refresh token exists in Redis
func (s *TokenStore) IsRefreshTokenValid(ctx context.Context, userID, token string) (bool, error) {
	key := fmt.Sprintf("refresh:%s:%s", userID, token)
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// DeleteRefreshToken removes refresh token (logout)
func (s *TokenStore) DeleteRefreshToken(ctx context.Context, userID, token string) error {
	key := fmt.Sprintf("refresh:%s:%s", userID, token)
	return s.client.Del(ctx, key).Err()
}

// BlacklistAccessToken adds access token to blacklist until it expires
func (s *TokenStore) BlacklistAccessToken(ctx context.Context, token string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	return s.client.Set(ctx, key, "1", ttl).Err()
}

// IsAccessTokenBlacklisted checks if access token is blacklisted
func (s *TokenStore) IsAccessTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// DeleteAllRefreshTokens removes all refresh tokens for a user (logout from all devices)
func (s *TokenStore) DeleteAllRefreshTokens(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("refresh:%s:*", userID)
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}
