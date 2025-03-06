package slack

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// EventHandler creates a handler for Slack events
func EventHandler(cfg *config.Config) http.HandlerFunc {
	// Create the configuration store
	configStore := NewInMemoryConfigStoreWithConfig(cfg)

	// Create the Slack API client
	api := NewRealSlackAPI(cfg.SlackBotToken)

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
			go handleCallbackEventWithStore(eventsAPIEvent, configStore, api)

			// Immediately return a 200 OK to Slack
			w.WriteHeader(http.StatusOK)
			return
		}

		// If we reach here, it's an unknown event type
		log.Printf("Unknown event type: %s", eventsAPIEvent.Type)
		http.Error(w, "Unknown event type", http.StatusBadRequest)
	}
}

// handleCallbackEventWithStore processes Slack callback events using the config store
func handleCallbackEventWithStore(event slackevents.EventsAPIEvent, configStore ChannelConfigStore, api SlackAPI) {
	innerEvent := event.InnerEvent

	// Check if it's a message event
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		// Process the message using our updated function
		err := ProcessMessageEvent(ev, configStore, api)
		if err != nil {
			log.Printf("Error processing message event: %v", err)
		}
	}
}
