package slack

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CommandParseResult holds the parsed item name and price
type CommandParseResult struct {
	ItemName  string
	ItemPrice float64
}

var (
	// ErrInvalidCommand is returned when the command syntax is invalid
	ErrInvalidCommand = errors.New("invalid command syntax")

	// ErrMissingItem is returned when the item name is missing
	ErrMissingItem = errors.New("missing item name")

	// ErrMissingPrice is returned when the price is missing
	ErrMissingPrice = errors.New("missing price value")

	// ErrInvalidPrice is returned when the price is not a valid positive number
	ErrInvalidPrice = errors.New("price must be a positive number")
)

// ParseConfigCommand parses a Slack slash command for configuring the bot.
// Expected format: /snagbot item "item name" price 5.00
// The item name can be in quotes (for multi-word items) or a single word without quotes.
// Note: Case is preserved for the item name to allow for proper pluralization.
func ParseConfigCommand(commandText string) (CommandParseResult, error) {
	result := CommandParseResult{}

	// Normalize whitespace in the command text
	// This replaces multiple spaces with a single space throughout the string
	spaceRegex := regexp.MustCompile(`\s+`)
	commandText = spaceRegex.ReplaceAllString(strings.TrimSpace(commandText), " ")

	// Check if the command starts with "item"
	if !strings.HasPrefix(strings.ToLower(commandText), "item") {
		return result, fmt.Errorf("%w: command must start with 'item'", ErrInvalidCommand)
	}

	// Remove the "item" prefix
	// We use len("item") instead of the regex-matched length to handle case differences
	commandText = commandText[len("item"):]
	commandText = strings.TrimSpace(commandText)

	// Check if we have "price" immediately after "item" (missing item)
	if strings.HasPrefix(strings.ToLower(commandText), "price") {
		return result, ErrMissingItem
	}

	// Extract item name in quotes if present
	itemName := ""
	remainingText := ""

	if strings.HasPrefix(commandText, "\"") {
		// Look for the closing quote
		quoteRegex := regexp.MustCompile(`^"([^"]+)"`)
		matches := quoteRegex.FindStringSubmatch(commandText)

		if len(matches) > 1 {
			itemName = matches[1]
			remainingText = strings.TrimSpace(commandText[len(matches[0]):])
		} else {
			// Opening quote without closing quote
			return result, fmt.Errorf("%w: unclosed quote in item name", ErrInvalidCommand)
		}
	} else {
		// No quotes, so the item name is the first word
		parts := strings.SplitN(commandText, " ", 2)
		if len(parts) == 0 || parts[0] == "" || strings.HasPrefix(strings.ToLower(parts[0]), "price") {
			return result, ErrMissingItem
		}

		itemName = parts[0]
		if len(parts) > 1 {
			remainingText = strings.TrimSpace(parts[1])
		}
	}

	// Validate item name
	if itemName == "" {
		return result, ErrMissingItem
	}

	// Check for price keyword
	if remainingText == "" {
		return result, ErrMissingPrice
	}

	// Handle case insensitivity for "price" keyword
	if !strings.HasPrefix(strings.ToLower(remainingText), "price") {
		return result, fmt.Errorf("%w: expected 'price' keyword after item name", ErrInvalidCommand)
	}

	// Extract price value
	priceText := remainingText[strings.Index(strings.ToLower(remainingText), "price")+len("price"):]
	priceText = strings.TrimSpace(priceText)

	if priceText == "" {
		return result, ErrMissingPrice
	}

	// Parse price as float
	price, err := strconv.ParseFloat(priceText, 64)
	if err != nil {
		return result, fmt.Errorf("%w: %s is not a valid number", ErrInvalidPrice, priceText)
	}

	// Validate price is positive
	if price <= 0 {
		return result, ErrInvalidPrice
	}

	// Set result values
	result.ItemName = itemName
	result.ItemPrice = price

	return result, nil
}

// FormatCommandResponse formats a response message for the command
func FormatCommandResponse(result CommandParseResult) string {
	return fmt.Sprintf("Configuration updated! Now converting dollar amounts to %s (at $%.2f each).", result.ItemName, result.ItemPrice)
}

// FormatCommandErrorResponse formats an error message for the command
func FormatCommandErrorResponse(err error) string {
	// Start with the basic error message
	errorMsg := err.Error()

	// Add a helpful usage example
	helpText := "\n\nUsage example: `/snagbot item \"coffee\" price 5.00`"

	// Depending on the error type, provide more specific instructions
	switch {
	case errors.Is(err, ErrInvalidCommand):
		errorMsg += "\nThe command format is invalid." + helpText
	case errors.Is(err, ErrMissingItem):
		errorMsg += "\nPlease provide an item name." + helpText
	case errors.Is(err, ErrMissingPrice):
		errorMsg += "\nPlease provide a price value." + helpText
	case errors.Is(err, ErrInvalidPrice):
		errorMsg += "\nThe price must be a positive number (e.g., 3.50)." + helpText
	default:
		errorMsg += helpText
	}

	return errorMsg
}
