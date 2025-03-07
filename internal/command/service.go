package command

import (
	"fmt"

	"github.com/mcncl/snagbot/internal/errors"
	"github.com/mcncl/snagbot/internal/logging"
	"github.com/mcncl/snagbot/internal/slack"
)

// CommandService handles Slack slash commands
type CommandService struct {
	ConfigStore slack.ChannelConfigStore
}

// NewCommandService creates a new CommandService
func NewCommandService(store slack.ChannelConfigStore) *CommandService {
	return &CommandService{
		ConfigStore: store,
	}
}

// handleConfigCommand processes the command text and updates the channel configuration
// This function is for backward compatibility
func handleConfigCommand(store slack.ChannelConfigStore, text, channelID string) string {
	// Parse the command
	result, err := ParseConfigCommand(text)
	if err != nil {
		logging.Error("Error parsing command: %v", err)
		return FormatCommandErrorResponse(err)
	}

	// Update the channel configuration
	err = store.UpdateConfig(channelID, result.ItemName, result.ItemPrice)
	if err != nil {
		logging.Error("Error updating channel config: %v", err)
		return fmt.Sprintf("Error updating configuration: %v", err)
	}

	// Return success message
	return FormatCommandResponse(result)
}

// HandleConfigCommand processes a configuration command
func (s *CommandService) HandleConfigCommand(text, channelID string) string {
	// Parse the command
	result, err := ParseConfigCommand(text)
	if err != nil {
		appErr := errors.Wrap(err, "Failed to parse command")
		logging.Error("Command parsing error: %v", appErr)
		return FormatCommandErrorResponse(err)
	}

	// Update the channel configuration
	err = s.ConfigStore.UpdateConfig(channelID, result.ItemName, result.ItemPrice)
	if err != nil {
		appErr := errors.Wrap(err, "Failed to update configuration")
		logging.Error("Config update error: %v", appErr)
		return "Error updating configuration: " + errors.UserFriendlyError(appErr)
	}

	// Return success message
	return FormatCommandResponse(result)
}

// HandleResetCommand resets a channel's configuration
func (s *CommandService) HandleResetCommand(channelID string) string {
	// Reset the config
	err := s.ConfigStore.ResetConfig(channelID)
	if err != nil {
		appErr := errors.Wrap(err, "Failed to reset configuration")
		logging.Error("Config reset error: %v", appErr)
		return "Error resetting configuration: " + errors.UserFriendlyError(appErr)
	}

	// Get default config after reset
	config, err := s.ConfigStore.GetConfig(channelID)
	if err != nil {
		logging.Error("Error getting default config after reset: %v", err)
		return "Configuration has been reset, but unable to retrieve default settings."
	}

	return "Configuration has been reset! Now using the default item: " +
		config.ItemName + " (at $" +
		FormatPrice(config.ItemPrice) + " each)."
}

// HandleStatusCommand returns the current configuration for a channel
func (s *CommandService) HandleStatusCommand(channelID string) string {
	config, err := s.ConfigStore.GetConfig(channelID)
	if err != nil {
		appErr := errors.Wrap(err, "Failed to get configuration")
		logging.Error("Config retrieval error: %v", appErr)
		return "Error retrieving configuration: " + errors.UserFriendlyError(appErr)
	}

	// Check if this is a custom or default config
	var statusPrefix string

	if checker, ok := s.ConfigStore.(slack.ConfigExistsChecker); ok && !checker.ConfigExists(channelID) {
		statusPrefix = "This channel is using the default configuration: "
	} else {
		statusPrefix = "Current configuration: "
	}

	return statusPrefix + config.ItemName + " (at $" +
		FormatPrice(config.ItemPrice) + " each)."
}

// FormatPrice formats a price with 2 decimal places
// This is a widely used utility function that could be moved to a common package
func FormatPrice(price float64) string {
	return fmt.Sprintf("%.2f", price)
}
