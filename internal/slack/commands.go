package slack

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/pkg/models"
	"github.com/slack-go/slack"
)

// CommandHandler creates a handler for Slack slash commands
func CommandHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the slash command
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			log.Printf("Error parsing slash command: %v", err)
			http.Error(w, "Error parsing slash command", http.StatusBadRequest)
			return
		}

		// Verify Slack signature
		sv, err := slack.NewSecretsVerifier(r.Header, cfg.SlackSigningSecret)
		if err != nil {
			log.Printf("Error creating secrets verifier: %v", err)
			http.Error(w, "Error verifying request", http.StatusBadRequest)
			return
		}

		if err := sv.Ensure(); err != nil {
			log.Printf("Error verifying signature: %v", err)
			http.Error(w, "Invalid request signature", http.StatusUnauthorized)
			return
		}

		// Handle command
		switch s.Command {
		case "/snagbot":
			response := handleSnagBotCommand(s, cfg)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
			return
		default:
			log.Printf("Unknown command: %s", s.Command)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Unknown command"))
			return
		}
	}
}

// handleSnagBotCommand processes the /snagbot slash command
func handleSnagBotCommand(s slack.SlashCommand, cfg *config.Config) string {
	// Trim the command text
	text := strings.TrimSpace(s.Text)

	// If empty, show help
	if text == "" {
		return formatHelpMessage()
	}

	// Split the text by spaces, but respect quoted strings
	args := parseCommandArgs(text)

	// Check for empty args after parsing
	if len(args) == 0 {
		return formatHelpMessage()
	}

	// Handle subcommands
	switch args[0] {
	case "set":
		return handleSetCommand(args[1:], s.ChannelID)

	case "get":
		return handleGetCommand(s.ChannelID)

	case "reset":
		return handleResetCommand(s.ChannelID, cfg)

	case "help":
		return formatHelpMessage()

	default:
		return formatHelpMessage()
	}
}

// parseCommandArgs parses command arguments, respecting quoted strings
func parseCommandArgs(text string) []string {
	// Split by spaces but preserve quoted strings
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	matches := r.FindAllString(text, -1)

	// Clean up quotes
	var args []string
	for _, match := range matches {
		// Remove surrounding quotes if present
		if strings.HasPrefix(match, "\"") && strings.HasSuffix(match, "\"") {
			match = match[1 : len(match)-1]
		}
		args = append(args, match)
	}

	return args
}

// handleSetCommand handles setting a custom item and price
func handleSetCommand(args []string, channelID string) string {
	// Expected format: set item "item name" price 5.00
	if len(args) < 4 {
		return "Error: Invalid format. Use: `/snagbot set item \"item name\" price 5.00`"
	}

	// Parse the command
	if args[0] != "item" {
		return "Error: Expected 'item' keyword. Use: `/snagbot set item \"item name\" price 5.00`"
	}

	itemName := args[1]

	if args[2] != "price" {
		return "Error: Expected 'price' keyword. Use: `/snagbot set item \"item name\" price 5.00`"
	}

	// Parse price
	price, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return fmt.Sprintf("Error: Invalid price format '%s'. Please provide a valid number.", args[3])
	}

	// Validate price
	if price <= 0 {
		return "Error: Price must be greater than zero."
	}

	// Update channel configuration
	err = UpdateChannelConfig(channelID, itemName, price)
	if err != nil {
		return fmt.Sprintf("Error updating configuration: %v", err)
	}

	return fmt.Sprintf("Success! SnagBot will now convert dollar values to %s at $%.2f each.", itemName, price)
}

// handleGetCommand gets the current channel configuration
func handleGetCommand(channelID string) string {
	config, ok := channelConfigs[channelID]
	if !ok {
		return "This channel is using the default configuration: Bunnings snags at $3.50 each."
	}

	return fmt.Sprintf("Current configuration: %s at $%.2f each.", config.ItemName, config.ItemPrice)
}

// handleResetCommand resets the channel to default configuration
func handleResetCommand(channelID string, cfg *config.Config) string {
	// Create new default config
	newConfig := models.NewChannelConfig(channelID)
	channelConfigs[channelID] = newConfig

	return "Configuration reset to default: Bunnings snags at $3.50 each."
}

// formatHelpMessage formats the help message
func formatHelpMessage() string {
	return "SnagBot Commands:\n" +
		"• `/snagbot set item \"[item name]\" price [price]` - Set custom item and price\n" +
		"• `/snagbot get` - Show current configuration\n" +
		"• `/snagbot reset` - Reset to default configuration\n" +
		"• `/snagbot help` - Show this help message\n\n" +
		"Example: `/snagbot set item \"coffee\" price 5.00`"
}
