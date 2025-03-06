package integration

import (
	"testing"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/slack"
	"github.com/stretchr/testify/assert"
)

// TestChannelConfigWithMessageHandling tests the integration between channel configuration
// and message handling logic
func TestChannelConfigWithMessageHandling(t *testing.T) {
	// Create the configuration
	cfg := &config.Config{
		DefaultItemName:  "Bunnings snags",
		DefaultItemPrice: 3.50,
	}

	// Create the config store
	configStore := slack.NewInMemoryConfigStore()

	// Create the mock API
	mockAPI := slack.NewMockSlackAPI()

	// Create the service
	service := slack.NewSlackServiceWithDependencies(configStore, mockAPI, cfg)

	// Define test cases
	tests := []struct {
		name            string
		channelID       string
		setupFunc       func()
		messageText     string
		expectedMessage string
		shouldRespond   bool
	}{
		{
			name:            "Default config with dollar value",
			channelID:       "C12345",
			setupFunc:       nil, // No special setup needed for default
			messageText:     "This costs $35",
			expectedMessage: "That's nearly 10 Bunnings snags!",
			shouldRespond:   true,
		},
		{
			name:      "Custom config with dollar value",
			channelID: "C67890",
			setupFunc: func() {
				configStore.UpdateConfig("C67890", "coffee", 5.00)
			},
			messageText:     "This costs $35",
			expectedMessage: "That's nearly 7 coffees!",
			shouldRespond:   true,
		},
		{
			name:      "Reset config and test with dollar value",
			channelID: "C13579",
			setupFunc: func() {
				// First set a custom config
				configStore.UpdateConfig("C13579", "donut", 2.00)
				// Then reset it to default
				configStore.ResetConfig("C13579")
			},
			messageText:     "This costs $35",
			expectedMessage: "That's nearly 10 Bunnings snags!", // Back to default
			shouldRespond:   true,
		},
		{
			name:      "Message with multiple dollar values and custom config",
			channelID: "C24680",
			setupFunc: func() {
				configStore.UpdateConfig("C24680", "cookie", 1.50)
			},
			messageText:     "This costs $20 and that costs $15",
			expectedMessage: "That's nearly 24 cookies!",
			shouldRespond:   true,
		},
		{
			name:          "Message with no dollar values",
			channelID:     "C12345",
			setupFunc:     nil,
			messageText:   "This message has no dollar values",
			shouldRespond: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset the mock API for each test
			mockAPI.SentMessages = nil

			// Run setup if provided
			if test.setupFunc != nil {
				test.setupFunc()
			}

			// Create a mock Slack message event
			event := &slack.MockMessageEvent{
				ChannelID: test.channelID,
				UserID:    "U12345",
				Text:      test.messageText,
				BotID:     "", // Not a bot message
				TS:        "1234567890.123456",
				SubType:   "",
			}

			// Handle the message event
			err := slack.HandleMockMessageEvent(event)
			assert.NoError(t, err)

			// Check if we should have a response
			if test.shouldRespond {
				assert.Len(t, mockAPI.SentMessages, 1)
				assert.Equal(t, test.expectedMessage, mockAPI.SentMessages[0].Text)
				assert.Equal(t, test.channelID, mockAPI.SentMessages[0].ChannelID)
			} else {
				assert.Len(t, mockAPI.SentMessages, 0)
			}
		})
	}
}
