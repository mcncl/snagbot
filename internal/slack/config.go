package slack

import (
	"fmt"
	"log"

	"github.com/mcncl/snagbot/pkg/models"
)

// In-memory storage for channel configurations
// In a production environment, this would be replaced with a persistent database
var channelConfigs = make(map[string]*models.ChannelConfig)

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
