package command

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/errors"
	"github.com/mcncl/snagbot/internal/logging"
	slack "github.com/mcncl/snagbot/internal/slack"
	slackgo "github.com/slack-go/slack"
)

// Global store for backward compatibility
// This is needed for some tests but should be phased out in favor of dependency injection
// TODO: Remove this global variable and update tests to use dependency injection
var globalConfigStore slack.ChannelConfigStore

// SetGlobalStore sets the global store for tests
// DEPRECATED: Tests should be updated to use dependency injection instead
func SetGlobalStore(store slack.ChannelConfigStore) {
	globalConfigStore = store
}

// CommandHandler creates a handler for Slack slash commands
func CommandHandler(cfg *config.Config) http.HandlerFunc {
	// Create a single instance of the config store for all requests
	configStore := slack.NewInMemoryConfigStoreWithConfig(cfg)

	// Set the global store for backward compatibility
	globalConfigStore = configStore

	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests for commands
		if r.Method != http.MethodPost {
			logging.Warn("Method not allowed for command: %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check if Slack signing secret is configured
		if cfg.SlackSigningSecret == "" {
			logging.Error("Slack signing secret not configured")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Read and verify the request from Slack
		_, err := verifySlackRequest(r, cfg.SlackSigningSecret)
		if err != nil {
			appErr := errors.Wrap(err, "Failed to verify Slack request")
			logging.Error("Slack verification error: %v", appErr)
			http.Error(w, "Invalid request", http.StatusUnauthorized)
			return
		}

		// Parse the form to get command data
		err = r.ParseForm()
		if err != nil {
			appErr := errors.WrapAndLog(err, "Error parsing form")
			http.Error(w, appErr.Message, http.StatusBadRequest)
			return
		}

		// Extract command data
		command := r.Form.Get("command")
		text := r.Form.Get("text")
		channelID := r.Form.Get("channel_id")
		userID := r.Form.Get("user_id")
		userName := r.Form.Get("user_name")

		// Log the command
		logging.Info("Received command %s with text '%s' from user %s (%s) in channel %s",
			command, text, userName, userID, channelID)

		// Only process /snagbot command
		if command != "/snagbot" {
			logging.Warn("Received unknown command: %s", command)
			http.Error(w, "Unknown command", http.StatusBadRequest)
			return
		}

		// Handle different subcommands with error handling
		response := ""
		var cmdErr error

		trimmedText := strings.TrimSpace(strings.ToLower(text))
		switch {
		case trimmedText == "reset":
			response, cmdErr = safeHandleResetCommand(configStore, channelID)
		case trimmedText == "status" || trimmedText == "":
			// Empty command will show status too
			response, cmdErr = safeHandleStatusCommand(configStore, channelID)
		case strings.HasPrefix(trimmedText, "help"):
			response = handleHelpCommand()
		default:
			response, cmdErr = safeHandleConfigCommand(configStore, text, channelID)
		}

		// If there was an error, include a user-friendly error message
		if cmdErr != nil {
			logging.Error("Error handling command: %v", cmdErr)
			response = fmt.Sprintf("Error: %s\n\nTry `/snagbot help` for usage information.",
				errors.UserFriendlyError(cmdErr))
		}

		// Return the response immediately with 200 OK
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Format the response as JSON
		respJSON, err := json.Marshal(map[string]string{
			"response_type": "ephemeral",
			"text":          response,
		})

		if err != nil {
			logging.Error("Error marshalling response: %v", err)
			w.Write([]byte(`{"response_type": "ephemeral", "text": "Error generating response"}`))
			return
		}

		w.Write(respJSON)
	}
}

// verifySlackRequest verifies that a request is coming from Slack
// Returns the request body if verification succeeds, or an error if it fails
func verifySlackRequest(r *http.Request, signingSecret string) ([]byte, error) {
	// Verify that the request is coming from Slack
	sv, err := slackgo.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create secrets verifier")
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read request body")
	}

	// Replace the body for later use (since ReadAll depletes it)
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Add the body to the signature verification
	sv.Write(body)
	if err := sv.Ensure(); err != nil {
		return nil, errors.Wrap(err, "Failed to verify Slack signature")
	}

	return body, nil
}

// safeHandleConfigCommand processes the command text and updates the channel configuration
// with error handling
func safeHandleConfigCommand(store slack.ChannelConfigStore, text, channelID string) (string, error) {
	// Parse the command
	result, err := ParseConfigCommand(text)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse command")
	}

	// Update the channel configuration
	err = store.UpdateConfig(channelID, result.ItemName, result.ItemPrice)
	if err != nil {
		return "", errors.Wrap(err, "Failed to update configuration")
	}

	// Return success message
	return FormatCommandResponse(result), nil
}

// safeHandleResetCommand resets a channel's configuration to the default with error handling
func safeHandleResetCommand(store slack.ChannelConfigStore, channelID string) (string, error) {
	// Reset the config
	err := store.ResetConfig(channelID)
	if err != nil {
		return "", errors.Wrap(err, "Failed to reset configuration")
	}

	// Get default config after reset
	config, err := store.GetConfig(channelID)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get default configuration")
	}

	return fmt.Sprintf("Configuration has been reset! Now using the default item: %s (at $%.2f each).",
		config.ItemName, config.ItemPrice), nil
}

// safeHandleStatusCommand returns the current configuration for a channel with error handling
func safeHandleStatusCommand(store slack.ChannelConfigStore, channelID string) (string, error) {
	config, err := store.GetConfig(channelID)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get configuration")
	}

	// Check if this is a custom or default config
	isCustom := false
	if checker, ok := store.(slack.ConfigExistsChecker); ok {
		isCustom = checker.ConfigExists(channelID)
	}

	if isCustom {
		return fmt.Sprintf("Current configuration: %s (at $%.2f each).",
			config.ItemName, config.ItemPrice), nil
	} else {
		return fmt.Sprintf("This channel is using the default configuration: %s (at $%.2f each).",
			config.ItemName, config.ItemPrice), nil
	}
}

// handleHelpCommand returns help information about how to use the bot
func handleHelpCommand() string {
	return `*SnagBot Help*

SnagBot automatically responds to messages containing dollar amounts by converting them to a fun comparison.

*Available Commands:*
• /snagbot or /snagbot status - Show current configuration
• /snagbot item "coffee" price 5.00 - Set custom item and price
• /snagbot reset - Reset to default configuration
• /snagbot help - Show this help message

By default, dollar amounts are converted to Bunnings snags at $3.50 each.`
}

// handleConfigCommandWithService processes a configuration command with the specified service
// This function is used by tests and addresses the missing function issue
func handleConfigCommandWithService(text, channelID string, service *slack.SlackService) string {
	// Parse the command
	result, err := ParseConfigCommand(text)
	if err != nil {
		logging.Error("Error parsing command: %v", err)
		return FormatCommandErrorResponse(err)
	}

	// Update the channel configuration
	err = service.ConfigStore.UpdateConfig(channelID, result.ItemName, result.ItemPrice)
	if err != nil {
		logging.Error("Error updating channel config: %v", err)
		return fmt.Sprintf("Error updating configuration: %v", err)
	}

	// Return success message
	return FormatCommandResponse(result)
}
