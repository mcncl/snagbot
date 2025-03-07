package slack

import (
	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/logging"
	"github.com/slack-go/slack/slackevents"
)

// SlackService represents the main service for handling Slack interactions
type SlackService struct {
	ConfigStore ChannelConfigStore
	SlackAPI    SlackAPI
	Config      *config.Config
}

// NewSlackService creates a new SlackService
func NewSlackService(cfg *config.Config) *SlackService {
	return &SlackService{
		ConfigStore: NewInMemoryConfigStoreWithConfig(cfg),
		SlackAPI:    NewRealSlackAPI(cfg.SlackBotToken),
		Config:      cfg,
	}
}

// ProcessMessageEvent processes a Slack message event
func (s *SlackService) ProcessMessageEvent(ev *slackevents.MessageEvent) error {
	// Skip bot messages to prevent loops
	if ev.BotID != "" || ev.SubType == "bot_message" {
		return nil
	}

	// Skip message changes/edits for now (can be implemented later)
	if ev.SubType == "message_changed" {
		return nil
	}

	// Get channel configuration
	config, err := s.ConfigStore.GetConfig(ev.Channel)
	if err != nil {
		logging.Error("Failed to get channel configuration: %v", err)
		return err
	}

	// Process the message using the shared utility function
	message := calculator.ProcessMessageWithConfig(ev.Text, config)

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

	return s.SlackAPI.PostMessage(response)
}
