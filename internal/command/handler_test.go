package command

import (
	"testing"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/slack"
	"github.com/stretchr/testify/assert"
)

// TestHandleConfigCommand tests the command handler logic
func TestHandleConfigCommand(t *testing.T) {
	// Store the original channelConfigs map
	originalStore := globalConfigStore
	// Restore it after the test
	defer func() { globalConfigStore = originalStore }()

	tests := []struct {
		name              string
		commandText       string
		channelID         string
		expectedSuccess   bool
		expectedItemName  string
		expectedItemPrice float64
	}{
		{
			name:              "Valid command with coffee",
			commandText:       "item \"coffee\" price 5.00",
			channelID:         "C12345",
			expectedSuccess:   true,
			expectedItemName:  "coffee",
			expectedItemPrice: 5.00,
		},
		{
			name:              "Valid command with single word",
			commandText:       "item donut price 2.50",
			channelID:         "C67890",
			expectedSuccess:   true,
			expectedItemName:  "donut",
			expectedItemPrice: 2.50,
		},
		{
			name:              "Valid command with multi-word item",
			commandText:       "item \"croissant and coffee\" price 7.50",
			channelID:         "C12345",
			expectedSuccess:   true,
			expectedItemName:  "croissant and coffee",
			expectedItemPrice: 7.50,
		},
		{
			name:            "Invalid command",
			commandText:     "invalid command",
			channelID:       "C12345",
			expectedSuccess: false,
		},
		{
			name:            "Missing price",
			commandText:     "item coffee",
			channelID:       "C12345",
			expectedSuccess: false,
		},
		{
			name:            "Invalid price",
			commandText:     "item coffee price abc",
			channelID:       "C12345",
			expectedSuccess: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a fresh config store for each test
			configStore := slack.NewInMemoryConfigStore()
			globalConfigStore = configStore

			// Process the command
			response := handleConfigCommand(configStore, test.commandText, test.channelID)

			// Check if response indicates success or failure
			if test.expectedSuccess {
				assert.Contains(t, response, "Configuration updated!")
				assert.Contains(t, response, test.expectedItemName)
				assert.Contains(t, response, "at $")

				// Verify the channel config was updated correctly
				config := configStore.GetConfig(test.channelID)
				assert.Equal(t, test.expectedItemName, config.ItemName)
				assert.Equal(t, test.expectedItemPrice, config.ItemPrice)
			} else {
				assert.NotContains(t, response, "Configuration updated!")
				assert.Contains(t, response, "Usage example:")
			}
		})
	}
}

// TestHandleConfigCommandWithService tests the service-based command handler
func TestHandleConfigCommandWithService(t *testing.T) {
	tests := []struct {
		name              string
		commandText       string
		channelID         string
		expectedSuccess   bool
		expectedItemName  string
		expectedItemPrice float64
	}{
		{
			name:              "Valid command with service",
			commandText:       "item \"coffee\" price 5.00",
			channelID:         "C12345",
			expectedSuccess:   true,
			expectedItemName:  "coffee",
			expectedItemPrice: 5.00,
		},
		{
			name:            "Invalid command with service",
			commandText:     "invalid command",
			channelID:       "C12345",
			expectedSuccess: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup test dependencies
			configStore := slack.NewInMemoryConfigStore()
			mockAPI := slack.NewMockSlackAPI()
			cfg := &config.Config{
				DefaultItemName:  "Bunnings snags",
				DefaultItemPrice: 3.50,
			}
			service := slack.NewSlackServiceWithDependencies(configStore, mockAPI, cfg)

			// Process the command
			response := handleConfigCommandWithService(test.commandText, test.channelID, service)

			// Check if response indicates success or failure
			if test.expectedSuccess {
				assert.Contains(t, response, "Configuration updated!")

				// Verify the channel config was updated correctly
				config := configStore.GetConfig(test.channelID)
				assert.Equal(t, test.expectedItemName, config.ItemName)
				assert.Equal(t, test.expectedItemPrice, config.ItemPrice)
			} else {
				assert.NotContains(t, response, "Configuration updated!")
				assert.Contains(t, response, "Usage example:")
			}
		})
	}
}

// TestLegacyCompatibility tests the backward compatibility functions
func TestLegacyCompatibility(t *testing.T) {
	// Store the original channelConfigs map
	originalStore := globalConfigStore
	// Restore it after the test
	defer func() { globalConfigStore = originalStore }()

	// Create a fresh config store
	configStore := slack.NewInMemoryConfigStore()
	SetGlobalStore(configStore)

	// Test the legacy function
	channelID := "C12345"
	commandText := "item \"coffee\" price 5.00"

	response := handleConfigCommandLegacy(commandText, channelID)

	// Verify the response
	assert.Contains(t, response, "Configuration updated!")
	assert.Contains(t, response, "coffee")

	// Verify the config was updated
	config := configStore.GetConfig(channelID)
	assert.Equal(t, "coffee", config.ItemName)
	assert.Equal(t, 5.00, config.ItemPrice)
}
