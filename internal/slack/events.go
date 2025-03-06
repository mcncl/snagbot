package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/errors"
	"github.com/mcncl/snagbot/internal/logging"
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
		logging.Warn("Debug endpoint accessed - THIS SHOULD BE DISABLED IN PRODUCTION")

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

		if err := json.NewEncoder(w).Encode(debug); err != nil {
			logging.Error("Failed to encode debug response: %v", err)
			http.Error(w, "Error generating debug information", http.StatusInternalServerError)
			return
		}
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
			logging.Warn("Method not allowed: %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check if the Slack signing secret is configured
		if cfg.SlackSigningSecret == "" {
			logging.Error("Slack signing secret not configured")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			appErr := errors.WrapAndLog(err, "Error reading request body")
			http.Error(w, appErr.Message, http.StatusBadRequest)
			return
		}

		// Verify Slack signature
		logging.Debug("Verifying Slack signature with secret of length: %d", len(cfg.SlackSigningSecret))
		sv, err := slack.NewSecretsVerifier(r.Header, cfg.SlackSigningSecret)
		if err != nil {
			appErr := errors.WrapAndLog(err, "Error creating secrets verifier")
			http.Error(w, appErr.Message, http.StatusBadRequest)
			return
		}

		if _, err := sv.Write(body); err != nil {
			appErr := errors.WrapAndLog(err, "Error writing to verifier")
			http.Error(w, appErr.Message, http.StatusBadRequest)
			return
		}

		if err := sv.Ensure(); err != nil {
			// Handle signature validation error
			logging.Error("Signature verification failed: %v", err)
			logging.Debug("Request headers: %v", r.Header)
			logging.Debug("Body length: %d bytes", len(body))

			// Don't log the entire body as it may contain sensitive information
			if len(body) > 0 {
				// Log just the first few bytes for debugging
				maxBytes := min(len(body), 20)
				logging.Debug("First %d bytes of body: %v", maxBytes, body[:maxBytes])
			}

			http.Error(w, "Invalid request signature", http.StatusUnauthorized)
			return
		}

		// Parse the event
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			appErr := errors.WrapAndLog(err, "Error parsing Slack event")
			http.Error(w, appErr.Message, http.StatusBadRequest)
			return
		}

		// Handle URL verification (required by Slack when setting up Events API)
		if eventsAPIEvent.Type == slackevents.URLVerification {
			logging.Info("Handling URL verification challenge from Slack")
			var r *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &r); err != nil {
				appErr := errors.WrapAndLog(err, "Error unmarshalling challenge")
				http.Error(w, appErr.Message, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(r.Challenge))
			logging.Info("Successfully responded to URL verification challenge")
			return
		}

		// Handle callback events
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			// Immediately return a 200 OK to Slack
			// This is important to do quickly, before any processing
			w.WriteHeader(http.StatusOK)

			// Log the event type received
			if innerEvent := eventsAPIEvent.InnerEvent; innerEvent.Data != nil {
				logging.Info("Received Slack callback event: %T", innerEvent.Data)
			}

			// Process the event in a goroutine to avoid blocking
			go func() {
				defer func() {
					// Recover from any panics in the goroutine to prevent crashing
					if r := recover(); r != nil {
						logging.Error("Panic in event handler: %v", r)
					}
				}()

				if err := handleCallbackEvent(eventsAPIEvent, configStore, api); err != nil {
					logging.Error("Error handling callback event: %v", err)
				}
			}()
			return
		}

		// If we reach here, it's an unknown event type
		logging.Warn("Unknown event type: %s", eventsAPIEvent.Type)
		http.Error(w, "Unknown event type", http.StatusBadRequest)
	}
}

// handleCallbackEvent processes Slack callback events
func handleCallbackEvent(event slackevents.EventsAPIEvent, configStore ChannelConfigStore, api SlackAPI) error {
	innerEvent := event.InnerEvent

	// Check if it's a message event
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		// Process the message
		return ProcessMessageEvent(ev, configStore, api)
	default:
		eventType := fmt.Sprintf("%T", innerEvent.Data)
		logging.Debug("Unhandled event type: %s", eventType)
		return errors.Newf(errors.ErrInvalidRequest, "unhandled event type: %s", eventType)
	}
}

// HandleErrorWithResponse sends an error message to the user via Slack
func HandleErrorWithResponse(err error, ev *slackevents.MessageEvent, api SlackAPI) {
	// Don't send any message for nil errors
	if err == nil {
		return
	}

	// Create a user-friendly error message
	message := "Oops! Something went wrong. I couldn't process that message properly."

	// Log the error
	logging.Error("Error processing message: %v", err)

	// Send the error message as a thread reply
	response := SlackResponse{
		ChannelID: ev.Channel,
		Text:      message,
		ThreadTS:  ev.TimeStamp,
	}

	if err := api.PostMessage(response); err != nil {
		logging.Error("Failed to send error response to Slack: %v", err)
	}
}
