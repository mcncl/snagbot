package calculator

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// ExtractDollarValues extracts all dollar values from a string
// Matches patterns like $35, $35.00, etc.
// ExtractDollarValues extracts all dollar values from a string
// Matches patterns like $35, $35.00, etc.
func ExtractDollarValues(text string) []float64 {
	// Regular expression to match dollar values
	// Handles both whole numbers and decimal values (up to 2 decimal places)
	re := regexp.MustCompile(`\$([0-9]+(\.[0-9]{1,2})?)`)
	matches := re.FindAllStringSubmatch(text, -1)

	// Process the matches to filter out duplicates
	var seen = make(map[string]bool)
	values := make([]float64, 0, len(matches))

	for _, match := range matches {
		if len(match) >= 2 {
			// Use the whole match as key to avoid duplicates
			if !seen[match[0]] {
				seen[match[0]] = true

				// Parse the value (without the $ symbol)
				value, err := strconv.ParseFloat(match[1], 64)
				if err == nil {
					values = append(values, value)
				}
			}
		}
	}

	return values
}

// SumDollarValues sums an array of dollar values
// Returns the total with 2 decimal place precision
func SumDollarValues(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}

	// Round to 2 decimal places for currency precision
	return math.Round(total*100) / 100
}

// CalculateItemCount calculates how many items the dollar amount could buy
// Always rounds up for fun!
func CalculateItemCount(total float64, pricePerItem float64) int {
	// Safety check for invalid inputs
	if total <= 0 || pricePerItem <= 0 {
		return 0
	}

	// Calculate count and round up
	count := math.Ceil(total / pricePerItem)
	return int(count)
}

// FormatResponse creates a fun response message with the item count
// Handles pluralization automatically
func FormatResponse(count int, itemName string) string {
	// Handle zero case
	if count <= 0 {
		return "That wouldn't even buy a single " + getSingularForm(itemName) + "!"
	}

	// Handle pluralization
	if count == 1 {
		return "That's nearly 1 " + getSingularForm(itemName) + "!"
	} else {
		return "That's nearly " + strconv.Itoa(count) + " " + getPluralForm(itemName) + "!"
	}
}

// ProcessMessage is a convenience function that combines all steps
// Takes a message text and price per item, returns the formatted response
func ProcessMessage(text string, pricePerItem float64) string {
	// Extract dollar values
	values := ExtractDollarValues(text)
	if len(values) == 0 {
		return "" // No dollar values found
	}

	// Sum the values
	total := SumDollarValues(values)

	// Calculate the item count
	count := CalculateItemCount(total, pricePerItem)

	// Format and return the response
	return FormatResponse(count, "Bunnings snag")
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

	// Add 's' for simple pluralization
	return itemName + "s"
}
