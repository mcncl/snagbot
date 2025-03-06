package slack

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/mcncl/snagbot/internal/calculator"
	"github.com/mcncl/snagbot/internal/config"
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

		// Verify Slack signature (commented out for now as it requires more setup)
		// sv, err := slack.NewSecretsVerifier(r.Header, cfg.SlackSigningSecret)
		// if err != nil {
		// 	log.Printf("Error creating secrets verifier: %v", err)
		// 	http.Error(w, "Error verifying request", http.StatusBadRequest)
		// 	return
		// }

		// if _, err := sv.Write(body); err != nil {
		// 	log.Printf("Error writing to verifier: %v", err)
		// 	http.Error(w, "Error verifying request", http.StatusBadRequest)
		// 	return
		// }

		// if err := sv.Ensure(); err != nil {
		// 	log.Printf("Error verifying signature: %v", err)
		// 	http.Error(w, "Invalid request signature", http.StatusUnauthorized)
		// 	return
		// }

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
			// For this prototype, we'll just log the event
			log.Printf("Received callback event: %v", eventsAPIEvent)

			// In a real implementation, we would process the event further
			// handleCallbackEvent(eventsAPIEvent, cfg)

			// For now, we'll simulate proper processing
			w.WriteHeader(http.StatusOK)
			return
		}

		// If we reach here, it's an unknown event type
		log.Printf("Unknown event type: %s", eventsAPIEvent.Type)
		http.Error(w, "Unknown event type", http.StatusBadRequest)
	}
}

// handleCallbackEvent processes Slack callback events
// This would be filled in with actual logic in a real implementation
func handleCallbackEvent(event slackevents.EventsAPIEvent, cfg *config.Config) {
	innerEvent := event.InnerEvent

	// Check if it's a message event
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		// Skip bot messages to prevent loops
		if ev.BotID != "" {
			return
		}

		// Detect dollar values and calculate conversions
		dollarValues := calculator.ExtractDollarValues(ev.Text)
		if len(dollarValues) == 0 {
			return
		}

		// Calculate total
		total := calculator.SumDollarValues(dollarValues)

		// Get channel configuration (would fetch from storage in real implementation)
		itemName := cfg.DefaultItemName
		itemPrice := cfg.DefaultItemPrice

		// Calculate number of items
		count := calculator.CalculateItemCount(total, itemPrice)

		// Format response message
		message := calculator.FormatResponse(count, itemName)

		// Send response (would use Slack API in real implementation)
		log.Printf("Would send message to channel %s, thread %s: %s",
			ev.Channel, ev.TimeStamp, message)
	}
}
