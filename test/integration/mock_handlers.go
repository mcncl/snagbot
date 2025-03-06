package integration

import (
	"github.com/mcncl/snagbot/internal/slack"
)

// HandleMockMessage is a helper function for integration tests
// that processes a mock message through the system
func HandleMockMessage(channelID, text string) (string, error) {
	// Create a mock event
	mockEvent := &slack.MockMessageEvent{
		ChannelID: channelID,
		UserID:    "U12345", // Standard test user
		Text:      text,
		TS:        "1234567890.123456", // Standard test timestamp
	}

	// Reset the global mock API
	slack.ResetGlobalMockAPI()

	// Handle the event
	err := slack.HandleMockMessageEvent(mockEvent)
	if err != nil {
		return "", err
	}

	// Get the mock API
	mockAPI := slack.GetGlobalMockAPI()

	// Return the response text or empty string if no response
	if len(mockAPI.SentMessages) > 0 {
		return mockAPI.SentMessages[0].Text, nil
	}

	return "", nil
}
