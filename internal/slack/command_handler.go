package slack

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/slack-go/slack"
)

// CommandHandler creates a handler for Slack slash commands
func CommandHandler(cfg *config.Config) http.HandlerFunc {
	// Create a single instance of the config store for all requests
	configStore := NewInMemoryConfigStoreWithConfig(cfg)

	return func(w http.ResponseWriter, r *http.Request) {
		// Verify the request is coming from Slack
		sv, err := slack.NewSecretsVerifier(r.Header, cfg.SlackSigningSecret)
		if err != nil {
			log.Printf("Error creating secrets verifier: %v", err)
			http.Error(w, "Error verifying request", http.StatusBadRequest)
			return
		}

		// Parse the form to get command data
		err = r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing request", http.StatusBadRequest)
			return
		}

		// Add the form values to the signature verification
		sv.Write([]byte(r.Form.Encode()))
		if err := sv.Ensure(); err != nil {
			log.Printf("Error verifying signature: %v", err)
			http.Error(w, "Invalid request signature", http.StatusUnauthorized)
			return
		}

		// Extract command data
		command := r.Form.Get("command")
		text := r.Form.Get("text")
		channelID := r.Form.Get("channel_id")
		userID := r.Form.Get("user_id")

		// Log the command
		log.Printf("Received command %s with text '%s' from user %s in channel %s",
			command, text, userID, channelID)

		// Only process /snagbot command
		if command != "/snagbot" {
			log.Printf("Received unknown command: %s", command)
			http.Error(w, "Unknown command", http.StatusBadRequest)
			return
		}

		// Handle different subcommands
		response := ""
		if strings.TrimSpace(strings.ToLower(text)) == "reset" {
			response = handleResetCommand(configStore, channelID)
		} else if strings.TrimSpace(strings.ToLower(text)) == "status" {
			response = handleStatusCommand(configStore, channelID)
		} else {
			response = handleConfigCommand(configStore, text, channelID)
		}

		// Return the response immediately with 200 OK
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"response_type": "ephemeral", "text": %q}`, response)))
	}
}

// handleConfigCommand processes the command text and updates the channel configuration
func handleConfigCommand(store ChannelConfigStore, text, channelID string) string {
	// Parse the command
	result, err := ParseConfigCommand(text)
	if err != nil {
		log.Printf("Error parsing command: %v", err)
		return FormatCommandErrorResponse(err)
	}

	// Update the channel configuration
	err = store.UpdateConfig(channelID, result.ItemName, result.ItemPrice)
	if err != nil {
		log.Printf("Error updating channel config: %v", err)
		return fmt.Sprintf("Error updating configuration: %v", err)
	}

	// Return success message
	return FormatCommandResponse(result)
}

// Backward compatibility function for existing code
func handleConfigCommandLegacy(text, channelID string) string {
	return handleConfigCommand(globalConfigStore, text, channelID)
}

// handleResetCommand resets a channel's configuration to the default
func handleResetCommand(store ChannelConfigStore, channelID string) string {
	// Reset the config
	err := store.ResetConfig(channelID)
	if err != nil {
		log.Printf("Error resetting channel config: %v", err)
		return fmt.Sprintf("Error resetting configuration: %v", err)
	}

	// Get default config after reset
	defaultConfig := store.GetConfig(channelID)

	return fmt.Sprintf("Configuration has been reset! Now using the default item: %s (at $%.2f each).",
		defaultConfig.ItemName, defaultConfig.ItemPrice)
}

// handleStatusCommand returns the current configuration for a channel
func handleStatusCommand(store ChannelConfigStore, channelID string) string {
	config := store.GetConfig(channelID)

	// Check if this is a custom or default config
	if checker, ok := store.(ConfigExistsChecker); ok && !checker.ConfigExists(channelID) {
		return fmt.Sprintf("This channel is using the default configuration: %s (at $%.2f each).",
			config.ItemName, config.ItemPrice)
	}

	return fmt.Sprintf("Current configuration: %s (at $%.2f each).",
		config.ItemName, config.ItemPrice)
}

// handleConfigCommandWithService is the implementation for the service pattern
func handleConfigCommandWithService(text, channelID string, service *SlackService) string {
	// Parse the command
	result, err := ParseConfigCommand(text)
	if err != nil {
		log.Printf("Error parsing command: %v", err)
		return FormatCommandErrorResponse(err)
	}

	// Update the channel configuration using the service
	err = service.configStore.UpdateConfig(channelID, result.ItemName, result.ItemPrice)
	if err != nil {
		log.Printf("Error updating channel config: %v", err)
		return fmt.Sprintf("Error updating configuration: %v", err)
	}

	// Return success message
	return FormatCommandResponse(result)
}
