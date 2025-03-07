package slack

import (
	"testing"

	"github.com/mcncl/snagbot/internal/config"
	"github.com/mcncl/snagbot/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryConfigStore_GetConfig(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		setupFunc func(*InMemoryConfigStore)
		expected  models.ChannelConfig
	}{
		{
			name:      "Get default config for new channel",
			channelID: "C12345",
			setupFunc: nil,
			expected: models.ChannelConfig{
				ChannelID: "C12345",
				ItemName:  "Test Snags", // Using our test default
				ItemPrice: 4.50,
			},
		},
		{
			name:      "Get existing config",
			channelID: "C67890",
			setupFunc: func(store *InMemoryConfigStore) {
				store.UpdateConfig("C67890", "coffee", 5.00)
			},
			expected: models.ChannelConfig{
				ChannelID: "C67890",
				ItemName:  "coffee",
				ItemPrice: 5.00,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a test config
			testCfg := &config.Config{
				DefaultItemName:  "Test Snags",
				DefaultItemPrice: 4.50,
			}

			// Create store with test config - use the new function
			store := NewInMemoryConfigStoreWithConfig(testCfg)

			// Run setup if provided
			if test.setupFunc != nil {
				test.setupFunc(store)
			}

			// Get config
			result, err := store.GetConfig(test.channelID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify result
			assert.Equal(t, test.expected.ChannelID, result.ChannelID)
			assert.Equal(t, test.expected.ItemName, result.ItemName)
			assert.Equal(t, test.expected.ItemPrice, result.ItemPrice)
		})
	}
}

func TestInMemoryConfigStore_UpdateConfig(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		itemName   string
		itemPrice  float64
		expectErr  bool
		errorMatch string
	}{
		{
			name:      "Valid update",
			channelID: "C12345",
			itemName:  "coffee",
			itemPrice: 5.00,
			expectErr: false,
		},
		{
			name:       "Zero price",
			channelID:  "C12345",
			itemName:   "coffee",
			itemPrice:  0,
			expectErr:  true,
			errorMatch: "item price must be greater than zero",
		},
		{
			name:       "Negative price",
			channelID:  "C12345",
			itemName:   "coffee",
			itemPrice:  -1.00,
			expectErr:  true,
			errorMatch: "item price must be greater than zero",
		},
		{
			name:       "Empty item name",
			channelID:  "C12345",
			itemName:   "",
			itemPrice:  5.00,
			expectErr:  true,
			errorMatch: "item name cannot be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create store with nil config for default values - use the new function
			store := NewInMemoryConfigStoreWithConfig(nil)

			// Update config
			err := store.UpdateConfig(test.channelID, test.itemName, test.itemPrice)

			// Check error
			if test.expectErr {
				assert.Error(t, err)
				if test.errorMatch != "" {
					assert.Contains(t, err.Error(), test.errorMatch)
				}
			} else {
				assert.NoError(t, err)

				// Verify the update was successful
				config, err := store.GetConfig(test.channelID)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				assert.Equal(t, test.channelID, config.ChannelID)
				assert.Equal(t, test.itemName, config.ItemName)
				assert.Equal(t, test.itemPrice, config.ItemPrice)
			}
		})
	}
}

func TestInMemoryConfigStore_ResetConfig(t *testing.T) {
	// Create a test config
	testCfg := &config.Config{
		DefaultItemName:  "Test Snags",
		DefaultItemPrice: 4.50,
	}

	// Create store with test config - use the new function
	store := NewInMemoryConfigStoreWithConfig(testCfg)

	// Setup initial state
	channelID := "C12345"
	err := store.UpdateConfig(channelID, "coffee", 5.00)
	assert.NoError(t, err)

	// Verify initial config
	config, err := store.GetConfig(channelID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, "coffee", config.ItemName)
	assert.Equal(t, 5.00, config.ItemPrice)

	// Reset config
	err = store.ResetConfig(channelID)
	assert.NoError(t, err)

	// Verify config has been reset
	config, err = store.GetConfig(channelID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, testCfg.DefaultItemName, config.ItemName)
	assert.Equal(t, testCfg.DefaultItemPrice, config.ItemPrice)

	// Verify the config is no longer stored
	assert.False(t, store.ConfigExists(channelID))
}

func TestInMemoryConfigStore_ConfigExists(t *testing.T) {
	// Use the new function
	store := NewInMemoryConfigStoreWithConfig(nil)
	channelID := "C12345"

	// Initially, no config exists
	assert.False(t, store.ConfigExists(channelID))

	// Add a config
	err := store.UpdateConfig(channelID, "coffee", 5.00)
	assert.NoError(t, err)

	// Now it exists
	assert.True(t, store.ConfigExists(channelID))

	// Reset it
	err = store.ResetConfig(channelID)
	assert.NoError(t, err)

	// Now it doesn't exist again
	assert.False(t, store.ConfigExists(channelID))
}
