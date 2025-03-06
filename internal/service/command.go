package service

import (
	"fmt"

	"github.com/mcncl/snagbot/internal/command"
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

// HandleConfigCommand processes a configuration command
func (s *CommandService) HandleConfigCommand(text, channelID string) string {
	// Parse the command
	result, err := command.ParseConfigCommand(text)
	if err != nil {
		return command.FormatCommandErrorResponse(err)
	}

	// Update the channel configuration
	err = s.ConfigStore.UpdateConfig(channelID, result.ItemName, result.ItemPrice)
	if err != nil {
		return "Error updating configuration: " + err.Error()
	}

	// Return success message
	return command.FormatCommandResponse(result)
}

// HandleResetCommand resets a channel's configuration
func (s *CommandService) HandleResetCommand(channelID string) string {
	// Reset the config
	err := s.ConfigStore.ResetConfig(channelID)
	if err != nil {
		return "Error resetting configuration: " + err.Error()
	}

	// Get default config after reset
	defaultConfig := s.ConfigStore.GetConfig(channelID)

	return "Configuration has been reset! Now using the default item: " +
		defaultConfig.ItemName + " (at $" +
		FormatPrice(defaultConfig.ItemPrice) + " each)."
}

// HandleStatusCommand returns the current configuration for a channel
func (s *CommandService) HandleStatusCommand(channelID string) string {
	config := s.ConfigStore.GetConfig(channelID)

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
func FormatPrice(price float64) string {
	return fmt.Sprintf("%.2f", price)
}
