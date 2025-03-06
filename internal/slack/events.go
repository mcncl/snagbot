package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// min returns the smaller of x or y.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// DebugHandler creates a simple handler for debugging environment variables
// WARNING: This should be removed in production as it exposes sensitive information
func DebugHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only use this endpoint for debugging and remove before production!
		w.Header().Set("Content-Type", "application/json")

		// Don't show the full secrets, just their length and first few characters
		signingSecretInfo := "Not set"
		if cfg.SlackSigningSecret != "" {
			prefix := cfg.SlackSigningSecret
			if len(prefix) > 4 {
				prefix = prefix[:4] + "..."
			}
			signingSecretInfo = fmt.Sprintf("Length: %d, Prefix: %s", len(cfg.SlackSigningSecret), prefix)
		}

		botTokenInfo := "Not set"
		if cfg.SlackBotToken != "" {
			prefix := cfg.SlackBotToken
			if len(prefix) > 8 {
				prefix = prefix[:8] + "..."
			}
			botTokenInfo = fmt.Sprintf("Length: %d, Prefix: %s", len(cfg.SlackBotToken), prefix)
		}

		debug := map[string]string{
			"port":              cfg.Port,
			"signingSecret":     signingSecretInfo,
			"botToken":          botTokenInfo,
			"defaultItemName":   cfg.DefaultItemName,
			"defaultItemPrice":  fmt.Sprintf("%.2f", cfg.DefaultItemPrice),
			"timestamp":         time.Now().Format(time.RFC3339),
			"environmentSource": "Go environment",
		}

		json.NewEncoder(w).Encode(debug)
	}
}

// EventHandler creates a handler for Slack events
func EventHandler(cfg *config.Config) http.HandlerFunc {
	// Create the configuration store
	configStore := NewInMemoryConfigStoreWithConfig(cfg)

	// Create the Slack API client
	api := NewRealSlackAPI(cfg.SlackBotToken)

	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests for events
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		// Verify Slack signature with additional debugging
		log.Printf("DEBUG: Verifying Slack signature with secret of length: %d", len(cfg.SlackSigningSecret))
		log.Printf("DEBUG: Request timestamp: %s", r.Header.Get("X-Slack-Request-Timestamp"))
		log.Printf("DEBUG: Request signature: %s", r.Header.Get("X-Slack-Signature"))

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
			log.Printf("DEBUG: Body length: %d bytes", len(body))
			log.Printf("DEBUG: First few bytes of body: %v", body[:min(len(body), 20)])
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
			// Immediately return a 200 OK to Slack
			// This is important to do quickly, before any processing
			w.WriteHeader(http.StatusOK)

			// Process the event in a goroutine to avoid blocking
			go handleCallbackEvent(eventsAPIEvent, configStore, api)
			return
		}

		// If we reach here, it's an unknown event type
		log.Printf("Unknown event type: %s", eventsAPIEvent.Type)
		http.Error(w, "Unknown event type", http.StatusBadRequest)
	}
}

// handleCallbackEvent processes Slack callback events
func handleCallbackEvent(event slackevents.EventsAPIEvent, configStore ChannelConfigStore, api SlackAPI) {
	innerEvent := event.InnerEvent

	// Check if it's a message event
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		// Process the message
		err := ProcessMessageEvent(ev, configStore, api)
		if err != nil {
			log.Printf("Error processing message event: %v", err)
		}
	default:
		log.Printf("Unhandled event type: %T", innerEvent.Data)
	}
}
