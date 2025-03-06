package slack

import (
	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/slack-go/slack/slackevents"
)

// ProcessMessageEvent handles a message event from Slack
func ProcessMessageEvent(ev *slackevents.MessageEvent, configStore ChannelConfigStore, api SlackAPI) error {
	// Skip bot messages to prevent loops
	if ev.BotID != "" || ev.SubType == "bot_message" {
		return nil
	}

	// Skip message changes/edits for now (can be implemented later)
	if ev.SubType == "message_changed" {
		return nil
	}

	// Get channel configuration
	config := configStore.GetConfig(ev.Channel)

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

	return api.PostMessage(response)
}
