package slack

import (
	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/errors"
	"github.com/mcncl/snagbot/internal/logging"
	"github.com/slack-go/slack/slackevents"
)

// ProcessMessageEvent handles a message event from Slack
func ProcessMessageEvent(ev *slackevents.MessageEvent, configStore ChannelConfigStore, api SlackAPI) error {
	// Skip processing if the event is nil
	if ev == nil {
		return errors.New(errors.ErrInvalidRequest, "nil message event")
	}

	// Skip bot messages to prevent loops
	if ev.BotID != "" || ev.SubType == "bot_message" {
		logging.Debug("Skipping bot message from BotID: %s", ev.BotID)
		return nil
	}

	// Skip message changes/edits for now (can be implemented later)
	if ev.SubType == "message_changed" {
		logging.Debug("Skipping message_changed event")
		return nil
	}

	// Get channel configuration
	config, err := configStore.GetConfig(ev.Channel)
	if err != nil {
		appErr := errors.Wrap(err, "Failed to get channel configuration")
		logging.Error("Config retrieval error: %v", appErr)
		HandleErrorWithResponse(appErr, ev, api)
		return appErr
	}

	logging.Debug("Processing message: %s", ev.Text)
	logging.Debug("Using channel config: item=%s, price=%.2f", config.ItemName, config.ItemPrice)

	// Extract dollar values from the message
	dollarValues, err := calculator.ExtractDollarValues(ev.Text)
	if err != nil {
		appErr := errors.Wrap(err, "Failed to extract dollar values")
		logging.Error("Dollar value extraction error: %v", appErr)
		return appErr
	}

	if len(dollarValues) == 0 {
		// No dollar values found, nothing to do
		logging.Debug("No dollar values found in message, skipping")
		return nil
	}

	logging.Info("Found %d dollar values in message", len(dollarValues))

	// Calculate total dollar amount
	total, err := calculator.SumDollarValues(dollarValues)
	if err != nil {
		appErr := errors.Wrap(err, "Failed to sum dollar values")
		logging.Error("Dollar value summation error: %v", appErr)
		HandleErrorWithResponse(appErr, ev, api)
		return appErr
	}

	logging.Debug("Total dollar amount: $%.2f", total)

	// For very small amounts that don't reach 1 item
	if total < config.ItemPrice {
		// Use the standard "zero" response
		message := calculator.FormatResponse(0, config.ItemName, true)
		logging.Debug("Amount too small for one item, using zero response: %s", message)

		return api.PostMessage(SlackResponse{
			ChannelID: ev.Channel,
			Text:      message,
			ThreadTS:  ev.TimeStamp,
		})
	}

	// Check if the division is exact (to decide whether to use "nearly")
	isExactDivision := calculator.IsExactDivision(total, config.ItemPrice)

	// Calculate number of items
	count, err := calculator.CalculateItemCount(total, config.ItemPrice)
	if err != nil {
		appErr := errors.Wrap(err, "Failed to calculate item count")
		logging.Error("Item count calculation error: %v", appErr)
		HandleErrorWithResponse(appErr, ev, api)
		return appErr
	}

	// Format response message
	message := calculator.FormatResponse(count, config.ItemName, isExactDivision)
	logging.Info("Responding with message: %s", message)

	// Send response as a thread
	response := SlackResponse{
		ChannelID: ev.Channel,
		Text:      message,
		ThreadTS:  ev.TimeStamp,
	}

	if err := api.PostMessage(response); err != nil {
		appErr := errors.Wrap(err, "Failed to post message to Slack")
		logging.Error("Slack API error: %v", appErr)
		return appErr
	}

	logging.Info("Successfully posted response to channel %s", ev.Channel)
	return nil
}
