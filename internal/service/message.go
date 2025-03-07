package service

import (
	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/logging"
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
	// Use the shared implementation from calculator package
	return calculator.ProcessMessageWithConfig(text, config)
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
	config, err := s.ChannelConfigStore.GetConfig(ev.Channel)
	if err != nil {
		logging.Error("Config retrieval error: %v", err)
		return err
	}

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
