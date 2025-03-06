package slack

import (
	"fmt"
	"log"

	"github.com/mcncl/snagbot/pkg/models"
)

// In-memory storage for channel configurations
// In a production environment, this would be replaced with a persistent database
var channelConfigs = make(map[string]*models.ChannelConfig)

// ChannelConfigStore interface for storing channel configurations
type ChannelConfigStore interface {
	GetConfig(channelID string) *models.ChannelConfig
	UpdateConfig(channelID, itemName string, itemPrice float64) error
}

// InMemoryConfigStore provides a simple in-memory implementation of ChannelConfigStore
type InMemoryConfigStore struct {
	configs map[string]*models.ChannelConfig
}

// NewInMemoryConfigStore creates a new in-memory config store
func NewInMemoryConfigStore() *InMemoryConfigStore {
	return &InMemoryConfigStore{
		configs: make(map[string]*models.ChannelConfig),
	}
}

// GetConfig retrieves the channel configuration or returns a default one
func (s *InMemoryConfigStore) GetConfig(channelID string) *models.ChannelConfig {
	if config, ok := s.configs[channelID]; ok {
		return config
	}

	// Create new default config
	newConfig := models.NewChannelConfig(channelID)
	s.configs[channelID] = newConfig
	return newConfig
}

// UpdateConfig updates the configuration for a channel
func (s *InMemoryConfigStore) UpdateConfig(channelID, itemName string, itemPrice float64) error {
	if itemPrice <= 0 {
		return fmt.Errorf("item price must be greater than zero")
	}

	// Get or create channel config
	config := s.GetConfig(channelID)

	// Update the configuration
	config.SetItem(itemName, itemPrice)
	log.Printf("Updated configuration for channel %s: item=%s, price=%0.2f",
		channelID, itemName, itemPrice)

	return nil
}

// DatabaseChannelConfigStore would implement a database-backed channel config store
// This is just a placeholder showing how it would be structured
type DatabaseChannelConfigStore struct {
	// db would be some database connection or ORM
}

// GetConfig would retrieve the channel configuration from the database
func (s *DatabaseChannelConfigStore) GetConfig(channelID string) *models.ChannelConfig {
	// Implementation would retrieve from database
	// For now, we'll return a default config
	return models.NewChannelConfig(channelID)
}

// UpdateConfig would update the configuration in the database
func (s *DatabaseChannelConfigStore) UpdateConfig(channelID, itemName string, itemPrice float64) error {
	// Implementation would update in database
	return nil
}

// UpdateChannelConfig updates the configuration for a channel
func UpdateChannelConfig(channelID, itemName string, itemPrice float64) error {
	if itemPrice <= 0 {
		return fmt.Errorf("item price must be greater than zero")
	}

	// Get or create channel config
	config, ok := channelConfigs[channelID]
	if !ok {
		config = models.NewChannelConfig(channelID)
		channelConfigs[channelID] = config
	}

	// Update the configuration
	config.SetItem(itemName, itemPrice)
	log.Printf("Updated configuration for channel %s: item=%s, price=%0.2f",
		channelID, itemName, itemPrice)

	return nil
}
