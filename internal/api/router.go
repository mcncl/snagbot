package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mcncl/snagbot/internal/command"
	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/slack"
)

// Response is a simple structure for API responses
type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// SetupSimpleRouter creates a simple HTTP router without using the mux package
func SetupSimpleRouter(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	// Hello world endpoint
	mux.HandleFunc("/hello", helloWorldHandler)

	// Debug endpoint - REMOVE IN PRODUCTION
	mux.HandleFunc("/debug", slack.DebugHandler(cfg))

	// Slack event endpoint
	mux.HandleFunc("/api/events", slack.EventHandler(cfg))

	// Slack command endpoint
	mux.HandleFunc("/api/commands", command.CommandHandler(cfg))

	// Log available routes
	log.Printf("Available routes: /health, /hello, /debug, /api/events, /api/commands")

	return mux
}

// healthCheckHandler is a simple health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Message: "Snags are cooking 🌭",
		Status:  "OK",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// helloWorldHandler is a simple hello world endpoint
func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Message: "Hello, world! SnagBot is running.",
		Status:  "OK",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
