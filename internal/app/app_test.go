package app

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAppInitialization(t *testing.T) {
	// Save original environment variables to restore after the test
	originalBotToken := os.Getenv("SLACK_BOT_TOKEN")
	originalSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")

	// Set test values
	os.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	os.Setenv("SLACK_SIGNING_SECRET", "test-signing-secret")

	// Cleanup environment after test
	defer func() {
		os.Setenv("SLACK_BOT_TOKEN", originalBotToken)
		os.Setenv("SLACK_SIGNING_SECRET", originalSigningSecret)
	}()

	// Initialize the application
	app, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, app)

	// Check that the configuration was loaded correctly
	assert.Equal(t, "xoxb-test-token", app.Config.SlackBotToken)
	assert.Equal(t, "test-signing-secret", app.Config.SlackSigningSecret)

	// Check that the server was initialized
	assert.NotNil(t, app.HttpServer)
	assert.NotNil(t, app.Router)
}

func TestAppRunWithContext(t *testing.T) {
	// Skip in CI environment as it requires network binding
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping in CI environment")
	}

	// Save original environment variables to restore after the test
	originalBotToken := os.Getenv("SLACK_BOT_TOKEN")
	originalSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")

	// Set test values
	os.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	os.Setenv("SLACK_SIGNING_SECRET", "test-signing-secret")

	// Cleanup environment after test
	defer func() {
		os.Setenv("SLACK_BOT_TOKEN", originalBotToken)
		os.Setenv("SLACK_SIGNING_SECRET", originalSigningSecret)
	}()

	// Initialize the application with a unique port
	os.Setenv("PORT", "9876")
	app, err := New()
	assert.NoError(t, err)

	// Create a context that will cancel after a short time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run the application with the context - it should exit when the context is canceled
	err = app.RunWithContext(ctx)
	assert.NoError(t, err)
}

func TestAppWithMissingConfig(t *testing.T) {
	// Save original environment variables to restore after the test
	originalBotToken := os.Getenv("SLACK_BOT_TOKEN")
	originalSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")

	// Clear required environment variables
	os.Setenv("SLACK_BOT_TOKEN", "")
	os.Setenv("SLACK_SIGNING_SECRET", "")

	// Cleanup environment after test
	defer func() {
		os.Setenv("SLACK_BOT_TOKEN", originalBotToken)
		os.Setenv("SLACK_SIGNING_SECRET", originalSigningSecret)
	}()

	// Initialize the application - should fail
	app, err := New()
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "Slack bot token is required")
}
