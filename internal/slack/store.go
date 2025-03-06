package slack

import (
	"fmt"
	"log"
	"sync"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/pkg/models"
)

// ChannelConfigStore interface for storing channel configurations
type ChannelConfigStore interface {
	GetConfig(channelID string) *models.ChannelConfig
	UpdateConfig(channelID, itemName string, itemPrice float64) error
	ResetConfig(channelID string) error
	ConfigExists(channelID string) bool
}

// InMemoryConfigStore provides a simple in-memory implementation of ChannelConfigStore
type InMemoryConfigStore struct {
	configs map[string]*models.ChannelConfig
	mutex   sync.RWMutex   // Adding a mutex for thread safety
	cfg     *config.Config // Application default config for fallback values
}

// NewInMemoryConfigStore creates a new in-memory config store
// For backwards compatibility, this now wraps the new implementation
func NewInMemoryConfigStore() *InMemoryConfigStore {
	return NewInMemoryConfigStoreWithConfig(nil)
}

// NewInMemoryConfigStoreWithConfig creates a new in-memory config store with provided configuration
// This is the new function that takes a config parameter
func NewInMemoryConfigStoreWithConfig(cfg *config.Config) *InMemoryConfigStore {
	return &InMemoryConfigStore{
		configs: make(map[string]*models.ChannelConfig),
		cfg:     cfg,
	}
}

// GetConfig retrieves the channel configuration or returns a default one
func (s *InMemoryConfigStore) GetConfig(channelID string) *models.ChannelConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if config, ok := s.configs[channelID]; ok {
		return config
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

	newConfig := &models.ChannelConfig{
		ChannelID: channelID,
		ItemName:  defaultItemName,
		ItemPrice: defaultItemPrice,
	}

	// We don't store this default config in the map to avoid memory bloat from channels
	// that may only query the config once and never use it again
	return newConfig
}

// UpdateConfig updates the configuration for a channel
func (s *InMemoryConfigStore) UpdateConfig(channelID, itemName string, itemPrice float64) error {
	if itemPrice <= 0 {
		return fmt.Errorf("item price must be greater than zero")
	}

	if itemName == "" {
		return fmt.Errorf("item name cannot be empty")
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

	log.Printf("Updated configuration for channel %s: item=%s, price=%0.2f",
		channelID, itemName, itemPrice)

	return nil
}

// ResetConfig resets a channel's configuration to the default
func (s *InMemoryConfigStore) ResetConfig(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Delete the config from the map
	delete(s.configs, channelID)
	log.Printf("Reset configuration for channel %s to default", channelID)

	return nil
}

// ConfigExists checks if a custom configuration exists for a channel
func (s *InMemoryConfigStore) ConfigExists(channelID string) bool {
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
	return channelIDs
}

// Count returns the number of stored channel configurations
func (s *InMemoryConfigStore) Count() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.configs)
}

// Global store instance for backward compatibility
var globalConfigStore = NewInMemoryConfigStore()
