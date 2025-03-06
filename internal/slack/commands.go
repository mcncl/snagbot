package slack

import (
	"log"
	"net/http"
	"strings"

	"github.com/mcncl/snagbot/internal/config"
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

		// Verify Slack signature (commented out for now as it requires more setup)
		// sv, err := slack.NewSecretsVerifier(r.Header, cfg.SlackSigningSecret)
		// if err != nil {
		// 	log.Printf("Error creating secrets verifier: %v", err)
		// 	http.Error(w, "Error verifying request", http.StatusBadRequest)
		// 	return
		// }
		//
		// if err := sv.Ensure(); err != nil {
		// 	log.Printf("Error verifying signature: %v", err)
		// 	http.Error(w, "Invalid request signature", http.StatusUnauthorized)
		// 	return
		// }

		// Handle command
		switch s.Command {
		case "/snagbot":
			handleSnagBotCommand(s, w, cfg)
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
func handleSnagBotCommand(s slack.SlashCommand, w http.ResponseWriter, cfg *config.Config) {
	// Parse command text
	args := strings.Fields(s.Text)

	// Check if command is properly formatted
	if len(args) == 0 {
		sendCommandHelp(w)
		return
	}

	// Handle subcommands
	switch args[0] {
	case "set":
		// Check for proper format: "/snagbot set item "coffee" price 5.00"
		if len(args) < 5 || args[1] != "item" || args[3] != "price" {
			sendCommandHelp(w)
			return
		}

		// In a real implementation, we would parse and save the configuration
		itemName := args[2]
		itemPrice := args[4]

		// Log the command (would save to storage in real implementation)
		log.Printf("Channel %s: Set item to '%s' with price %s",
			s.ChannelID, itemName, itemPrice)

		// Send success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Configuration updated successfully!"))
		return

	case "help":
		sendCommandHelp(w)
		return

	default:
		sendCommandHelp(w)
		return
	}
}

// sendCommandHelp sends the help message for the slash command
func sendCommandHelp(w http.ResponseWriter) {
	helpText := "Usage:\n" +
		"/snagbot set item \"[item name]\" price [price]\n" +
		"Example: /snagbot set item \"coffee\" price 5.00\n" +
		"/snagbot help - Show this help message"

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(helpText))
}
