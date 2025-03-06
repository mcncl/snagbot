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

	// For very small amounts that don't reach 1 item
	if total < config.ItemPrice {
		// Use the standard "zero" response
		message := calculator.FormatResponse(0, config.ItemName, true)

		return api.PostMessage(SlackResponse{
			ChannelID: ev.Channel,
			Text:      message,
			ThreadTS:  ev.TimeStamp,
		})
	}

	// Check if the division is exact (to decide whether to use "nearly")
	isExactDivision := (total / config.ItemPrice) == float64(int(total/config.ItemPrice))

	// Calculate number of items
	count := calculator.CalculateItemCount(total, config.ItemPrice)

	// Format response message
	message := calculator.FormatResponse(count, config.ItemName, isExactDivision)

	// Send response as a thread
	response := SlackResponse{
		ChannelID: ev.Channel,
		Text:      message,
		ThreadTS:  ev.TimeStamp,
	}

	return api.PostMessage(response)
}
