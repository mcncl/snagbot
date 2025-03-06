package slack

import (
	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/config"
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

// NewSlackServiceWithDependencies creates a new SlackService with custom dependencies
func NewSlackServiceWithDependencies(store ChannelConfigStore, api SlackAPI, cfg *config.Config) *SlackService {
	return &SlackService{
		ConfigStore: store,
		SlackAPI:    api,
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
	config := s.ConfigStore.GetConfig(ev.Channel)

	// Extract dollar values from the message
	dollarValues := calculator.ExtractDollarValues(ev.Text)
	if len(dollarValues) == 0 {
		// No dollar values found, nothing to do
		return nil
	}

	// Calculate total dollar amount
	total := calculator.SumDollarValues(dollarValues)

	// Calculate number of items
	count := calculator.CalculateItemCount(total, config.ItemPrice)

	// Format response message
	message := calculator.FormatResponse(count, config.ItemName)

	// Send response as a thread
	response := SlackResponse{
		ChannelID: ev.Channel,
		Text:      message,
		ThreadTS:  ev.TimeStamp, // Reply in thread
	}

	return s.SlackAPI.PostMessage(response)
}
