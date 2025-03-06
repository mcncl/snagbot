package slack

import (
	"sync"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/errors"
	"github.com/mcncl/snagbot/internal/logging"
	"github.com/mcncl/snagbot/pkg/models"
)

// ChannelConfigStore interface for storing channel configurations
type ChannelConfigStore interface {
	GetConfig(channelID string) (*models.ChannelConfig, error)
	UpdateConfig(channelID, itemName string, itemPrice float64) error
	ResetConfig(channelID string) error
	ConfigExists(channelID string) bool
}

// InMemoryConfigStore provides a simple in-memory implementation of ChannelConfigStore
type InMemoryConfigStore struct {
	configs map[string]*models.ChannelConfig
	mutex   sync.RWMutex
	cfg     *config.Config
}

// NewInMemoryConfigStore creates a new in-memory config store
// For backwards compatibility, this now wraps the new implementation
func NewInMemoryConfigStore() *InMemoryConfigStore {
	return NewInMemoryConfigStoreWithConfig(nil)
}

// NewInMemoryConfigStoreWithConfig creates a new in-memory config store with provided configuration
// This is the new function that takes a config parameter
func NewInMemoryConfigStoreWithConfig(cfg *config.Config) *InMemoryConfigStore {
	logging.Debug("Creating new in-memory config store")
	return &InMemoryConfigStore{
		configs: make(map[string]*models.ChannelConfig),
		cfg:     cfg,
	}
}

// GetConfig retrieves the channel configuration or returns a default one
func (s *InMemoryConfigStore) GetConfig(channelID string) (*models.ChannelConfig, error) {
	if channelID == "" {
		return nil, errors.New(errors.ErrInvalidRequest, "empty channel ID")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if config, ok := s.configs[channelID]; ok {
		logging.Debug("Found existing configuration for channel %s", channelID)
		// Return a copy to prevent concurrent modification issues
		return &models.ChannelConfig{
			ChannelID: config.ChannelID,
			ItemName:  config.ItemName,
			ItemPrice: config.ItemPrice,
		}, nil
	}

	// Create new default config using application defaults
	var defaultItemName string
	var defaultItemPrice float64

	if s.cfg != nil {
		defaultItemName = s.cfg.DefaultItemName
		defaultItemPrice = s.cfg.DefaultItemPrice
	} else {
		// Fallback to hardcoded defaults if no config is provided
		defaultItemName = "Bunnings snags"
		defaultItemPrice = 3.50
	}

	logging.Debug("No configuration found for channel %s, using defaults: %s at $%.2f",
		channelID, defaultItemName, defaultItemPrice)

	newConfig := &models.ChannelConfig{
		ChannelID: channelID,
		ItemName:  defaultItemName,
		ItemPrice: defaultItemPrice,
	}

	// We don't store this default config in the map to avoid memory bloat from channels
	// that may only query the config once and never use it again
	return newConfig, nil
}

// UpdateConfig updates the configuration for a channel
func (s *InMemoryConfigStore) UpdateConfig(channelID, itemName string, itemPrice float64) error {
	if channelID == "" {
		return errors.New(errors.ErrInvalidRequest, "empty channel ID")
	}

	if itemPrice <= 0 {
		return errors.Newf(errors.ErrInvalidRequest, "item price must be greater than zero: %.2f", itemPrice)
	}

	if itemName == "" {
		return errors.New(errors.ErrInvalidRequest, "item name cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get existing config or create a new one
	var config *models.ChannelConfig
	var ok bool

	if config, ok = s.configs[channelID]; !ok {
		// If config doesn't exist, create a new one
		config = &models.ChannelConfig{
			ChannelID: channelID,
		}
		s.configs[channelID] = config
	}

	// Update the configuration
	config.ItemName = itemName
	config.ItemPrice = itemPrice

	logging.Info("Updated configuration for channel %s: item=%s, price=%.2f",
		channelID, itemName, itemPrice)

	return nil
}

// ResetConfig resets a channel's configuration to the default
func (s *InMemoryConfigStore) ResetConfig(channelID string) error {
	if channelID == "" {
		return errors.New(errors.ErrInvalidRequest, "empty channel ID")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if config exists before deleting
	if _, ok := s.configs[channelID]; !ok {
		// Already using defaults, nothing to do
		logging.Debug("Channel %s already using defaults, no reset needed", channelID)
		return nil
	}

	// Delete the config from the map
	delete(s.configs, channelID)
	logging.Info("Reset configuration for channel %s to default", channelID)

	return nil
}

// ConfigExists checks if a custom configuration exists for a channel
func (s *InMemoryConfigStore) ConfigExists(channelID string) bool {
	if channelID == "" {
		logging.Warn("ConfigExists called with empty channel ID")
		return false
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.configs[channelID]
	return exists
}

// GetAllChannelIDs returns a list of all channel IDs that have custom configs
func (s *InMemoryConfigStore) GetAllChannelIDs() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	channelIDs := make([]string, 0, len(s.configs))
	for id := range s.configs {
		channelIDs = append(channelIDs, id)
	}
	logging.Debug("Retrieved %d channel IDs with custom configurations", len(channelIDs))
	return channelIDs
}

// Count returns the number of stored channel configurations
func (s *InMemoryConfigStore) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.configs)
}

// BackupConfigs returns a copy of all configurations for backup
func (s *InMemoryConfigStore) BackupConfigs() map[string]models.ChannelConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	backup := make(map[string]models.ChannelConfig, len(s.configs))
	for id, config := range s.configs {
		backup[id] = *config
	}
	logging.Debug("Created backup of %d channel configurations", len(backup))
	return backup
}

// RestoreConfigs restores configurations from a backup
func (s *InMemoryConfigStore) RestoreConfigs(backup map[string]models.ChannelConfig) error {
	if backup == nil {
		return errors.New(errors.ErrInvalidRequest, "nil backup data")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clear existing configs
	s.configs = make(map[string]*models.ChannelConfig, len(backup))

	// Restore from backup
	for id, config := range backup {
		// Create a copy to avoid issues with map values
		copyConfig := config
		s.configs[id] = &copyConfig
	}

	logging.Info("Restored %d channel configurations from backup", len(backup))
	return nil
}

// Global store instance for backward compatibility and testing
var globalConfigStore ChannelConfigStore = NewInMemoryConfigStore()
