package calculator

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/mcncl/snagbot/internal/errors"
	"github.com/mcncl/snagbot/internal/logging"
)

// ExtractDollarValues extracts all dollar values from a string
// Matches patterns like $35, $35.00, etc.
func ExtractDollarValues(text string) ([]float64, error) {
	if text == "" {
		logging.Debug("Empty text provided to ExtractDollarValues")
		return []float64{}, nil
	}

	// Regular expression to match dollar values
	// Handles both whole numbers and decimal values (up to 2 decimal places)
	re := regexp.MustCompile(`\$([0-9]+(\.[0-9]{1,2})?)`)
	matches := re.FindAllStringSubmatch(text, -1)

	// Process the matches to filter out duplicates
	var seen = make(map[string]bool)
	values := make([]float64, 0, len(matches))
	invalidValues := make([]string, 0)

	for _, match := range matches {
		if len(match) >= 2 {
			// Use the whole match as key to avoid duplicates
			if !seen[match[0]] {
				seen[match[0]] = true

				// Parse the value (without the $ symbol)
				value, err := strconv.ParseFloat(match[1], 64)
				if err == nil {
					values = append(values, value)
				} else {
					invalidValues = append(invalidValues, match[1])
					logging.Warn("Failed to parse dollar value: %s, error: %v", match[1], err)
				}
			}
		}
	}

	// Log any issues with parsing, but still return what we could parse
	if len(invalidValues) > 0 {
		logging.Warn("Some dollar values could not be parsed: %v", invalidValues)
	}

	logging.Debug("Extracted %d dollar values from text", len(values))
	return values, nil
}

// SumDollarValues sums an array of dollar values
// Returns the total with 2 decimal place precision
func SumDollarValues(values []float64) (float64, error) {
	if len(values) == 0 {
		logging.Debug("Empty array provided to SumDollarValues")
		return 0, nil
	}

	var total float64
	for i, value := range values {
		// Check for negative values, which might be a mistake
		if value < 0 {
			logging.Warn("Negative dollar value found at index %d: %.2f", i, value)
		}
		total += value
	}

	// Round to 2 decimal places for currency precision
	total = math.Round(total*100) / 100

	logging.Debug("Summed %d dollar values to get %.2f", len(values), total)
	return total, nil
}

// CalculateItemCount calculates how many items the dollar amount could buy
// Always rounds up for fun!
func CalculateItemCount(total float64, pricePerItem float64) (int, error) {
	// Safety check for invalid inputs
	if total < 0 {
		err := errors.Newf(errors.ErrInvalidDollarValue, "negative total amount: %.2f", total)
		logging.Warn(err.Error())
		return 0, err
	}

	if pricePerItem <= 0 {
		err := errors.Newf(errors.ErrInvalidDollarValue, "invalid price per item: %.2f", pricePerItem)
		logging.Warn(err.Error())
		return 0, err
	}

	// Calculate count and round up
	count := math.Ceil(total / pricePerItem)
	result := int(count)

	logging.Debug("Calculated item count: $%.2f at $%.2f per item = %d items", total, pricePerItem, result)
	return result, nil
}

// IsExactDivision checks if the division results in a whole number
func IsExactDivision(total float64, pricePerItem float64) bool {
	if pricePerItem <= 0 {
		return false
	}

	// Calculate division and check if it's a whole number
	quotient := total / pricePerItem
	return quotient == float64(int(quotient))
}

// FormatResponse creates a fun response message with the item count
// Handles pluralization automatically and only uses "nearly" for non-exact conversions
func FormatResponse(count int, itemName string, isExactDivision bool) string {
	if itemName == "" {
		logging.Warn("Empty item name provided to FormatResponse, using default")
		itemName = "item"
	}

	// Handle zero case (when the amount is too small to buy even one item)
	if count <= 0 {
		return "That wouldn't even buy a single " + getSingularForm(itemName) + "!"
	}

	// Format the beginning of the response based on whether it's an exact division
	prefix := "That's "
	if !isExactDivision {
		prefix = "That's nearly "
	}

	// Handle pluralization
	if count == 1 {
		return prefix + "1 " + getSingularForm(itemName) + "!"
	} else {
		return prefix + strconv.Itoa(count) + " " + getPluralForm(itemName) + "!"
	}
}

// ProcessMessage is a convenience function that combines all steps
// Takes a message text and price per item, returns the formatted response
func ProcessMessage(text string, pricePerItem float64) (string, error) {
	// Extract dollar values
	values, err := ExtractDollarValues(text)
	if err != nil {
		return "", errors.WrapAndLog(err, "Failed to extract dollar values")
	}

	if len(values) == 0 {
		logging.Debug("No dollar values found in text")
		return "", nil // No dollar values found
	}

	// Sum the values
	total, err := SumDollarValues(values)
	if err != nil {
		return "", errors.WrapAndLog(err, "Failed to sum dollar values")
	}

	// For very small amounts that don't reach 1 item
	if total < pricePerItem {
		// Use the standard "zero" response
		return FormatResponse(0, "Bunnings snag", true), nil
	}

	// Check if the division is exact (to decide whether to use "nearly")
	isExactDivision := IsExactDivision(total, pricePerItem)

	// Calculate the item count
	count, err := CalculateItemCount(total, pricePerItem)
	if err != nil {
		return "", errors.WrapAndLog(err, "Failed to calculate item count")
	}

	// Format and return the response
	response := FormatResponse(count, "Bunnings snag", isExactDivision)
	logging.Debug("Processed message: Total $%.2f, Count %d, Response: %s", total, count, response)
	return response, nil
}

// getSingularForm ensures we have the singular form of the item name
func getSingularForm(itemName string) string {
	// If the item name ends with 's', try to get the singular form
	if strings.HasSuffix(strings.ToLower(itemName), "s") {
		// Check common pluralization patterns
		if strings.HasSuffix(strings.ToLower(itemName), "ies") {
			// Handle words like "candies" -> "candy"
			return itemName[:len(itemName)-3] + "y"
		} else if strings.HasSuffix(strings.ToLower(itemName), "es") {
			// Handle words like "watches" -> "watch"
			return itemName[:len(itemName)-2]
		} else {
			// Simple case like "snags" -> "snag"
			return itemName[:len(itemName)-1]
		}
	}
	return itemName
}

// getPluralForm ensures we have the plural form of the item name
func getPluralForm(itemName string) string {
	// If already plural (ending with 's'), return as is
	if strings.HasSuffix(strings.ToLower(itemName), "s") {
		return itemName
	}

	// Check for special cases that end in 'y'
	if strings.HasSuffix(strings.ToLower(itemName), "y") {
		// Convert "candy" -> "candies" pattern
		return itemName[:len(itemName)-1] + "ies"
	}

	// Add 's' for simple pluralization
	return itemName + "s"
}
