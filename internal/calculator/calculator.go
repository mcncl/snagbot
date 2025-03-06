package calculator

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// ExtractDollarValues extracts all dollar values from a string
func ExtractDollarValues(text string) []float64 {
	// Regular expression to match dollar values
	// Matches patterns like $35, $35.00, etc.
	re := regexp.MustCompile(`\$([0-9]+(\.[0-9]{1,2})?)`)
	matches := re.FindAllStringSubmatch(text, -1)

	// Extract the numeric values
	values := make([]float64, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 {
			// Parse the value (without the $ symbol)
			value, err := strconv.ParseFloat(match[1], 64)
			if err == nil {
				values = append(values, value)
			}
		}
	}

	return values
}

// SumDollarValues sums an array of dollar values
func SumDollarValues(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}

	return math.Round(total*100) / 100
}

// CalculateItemCount calculates how many items the dollar amount could buy
// Always rounds up for fun!
func CalculateItemCount(total float64, pricePerItem float64) int {
	if pricePerItem <= 0 {
		return 0
	}

	// Calculate count and round up
	count := math.Ceil(total / pricePerItem)
	return int(count)
}

// FormatResponse creates a fun response message with the item count
func FormatResponse(count int, itemName string) string {
	// Handle pluralization
	item := itemName
	if count != 1 && !strings.HasSuffix(strings.ToLower(itemName), "s") {
		item = itemName + "s"
	}

	return "That's nearly " + strconv.Itoa(count) + " " + item + "!"
}
