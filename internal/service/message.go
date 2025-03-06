package service

import (
	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/slack"
	"github.com/mcncl/snagbot/pkg/models"
	"github.com/slack-go/slack/slackevents"
)

// SlackService represents the main service for handling Slack interactions
type SlackService struct {
	ChannelConfigStore slack.ChannelConfigStore
	SlackAPI           slack.SlackAPI
}

// NewSlackService creates a new SlackService
func NewSlackService(store slack.ChannelConfigStore, api slack.SlackAPI) *SlackService {
	return &SlackService{
		ChannelConfigStore: store,
		SlackAPI:           api,
	}
}

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

// HandleMessageEvent processes a Slack message event using the service
func (s *SlackService) HandleMessageEvent(ev *slackevents.MessageEvent) error {
	// Skip bot messages to prevent loops
	if ev.BotID != "" || ev.SubType == "bot_message" {
		return nil
	}

	// Skip message changes/edits for now (can be implemented later)
	if ev.SubType == "message_changed" {
		return nil
	}

	// Get channel configuration
	config := s.ChannelConfigStore.GetConfig(ev.Channel)

	// Process the message
	message := ProcessMessageWithConfig(ev.Text, config)

	// If no message was generated, no dollar values were found
	if message == "" {
		return nil
	}

	// Send response as a thread
	response := slack.SlackResponse{
		ChannelID: ev.Channel,
		Text:      message,
		ThreadTS:  ev.TimeStamp, // Reply in thread
	}

	return s.SlackAPI.PostMessage(response)
}
