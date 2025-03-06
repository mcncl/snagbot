package slack

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/pkg/models"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

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
		// Process the message using our new function
		ProcessMessageEvent(ev, cfg)
	}
}

// ProcessMessageEvent is the implementation that integrates with the real Slack events handler
func ProcessMessageEvent(ev *slackevents.MessageEvent, cfg *config.Config) {
	// Skip bot messages to prevent loops
	if ev.BotID != "" || ev.SubType == "bot_message" {
		return
	}

	// Skip message changes/edits for now (can be implemented later)
	if ev.SubType == "message_changed" {
		return
	}

	// Detect dollar values and calculate conversions
	dollarValues := calculator.ExtractDollarValues(ev.Text)
	if len(dollarValues) == 0 {
		return // No dollar values found, nothing to do
	}

	// Calculate total
	total := calculator.SumDollarValues(dollarValues)

	// Get or create channel configuration using our existing in-memory store
	// In production, this would use a database-backed store
	channelConfig, ok := channelConfigs[ev.Channel]
	if !ok {
		channelConfig = models.NewChannelConfig(ev.Channel)
		channelConfigs[ev.Channel] = channelConfig
	}

	// Calculate number of items
	count := calculator.CalculateItemCount(total, channelConfig.ItemPrice)

	// Format response message
	message := calculator.FormatResponse(count, channelConfig.ItemName)

	// Create Slack API client
	api := slack.New(cfg.SlackBotToken)

	// Send response as a thread
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

// EventHandlerWithService creates an HTTP handler for Slack events using our service
func EventHandlerWithService(service *SlackService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This is a placeholder showing how it would be structured
		// For a full service-based architecture implementation
	}
}

// HandleMessageEvent processes a Slack message event
func (s *SlackService) HandleMessageEvent(ev *slackevents.MessageEvent) error {
	// Skip bot messages to prevent loops
	if ev.BotID != "" || ev.SubType == "bot_message" {
		return nil
	}

	// Skip message changes/edits for now (can be implemented later)
	if ev.SubType == "message_changed" {
		return nil
	}

	// Detect dollar values and calculate conversions
	dollarValues := calculator.ExtractDollarValues(ev.Text)
	if len(dollarValues) == 0 {
		return nil // No dollar values found, nothing to do
	}

	// Calculate total
	total := calculator.SumDollarValues(dollarValues)

	// Get channel configuration
	channelConfig := s.configStore.GetConfig(ev.Channel)

	// Calculate number of items
	count := calculator.CalculateItemCount(total, channelConfig.ItemPrice)

	// Format response message
	message := calculator.FormatResponse(count, channelConfig.ItemName)

	// Send response
	response := SlackResponse{
		ChannelID: ev.Channel,
		Text:      message,
		ThreadTS:  ev.TimeStamp, // Reply in thread
	}

	return s.slackAPI.PostMessage(response)
}
