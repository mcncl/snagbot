package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/pkg/models"
)

// RedisConfigStore implements ChannelConfigStore using Redis
type RedisConfigStore struct {
	client  *redis.Client
	ctx     context.Context
	appCfg  *config.Config
	keyBase string
}

// NewRedisConfigStore creates a new Redis-backed configuration store
func NewRedisConfigStore(redisURL string, appCfg *config.Config) (*RedisConfigStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing Redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	// Test connection
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Redis: %w", err)
	}

	return &RedisConfigStore{
		client:  client,
		ctx:     ctx,
		appCfg:  appCfg,
		keyBase: "snagbot:channel_config:",
	}, nil
}

// getConfigKey returns the Redis key for a channel's configuration
func (s *RedisConfigStore) getConfigKey(channelID string) string {
	return s.keyBase + channelID
}

// GetConfig retrieves a channel's configuration or returns the default
func (s *RedisConfigStore) GetConfig(channelID string) (*models.ChannelConfig, error) {
	key := s.getConfigKey(channelID)
	
	// Check if the config exists
	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("error checking if config exists: %w", err)
	}
	
	// If config doesn't exist, return a new one with defaults
	if exists == 0 {
		return &models.ChannelConfig{
			ChannelID: channelID,
			ItemName:  s.appCfg.DefaultItemName,
			ItemPrice: s.appCfg.DefaultItemPrice,
		}, nil
	}
	
	// Get the stored config
	jsonData, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("error retrieving config from Redis: %w", err)
	}
	
	// Unmarshal the JSON data
	var config models.ChannelConfig
	if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	return &config, nil
}

// UpdateConfig updates or creates a channel's configuration
func (s *RedisConfigStore) UpdateConfig(channelID, itemName string, itemPrice float64) error {
	config := &models.ChannelConfig{
		ChannelID: channelID,
		ItemName:  itemName,
		ItemPrice: itemPrice,
	}
	
	// Marshal the config to JSON
	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}
	
	// Store in Redis with 30-day expiry
	key := s.getConfigKey(channelID)
	err = s.client.Set(s.ctx, key, jsonData, 30*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error storing config in Redis: %w", err)
	}
	
	return nil
}

// ResetConfig removes a channel's configuration so it uses defaults
func (s *RedisConfigStore) ResetConfig(channelID string) error {
	key := s.getConfigKey(channelID)
	err := s.client.Del(s.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error deleting config from Redis: %w", err)
	}
	
	return nil
}

// ConfigExists checks if a custom configuration exists for a channel
func (s *RedisConfigStore) ConfigExists(channelID string) bool {
	key := s.getConfigKey(channelID)
	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		// Log error and default to false
		fmt.Printf("Error checking if config exists: %v\n", err)
		return false
	}
	
	return exists > 0
}

// Close closes the Redis connection
func (s *RedisConfigStore) Close() error {
	return s.client.Close()
}