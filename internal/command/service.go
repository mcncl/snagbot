package command

import (
	"github.com/mcncl/snagbot/internal/slack"
)

// Global store for backward compatibility
var globalConfigStore slack.ChannelConfigStore

// Set the global store for tests
func SetGlobalStore(store slack.ChannelConfigStore) {
	globalConfigStore = store
}

// handleConfigCommandWithService processes the command text using a slack service
func handleConfigCommandWithService(text, channelID string, service *slack.SlackService) string {
	// Parse the command
	result, err := ParseConfigCommand(text)
	if err != nil {
		return FormatCommandErrorResponse(err)
	}

	// Update the channel configuration
	err = service.ConfigStore.UpdateConfig(channelID, result.ItemName, result.ItemPrice)
	if err != nil {
		return "Error updating configuration: " + err.Error()
	}

	// Return success message
	return FormatCommandResponse(result)
}

// handleConfigCommandLegacy is a backward compatibility function
func handleConfigCommandLegacy(text, channelID string) string {
	// Use the global store
	return handleConfigCommand(globalConfigStore, text, channelID)
}
