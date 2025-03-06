package service

import (
	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/slack"
	"github.com/mcncl/snagbot/pkg/models"
	"github.com/slack-go/slack/slackevents"
)

// ProcessMessageWithConfig processes a message with the given channel configuration
// and returns the formatted response string
func ProcessMessageWithConfig(text string, config *models.ChannelConfig) string {
	// Extract dollar values from the message
	dollarValues := calculator.ExtractDollarValues(text)
	if len(dollarValues) == 0 {
		// No dollar values found, nothing to do
		return ""
	}

	// Calculate total dollar amount
	total := calculator.SumDollarValues(dollarValues)

	// Calculate number of items
	count := calculator.CalculateItemCount(total, config.ItemPrice)

	// Format response message
	return calculator.FormatResponse(count, config.ItemName)
}

// ProcessMessageEvent processes a Slack message event using the configured
// channel configuration from the store
func ProcessMessageEvent(ev *slackevents.MessageEvent, store ChannelConfigStore, api SlackAPI) error {
	// Skip bot messages to prevent loops
	if ev.BotID != "" || ev.SubType == "bot_message" {
		return nil
	}

	// Skip message changes/edits for now (can be implemented later)
	if ev.SubType == "message_changed" {
		return nil
	}

	// Get channel configuration
	config := store.GetConfig(ev.Channel)

	// Process the message
	message := ProcessMessageWithConfig(ev.Text, config)

	// If no message was generated, no dollar values were found
	if message == "" {
		return nil
	}

	// Send response as a thread
	response := SlackResponse{
		ChannelID: ev.Channel,
		Text:      message,
		ThreadTS:  ev.TimeStamp, // Reply in thread
	}

	return api.PostMessage(response)
}

// MockMessageEvent represents a mock Slack message event for testing
type MockMessageEvent struct {
	ChannelID string
	UserID    string
	Text      string
	BotID     string
	SubType   string
	TS        string
}

// HandleMockMessageEvent handles a mock message event for testing
func (s *slack.SlackService) HandleMockMessageEvent(event *MockMessageEvent) error {
	// Skip bot messages to prevent loops
	if event.BotID != "" || event.SubType == "bot_message" {
		return nil
	}

	// Skip message changes/edits
	if event.SubType == "message_changed" {
		return nil
	}

	// Process the message using our calculator and configuration
	// Get configured item
	config := s.configStore.GetConfig(event.ChannelID)

	// Call the calculator to process the message
	response := ProcessMessageWithConfig(event.Text, config)

	// If no response, no dollar values were found
	if response == "" {
		return nil
	}

	// Send response
	slackResponse := SlackResponse{
		ChannelID: event.ChannelID,
		Text:      response,
		ThreadTS:  event.TS,
	}

	return s.slackAPI.PostMessage(slackResponse)
}

// Add method to the SlackService to process standard message events
func (s *SlackService) HandleMessageEvent(ev *slackevents.MessageEvent) error {
	return ProcessMessageEvent(ev, s.configStore, s.slackAPI)
}
