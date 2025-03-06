package main

import (
	"github.com/mcncl/snagbot/internal/app"
	"github.com/mcncl/snagbot/internal/logging"
)

func main() {
	// Initialize logging
	logging.SetGlobalLevel(logging.INFO)
	logging.Info("Starting SnagBot...")

	// Create and run the application
	application, err := app.New()
	if err != nil {
		logging.Fatal("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		logging.Fatal("Application error: %v", err)
	}
}
