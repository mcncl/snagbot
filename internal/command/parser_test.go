package command

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfigCommand(t *testing.T) {
	tests := []struct {
		name        string
		commandText string
		expected    CommandParseResult
		expectError bool
		errorType   error
	}{
		{
			name:        "Valid command with quoted item",
			commandText: "item \"coffee\" price 5.00",
			expected:    CommandParseResult{ItemName: "coffee", ItemPrice: 5.00},
			expectError: false,
		},
		{
			name:        "Valid command with single word item",
			commandText: "item coffee price 5.00",
			expected:    CommandParseResult{ItemName: "coffee", ItemPrice: 5.00},
			expectError: false,
		},
		{
			name:        "Valid command with quoted multi-word item",
			commandText: "item \"Bunnings snags\" price 3.50",
			expected:    CommandParseResult{ItemName: "Bunnings snags", ItemPrice: 3.50},
			expectError: false,
		},
		{
			name:        "Valid command with extra whitespace",
			commandText: "  item   \"coffee\"    price   5.00  ",
			expected:    CommandParseResult{ItemName: "coffee", ItemPrice: 5.00},
			expectError: false,
		},
		{
			name:        "Valid command with integer price",
			commandText: "item coffee price 5",
			expected:    CommandParseResult{ItemName: "coffee", ItemPrice: 5.0},
			expectError: false,
		},
		{
			name:        "Missing item prefix",
			commandText: "coffee price 5.00",
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name:        "Missing item name",
			commandText: "item price 5.00",
			expectError: true,
			errorType:   ErrMissingItem,
		},
		{
			name:        "Missing price keyword",
			commandText: "item coffee 5.00",
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name:        "Missing price value",
			commandText: "item coffee price",
			expectError: true,
			errorType:   ErrMissingPrice,
		},
		{
			name:        "Invalid price format",
			commandText: "item coffee price abc",
			expectError: true,
			errorType:   ErrInvalidPrice,
		},
		{
			name:        "Negative price",
			commandText: "item coffee price -5.00",
			expectError: true,
			errorType:   ErrInvalidPrice,
		},
		{
			name:        "Zero price",
			commandText: "item coffee price 0",
			expectError: true,
			errorType:   ErrInvalidPrice,
		},
		{
			name:        "Unclosed quote",
			commandText: "item \"coffee price 5.00",
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name:        "Case insensitivity with preservation",
			commandText: "item \"Coffee\" PRICE 5.00",
			expected:    CommandParseResult{ItemName: "Coffee", ItemPrice: 5.00},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseConfigCommand(test.commandText)

			if test.expectError {
				assert.Error(t, err)
				if test.errorType != nil {
					assert.True(t, errors.Is(err, test.errorType), "Expected error type %v, got %v", test.errorType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected.ItemName, result.ItemName)
				assert.Equal(t, test.expected.ItemPrice, result.ItemPrice)
			}
		})
	}
}

func TestFormatCommandResponse(t *testing.T) {
	result := CommandParseResult{
		ItemName:  "coffee",
		ItemPrice: 5.00,
	}

	response := FormatCommandResponse(result)
	expected := "Configuration updated! Now converting dollar amounts to coffee (at $5.00 each)."
	assert.Equal(t, expected, response)
}

func TestFormatCommandErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		shouldContain  []string
		shouldNotMatch string
	}{
		{
			name:          "Invalid command error",
			err:           fmt.Errorf("%w: command must start with 'item'", ErrInvalidCommand),
			shouldContain: []string{"invalid command syntax", "command must start with 'item'", "Usage example:"},
		},
		{
			name:          "Missing item error",
			err:           ErrMissingItem,
			shouldContain: []string{"missing item name", "Please provide an item name", "Usage example:"},
		},
		{
			name:          "Missing price error",
			err:           ErrMissingPrice,
			shouldContain: []string{"missing price value", "Please provide a price value", "Usage example:"},
		},
		{
			name:          "Invalid price error",
			err:           fmt.Errorf("%w: 3.5.0 is not a valid number", ErrInvalidPrice),
			shouldContain: []string{"price must be a positive number", "3.5.0 is not a valid number", "Usage example:"},
		},
		{
			name:          "Default error case",
			err:           errors.New("unexpected error"),
			shouldContain: []string{"unexpected error", "Usage example:"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := FormatCommandErrorResponse(test.err)

			for _, text := range test.shouldContain {
				assert.Contains(t, response, text)
			}

			if test.shouldNotMatch != "" {
				assert.NotContains(t, response, test.shouldNotMatch)
			}
		})
	}
}
