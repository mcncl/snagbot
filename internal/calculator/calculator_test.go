package calculator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDollarValues(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []float64
	}{
		{
			name:     "No dollar values",
			text:     "This has no dollar values",
			expected: []float64{},
		},
		{
			name:     "Single dollar value",
			text:     "This costs $35",
			expected: []float64{35.0},
		},
		{
			name:     "Multiple dollar values",
			text:     "This costs $35 and that costs $50",
			expected: []float64{35.0, 50.0},
		},
		{
			name:     "Decimal values",
			text:     "This costs $35.50 and that costs $24.99",
			expected: []float64{35.50, 24.99},
		},
		{
			name:     "Values with text in between",
			text:     "The project costs $35 for setup and $20 per month",
			expected: []float64{35.0, 20.0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractDollarValues(test.text)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestSumDollarValues(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "Empty list",
			values:   []float64{},
			expected: 0,
		},
		{
			name:     "Single value",
			values:   []float64{35.0},
			expected: 35.0,
		},
		{
			name:     "Multiple values",
			values:   []float64{35.0, 50.0, 20.0},
			expected: 105.0,
		},
		{
			name:     "Decimal values",
			values:   []float64{35.50, 24.99},
			expected: 60.49,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SumDollarValues(test.values)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCalculateItemCount(t *testing.T) {
	tests := []struct {
		name         string
		total        float64
		pricePerItem float64
		expected     int
	}{
		{
			name:         "Exact division",
			total:        35.0,
			pricePerItem: 3.5,
			expected:     10,
		},
		{
			name:         "Round up",
			total:        36.0,
			pricePerItem: 3.5,
			expected:     11,
		},
		{
			name:         "Zero total",
			total:        0,
			pricePerItem: 3.5,
			expected:     0,
		},
		{
			name:         "Zero price",
			total:        35.0,
			pricePerItem: 0,
			expected:     0,
		},
		{
			name:         "Negative price",
			total:        35.0,
			pricePerItem: -1.0,
			expected:     0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CalculateItemCount(test.total, test.pricePerItem)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFormatResponse(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		itemName string
		expected string
	}{
		{
			name:     "Single item",
			count:    1,
			itemName: "Bunnings snag",
			expected: "That's nearly 1 Bunnings snag!",
		},
		{
			name:     "Multiple items",
			count:    10,
			itemName: "Bunnings snag",
			expected: "That's nearly 10 Bunnings snags!",
		},
		{
			name:     "Already plural",
			count:    10,
			itemName: "Bunnings snags",
			expected: "That's nearly 10 Bunnings snags!",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FormatResponse(test.count, test.itemName)
			assert.Equal(t, test.expected, result)
		})
	}
}
