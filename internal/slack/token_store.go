package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/logging"
	"github.com/mcncl/snagbot/pkg/models"
)

// TokenStore interface for workspace token operations
type TokenStore interface {
	SaveToken(token *models.WorkspaceToken) error
	GetToken(workspaceID string) (*models.WorkspaceToken, error)
	DeleteToken(workspaceID string) error
	ListWorkspaces() ([]string, error)
}

// RedisTokenStore implements token storage using Redis
type RedisTokenStore struct {
	client  *redis.Client
	ctx     context.Context
	keyBase string
}

// NewRedisTokenStore creates a new Redis-backed token store
func NewRedisTokenStore(redisClient *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{
		client:  redisClient,
		ctx:     context.Background(),
		keyBase: "snagbot:workspace_token:",
	}
}

// getTokenKey returns the Redis key for a workspace token
func (s *RedisTokenStore) getTokenKey(workspaceID string) string {
	return s.keyBase + workspaceID
}

// SaveToken saves a workspace token to Redis
func (s *RedisTokenStore) SaveToken(token *models.WorkspaceToken) error {
	if token.WorkspaceID == "" {
		return errors.New("workspace ID is required")
	}

	jsonData, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("error marshaling token: %w", err)
	}

	key := s.getTokenKey(token.WorkspaceID)
	err = s.client.Set(s.ctx, key, jsonData, 365*24*time.Hour).Err() // 1 year expiry
	if err != nil {
		return fmt.Errorf("error storing token in Redis: %w", err)
	}

	// Also add to workspace index
	indexKey := "snagbot:workspaces"
	err = s.client.SAdd(s.ctx, indexKey, token.WorkspaceID).Err()
	if err != nil {
		logging.Warn("Failed to add workspace to index: %v", err)
	}

	return nil
}

// GetToken retrieves a workspace token from Redis
func (s *RedisTokenStore) GetToken(workspaceID string) (*models.WorkspaceToken, error) {
	key := s.getTokenKey(workspaceID)
	
	jsonData, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("token not found for workspace %s", workspaceID)
		}
		return nil, fmt.Errorf("error retrieving token from Redis: %w", err)
	}
	
	var token models.WorkspaceToken
	if err := json.Unmarshal([]byte(jsonData), &token); err != nil {
		return nil, fmt.Errorf("error unmarshaling token: %w", err)
	}
	
	return &token, nil
}

// DeleteToken removes a workspace token from Redis
func (s *RedisTokenStore) DeleteToken(workspaceID string) error {
	key := s.getTokenKey(workspaceID)
	err := s.client.Del(s.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error deleting token from Redis: %w", err)
	}
	
	// Also remove from workspace index
	indexKey := "snagbot:workspaces"
	err = s.client.SRem(s.ctx, indexKey, workspaceID).Err()
	if err != nil {
		logging.Warn("Failed to remove workspace from index: %v", err)
	}
	
	return nil
}

// ListWorkspaces lists all workspace IDs
func (s *RedisTokenStore) ListWorkspaces() ([]string, error) {
	indexKey := "snagbot:workspaces"
	return s.client.SMembers(s.ctx, indexKey).Result()
}

// SingleTokenStore is a simple implementation for single-workspace deployment
type SingleTokenStore struct {
	token *models.WorkspaceToken
}

// NewSingleTokenStore creates a token store with a fixed token
func NewSingleTokenStore(cfg *config.Config) *SingleTokenStore {
	token := &models.WorkspaceToken{
		WorkspaceID: "single",
		AccessToken: cfg.SlackBotToken,
		TokenType:   "bot",
		InstalledAt: time.Now(),
		LastUpdated: time.Now(),
	}
	
	return &SingleTokenStore{
		token: token,
	}
}

// SaveToken is a no-op for SingleTokenStore (always returns the configured token)
func (s *SingleTokenStore) SaveToken(token *models.WorkspaceToken) error {
	// No-op for single token store
	return nil
}

// GetToken always returns the configured token
func (s *SingleTokenStore) GetToken(workspaceID string) (*models.WorkspaceToken, error) {
	return s.token, nil
}

// DeleteToken is a no-op for SingleTokenStore
func (s *SingleTokenStore) DeleteToken(workspaceID string) error {
	// No-op for single token store
	return nil
}

// ListWorkspaces returns a single workspace ID
func (s *SingleTokenStore) ListWorkspaces() ([]string, error) {
	return []string{"single"}, nil
}