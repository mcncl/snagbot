package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleMessageEvent(t *testing.T) {
	tests := []struct {
		name            string
		messageText     string
		channelID       string
		itemName        string
		itemPrice       float64
		expectedMessage string
		shouldRespond   bool
	}{
		{
			name:            "No dollar values",
			messageText:     "This message has no dollar values",
			channelID:       "C12345",
			itemName:        "Bunnings snags",
			itemPrice:       3.50,
			expectedMessage: "",
			shouldRespond:   false,
		},
		{
			name:            "Single dollar value default config",
			messageText:     "This would cost $35",
			channelID:       "C12345",
			itemName:        "Bunnings snags",
			itemPrice:       3.50,
			expectedMessage: "That's nearly 10 Bunnings snags!",
			shouldRespond:   true,
		},
		{
			name:            "Multiple dollar values",
			messageText:     "This costs $35 and that costs $15",
			channelID:       "C12345",
			itemName:        "Bunnings snags",
			itemPrice:       3.50,
			expectedMessage: "That's nearly 15 Bunnings snags!",
			shouldRespond:   true,
		},
		{
			name:            "Custom item singular",
			messageText:     "This costs $10",
			channelID:       "C67890",
			itemName:        "coffee",
			itemPrice:       5.00,
			expectedMessage: "That's nearly 2 coffees!",
			shouldRespond:   true,
		},
		{
			name:            "Custom item already plural",
			messageText:     "This costs $25",
			channelID:       "C67890",
			itemName:        "coffees",
			itemPrice:       5.00,
			expectedMessage: "That's nearly 5 coffees!",
			shouldRespond:   true,
		},
		{
			name:            "Fractional amount",
			messageText:     "This costs $3.25",
			channelID:       "C12345",
			itemName:        "Bunnings snags",
			itemPrice:       3.50,
			expectedMessage: "That's nearly 1 Bunnings snag!",
			shouldRespond:   true,
		},
		{
			name:            "Multiple fractional amounts",
			messageText:     "This costs $3.25 and that costs $7.75",
			channelID:       "C12345",
			itemName:        "Bunnings snags",
			itemPrice:       3.50,
			expectedMessage: "That's nearly 4 Bunnings snags!",
			shouldRespond:   true,
		},
		{
			name:            "Zero dollar amount",
			messageText:     "This costs $0",
			channelID:       "C12345",
			itemName:        "Bunnings snag",
			itemPrice:       3.50,
			expectedMessage: "That wouldn't even buy a single Bunnings snag!",
			shouldRespond:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup mock store and API
			configStore := NewInMemoryConfigStore()
			mockAPI := NewMockSlackAPI()

			// Configure the channel config if needed
			if test.itemName != "Bunnings snags" || test.itemPrice != 3.50 {
				err := configStore.UpdateConfig(test.channelID, test.itemName, test.itemPrice)
				assert.NoError(t, err)
			}

			// Create handler
			handler := NewMessageEventHandler(configStore, mockAPI)

			// Create event
			event := SlackMessageEvent{
				ChannelID: test.channelID,
				UserID:    "U12345",
				Text:      test.messageText,
				Timestamp: "1623436858.000100",
			}

			// Handle event
			err := handler.HandleMessageEvent(event)
			assert.NoError(t, err)

			// Check if we should have a response
			if test.shouldRespond {
				assert.Len(t, mockAPI.SentMessages, 1)
				assert.Equal(t, test.expectedMessage, mockAPI.SentMessages[0].Text)
				assert.Equal(t, test.channelID, mockAPI.SentMessages[0].ChannelID)
				assert.Equal(t, event.Timestamp, mockAPI.SentMessages[0].ThreadTS)
			} else {
				assert.Len(t, mockAPI.SentMessages, 0)
			}
		})
	}
}

func TestMultipleChannelConfigurations(t *testing.T) {
	// Setup
	configStore := NewInMemoryConfigStore()
	mockAPI := NewMockSlackAPI()
	handler := NewMessageEventHandler(configStore, mockAPI)

	// Configure multiple channels
	channel1 := "C12345"
	channel2 := "C67890"

	// Update channel 2 with custom config
	err := configStore.UpdateConfig(channel2, "coffee", 5.00)
	assert.NoError(t, err)

	// Test message in channel 1 (default config)
	event1 := SlackMessageEvent{
		ChannelID: channel1,
		UserID:    "U12345",
		Text:      "This would cost $35",
		Timestamp: "1623436858.000100",
	}

	err = handler.HandleMessageEvent(event1)
	assert.NoError(t, err)
	assert.Len(t, mockAPI.SentMessages, 1)
	assert.Equal(t, "That's nearly 10 Bunnings snags!", mockAPI.SentMessages[0].Text)

	// Reset mock API for next test
	mockAPI.SentMessages = nil

	// Test message in channel 2 (custom config)
	event2 := SlackMessageEvent{
		ChannelID: channel2,
		UserID:    "U12345",
		Text:      "This would cost $35",
		Timestamp: "1623436859.000100",
	}

	err = handler.HandleMessageEvent(event2)
	assert.NoError(t, err)
	assert.Len(t, mockAPI.SentMessages, 1)
	assert.Equal(t, "That's nearly 7 coffees!", mockAPI.SentMessages[0].Text)
}

func TestInMemoryConfigStore(t *testing.T) {
	store := NewInMemoryConfigStore()

	// Test getting a default config
	channelID := "C12345"
	config := store.GetConfig(channelID)
	assert.Equal(t, channelID, config.ChannelID)
	assert.Equal(t, "Bunnings snags", config.ItemName)
	assert.Equal(t, 3.50, config.ItemPrice)

	// Test updating a config
	err := store.UpdateConfig(channelID, "coffee", 5.00)
	assert.NoError(t, err)

	// Verify the update
	config = store.GetConfig(channelID)
	assert.Equal(t, "coffee", config.ItemName)
	assert.Equal(t, 5.00, config.ItemPrice)

	// Test invalid price
	err = store.UpdateConfig(channelID, "invalid", 0)
	assert.Error(t, err)

	// Config should remain unchanged
	config = store.GetConfig(channelID)
	assert.Equal(t, "coffee", config.ItemName)
	assert.Equal(t, 5.00, config.ItemPrice)
}

func TestMockSlackAPI(t *testing.T) {
	mockAPI := NewMockSlackAPI()

	// Test posting a message
	response := SlackResponse{
		ChannelID: "C12345",
		Text:      "Test message",
		ThreadTS:  "1623436858.000100",
	}

	err := mockAPI.PostMessage(response)
	assert.NoError(t, err)
	assert.Len(t, mockAPI.SentMessages, 1)
	assert.Equal(t, response, mockAPI.SentMessages[0])

	// Test posting another message
	response2 := SlackResponse{
		ChannelID: "C67890",
		Text:      "Another test message",
		ThreadTS:  "1623436859.000100",
	}

	err = mockAPI.PostMessage(response2)
	assert.NoError(t, err)
	assert.Len(t, mockAPI.SentMessages, 2)
	assert.Equal(t, response2, mockAPI.SentMessages[1])
}
