package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcncl/snagbot/internal/api"
	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/internal/errors"
	"github.com/mcncl/snagbot/internal/logging"
)

// Application represents the main application
type Application struct {
	Config     *config.Config
	HttpServer *http.Server
	Router     http.Handler
}

// New creates a new Application instance
func New() (*Application, error) {
	// Initialize logging
	logging.SetGlobalLevel(logging.INFO)
	logging.Info("Initializing SnagBot application")

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load configuration")
	}

	// Set up routes
	router := api.SetupSimpleRouter(cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	app := &Application{
		Config:     cfg,
		HttpServer: server,
		Router:     router,
	}

	return app, nil
}

// loadConfig loads the application configuration
func loadConfig() (*config.Config, error) {
	cfg := config.New()

	// Validate important configuration values
	if cfg.SlackBotToken == "" {
		logging.Warn("SLACK_BOT_TOKEN environment variable not set")
		return nil, errors.New(errors.ErrInternalServer, "Slack bot token is required")
	}

	if cfg.SlackSigningSecret == "" {
		logging.Warn("SLACK_SIGNING_SECRET environment variable not set")
		return nil, errors.New(errors.ErrInternalServer, "Slack signing secret is required")
	}

	logging.Info("Configuration loaded successfully")
	logging.Debug("Using default item: %s at $%.2f", cfg.DefaultItemName, cfg.DefaultItemPrice)
	return cfg, nil
}

// Start starts the application
func (a *Application) Start() error {
	logging.Info("Starting SnagBot on port %s", a.Config.Port)

	// Start server in a goroutine
	go func() {
		if err := a.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Fatal("Server failed: %v", err)
		}
	}()

	logging.Info("SnagBot is now running")
	return nil
}

// WaitForShutdown waits for a shutdown signal and gracefully shuts down the server
func (a *Application) WaitForShutdown() {
	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-quit
	logging.Info("Received signal: %v", sig)
	logging.Info("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := a.HttpServer.Shutdown(ctx); err != nil {
		logging.Error("Server forced to shutdown: %v", err)
	}

	logging.Info("Server exited properly")
}

// Run is a convenience function that starts the application and waits for shutdown
func (a *Application) Run() error {
	if err := a.Start(); err != nil {
		return err
	}
	a.WaitForShutdown()
	return nil
}

// RunWithContext runs the application with a context for testing
func (a *Application) RunWithContext(ctx context.Context) error {
	if err := a.Start(); err != nil {
		return err
	}

	// Wait for context cancellation or shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logging.Info("Context cancelled, shutting down...")
	case sig := <-quit:
		logging.Info("Received signal: %v", sig)
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := a.HttpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %v", err)
	}

	logging.Info("Server exited properly")
	return nil
}
