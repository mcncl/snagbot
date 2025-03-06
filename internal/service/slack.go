package service

import (
	"fmt"
	"log"

	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/pkg/models"
)

// SlackMessageEvent represents a simplified Slack message event
type SlackMessageEvent struct {
	ChannelID string
	UserID    string
	Text      string
	ThreadTS  string
	Timestamp string
}

// SlackResponse represents a response to be sent back to Slack
type SlackResponse struct {
	ChannelID string
	Text      string
	ThreadTS  string
}

// ChannelConfigStore represents an interface for storing and retrieving channel configurations
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

// SlackAPI represents an interface for interacting with the Slack API
type SlackAPI interface {
	PostMessage(response SlackResponse) error
}

// MockSlackAPI provides a mock implementation for testing
type MockSlackAPI struct {
	SentMessages []SlackResponse
}

// NewMockSlackAPI creates a new mock Slack API
func NewMockSlackAPI() *MockSlackAPI {
	return &MockSlackAPI{
		SentMessages: make([]SlackResponse, 0),
	}
}

// PostMessage simulates posting a message to Slack
func (m *MockSlackAPI) PostMessage(response SlackResponse) error {
	m.SentMessages = append(m.SentMessages, response)
	log.Printf("Message sent to channel %s: %s", response.ChannelID, response.Text)
	return nil
}

// MessageEventHandler handles Slack message events
type MessageEventHandler struct {
	configStore ChannelConfigStore
	slackAPI    SlackAPI
}

// NewMessageEventHandler creates a new message event handler
func NewMessageEventHandler(configStore ChannelConfigStore, slackAPI SlackAPI) *MessageEventHandler {
	return &MessageEventHandler{
		configStore: configStore,
		slackAPI:    slackAPI,
	}
}

// HandleMessageEvent processes a Slack message event
func (h *MessageEventHandler) HandleMessageEvent(event SlackMessageEvent) error {
	// Extract dollar values from the message
	dollarValues := calculator.ExtractDollarValues(event.Text)
	if len(dollarValues) == 0 {
		// No dollar values found, nothing to do
		return nil
	}

	// Get the channel configuration
	channelConfig := h.configStore.GetConfig(event.ChannelID)

	// Calculate total dollar amount
	total := calculator.SumDollarValues(dollarValues)

	// Calculate number of items
	count := calculator.CalculateItemCount(total, channelConfig.ItemPrice)

	// Format response message
	message := calculator.FormatResponse(count, channelConfig.ItemName)

	// Prepare the response
	response := SlackResponse{
		ChannelID: event.ChannelID,
		Text:      message,
		ThreadTS:  event.Timestamp, // Reply in thread using the message's timestamp
	}

	// Send the response
	return h.slackAPI.PostMessage(response)
}
