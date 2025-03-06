package slack

import (
	"github.com/slack-go/slack/slackevents"
)

// MockMessageEvent represents a mock Slack message event for testing
type MockMessageEvent struct {
	ChannelID string
	UserID    string
	Text      string
	BotID     string
	TS        string
	SubType   string
}

// ToSlackEvent converts a MockMessageEvent to a slackevents.MessageEvent
func (m *MockMessageEvent) ToSlackEvent() *slackevents.MessageEvent {
	return &slackevents.MessageEvent{
		Channel:   m.ChannelID,
		User:      m.UserID,
		Text:      m.Text,
		BotID:     m.BotID,
		TimeStamp: m.TS,
		SubType:   m.SubType,
	}
}

// HandleMockMessageEvent processes a mock message event for testing
func HandleMockMessageEvent(mockEvent *MockMessageEvent) error {
	// Convert mock event to Slack event
	slackEvent := mockEvent.ToSlackEvent()

	// Get static instances for testing
	configStore := globalConfigStore
	if configStore == nil {
		configStore = NewInMemoryConfigStore()
	}

	// Use the mock API
	mockAPI := globalMockAPI

	// Send the message to the mock API to be retrieved by tests
	err := ProcessMessageEvent(slackEvent, configStore, mockAPI)

	return err
}

// For integration tests, we need a global mock API that can be accessed
var globalMockAPI = NewMockSlackAPI()

// GetGlobalMockAPI returns the global mock API instance
func GetGlobalMockAPI() *MockSlackAPI {
	return globalMockAPI
}

// SetGlobalMockAPI sets the global mock API instance
func SetGlobalMockAPI(api *MockSlackAPI) {
	globalMockAPI = api
}

// ResetGlobalMockAPI clears the sent messages in the global mock API
func ResetGlobalMockAPI() {
	globalMockAPI.SentMessages = nil
}

// SetGlobalConfigStore sets the global config store for testing
func SetGlobalConfigStore(store ChannelConfigStore) {
	globalConfigStore = store
}
