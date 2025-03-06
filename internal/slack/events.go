package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/pkg/models"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// In-memory storage for channel configurations
// In a production environment, this would be replaced with a persistent database
var channelConfigs = make(map[string]*models.ChannelConfig)

// EventHandler creates a handler for Slack events
func EventHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		// Verify Slack signature
		sv, err := slack.NewSecretsVerifier(r.Header, cfg.SlackSigningSecret)
		if err != nil {
			log.Printf("Error creating secrets verifier: %v", err)
			http.Error(w, "Error verifying request", http.StatusBadRequest)
			return
		}

		if _, err := sv.Write(body); err != nil {
			log.Printf("Error writing to verifier: %v", err)
			http.Error(w, "Error verifying request", http.StatusBadRequest)
			return
		}

		if err := sv.Ensure(); err != nil {
			log.Printf("Error verifying signature: %v", err)
			http.Error(w, "Invalid request signature", http.StatusUnauthorized)
			return
		}

		// Parse the event
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Printf("Error parsing event: %v", err)
			http.Error(w, "Error parsing event", http.StatusBadRequest)
			return
		}

		// Handle URL verification (required by Slack when setting up Events API)
		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &r); err != nil {
				log.Printf("Error unmarshalling challenge: %v", err)
				http.Error(w, "Error parsing challenge", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(r.Challenge))
			return
		}

		// Handle callback events
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			// Process the event in a goroutine to avoid blocking
			go handleCallbackEvent(eventsAPIEvent, cfg)

			// Immediately return a 200 OK to Slack
			w.WriteHeader(http.StatusOK)
			return
		}

		// If we reach here, it's an unknown event type
		log.Printf("Unknown event type: %s", eventsAPIEvent.Type)
		http.Error(w, "Unknown event type", http.StatusBadRequest)
	}
}

// handleCallbackEvent processes Slack callback events
func handleCallbackEvent(event slackevents.EventsAPIEvent, cfg *config.Config) {
	innerEvent := event.InnerEvent

	// Check if it's a message event
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		// Skip bot messages to prevent loops
		if ev.BotID != "" || ev.SubType == "bot_message" {
			return
		}

		// Skip message changes/edits for now (can be implemented later)
		if ev.SubType == "message_changed" {
			return
		}

		// Process the message
		processMessageEvent(ev, cfg)
	}
}

// processMessageEvent handles message events and sends responses
func processMessageEvent(ev *slackevents.MessageEvent, cfg *config.Config) {
	// Detect dollar values and calculate conversions
	dollarValues := calculator.ExtractDollarValues(ev.Text)
	if len(dollarValues) == 0 {
		return // No dollar values found, nothing to do
	}

	// Calculate total
	total := calculator.SumDollarValues(dollarValues)

	// Get channel configuration
	channelConfig := getChannelConfig(ev.Channel, cfg)

	// Calculate number of items
	count := calculator.CalculateItemCount(total, channelConfig.ItemPrice)

	// Format response message
	message := calculator.FormatResponse(count, channelConfig.ItemName)

	// Send response as a thread
	api := slack.New(cfg.SlackBotToken)
	_, _, err := api.PostMessage(
		ev.Channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(ev.TimeStamp), // Reply in thread
	)

	if err != nil {
		log.Printf("Error sending message: %v", err)
	} else {
		log.Printf("Response sent to channel %s: %s", ev.Channel, message)
	}
}

// getChannelConfig retrieves the channel configuration or creates a default one
func getChannelConfig(channelID string, cfg *config.Config) *models.ChannelConfig {
	// Check if we have a config for this channel
	if config, ok := channelConfigs[channelID]; ok {
		return config
	}

	// Create new default config
	newConfig := models.NewChannelConfig(channelID)
	channelConfigs[channelID] = newConfig
	return newConfig
}

// UpdateChannelConfig updates the configuration for a channel
func UpdateChannelConfig(channelID, itemName string, itemPrice float64) error {
	if itemPrice <= 0 {
		return fmt.Errorf("item price must be greater than zero")
	}

	// Get or create channel config
	config, ok := channelConfigs[channelID]
	if !ok {
		config = models.NewChannelConfig(channelID)
		channelConfigs[channelID] = config
	}

	// Update the configuration
	config.SetItem(itemName, itemPrice)
	log.Printf("Updated configuration for channel %s: item=%s, price=%0.2f",
		channelID, itemName, itemPrice)

	return nil
}
