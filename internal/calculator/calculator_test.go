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
		{
			name:     "Currency with no space",
			text:     "That'll be$35please",
			expected: []float64{35.0},
		},
		{
			name:     "Dollar sign at end of word",
			text:     "USD$35 and AUD$20",
			expected: []float64{35.0, 20.0},
		},
		{
			name:     "Multiple decimals (should only match valid currency format)",
			text:     "$35.50.25 should only match $35.50 once",
			expected: []float64{35.50},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ExtractDollarValues(test.text)
			if err != nil {
				t.Fatal(err)
			}
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
		{
			name:     "Rounding precision test",
			values:   []float64{0.1, 0.2},
			expected: 0.3, // Should be exactly 0.3 after rounding, not 0.30000000000000004
		},
		{
			name:     "Negative values (should still work)",
			values:   []float64{35.0, -15.0},
			expected: 20.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := SumDollarValues(test.values)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCalculateItemCount(t *testing.T) {
	// Split tests into valid and invalid cases
	validTests := []struct {
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
			name:         "Almost exact",
			total:        34.99,
			pricePerItem: 3.5,
			expected:     10,
		},
		{
			name:         "Zero total",
			total:        0,
			pricePerItem: 3.5,
			expected:     0,
		},
		{
			name:         "Small price, large total",
			total:        1000.0,
			pricePerItem: 0.01,
			expected:     100000,
		},
	}

	// Tests that should return errors
	invalidTests := []struct {
		name         string
		total        float64
		pricePerItem float64
	}{
		{
			name:         "Zero price",
			total:        35.0,
			pricePerItem: 0,
		},
		{
			name:         "Negative price",
			total:        35.0,
			pricePerItem: -1.0,
		},
		{
			name:         "Negative total",
			total:        -35.0,
			pricePerItem: 3.5,
		},
	}

	// Test valid inputs
	for _, test := range validTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := CalculateItemCount(test.total, test.pricePerItem)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}

	// Test invalid inputs that should return errors
	for _, test := range invalidTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CalculateItemCount(test.total, test.pricePerItem)
			assert.Error(t, err, "Expected error for invalid input")
		})
	}
}

func TestFormatResponse(t *testing.T) {
	tests := []struct {
		name            string
		count           int
		itemName        string
		isExactDivision bool
		expected        string
	}{
		{
			name:            "Zero items",
			count:           0,
			itemName:        "Bunnings snag",
			isExactDivision: true,
			expected:        "That wouldn't even buy a single Bunnings snag!",
		},
		{
			name:            "Single item (exact)",
			count:           1,
			itemName:        "Bunnings snag",
			isExactDivision: true,
			expected:        "That's 1 Bunnings snag!",
		},
		{
			name:            "Single item (not exact)",
			count:           1,
			itemName:        "Bunnings snag",
			isExactDivision: false,
			expected:        "That's nearly 1 Bunnings snag!",
		},
		{
			name:            "Multiple items (exact)",
			count:           10,
			itemName:        "Bunnings snag",
			isExactDivision: true,
			expected:        "That's 10 Bunnings snags!",
		},
		{
			name:            "Multiple items (not exact)",
			count:           10,
			itemName:        "Bunnings snag",
			isExactDivision: false,
			expected:        "That's nearly 10 Bunnings snags!",
		},
		{
			name:            "Already plural (exact)",
			count:           10,
			itemName:        "Bunnings snags",
			isExactDivision: true,
			expected:        "That's 10 Bunnings snags!",
		},
		{
			name:            "Already plural (not exact)",
			count:           10,
			itemName:        "Bunnings snags",
			isExactDivision: false,
			expected:        "That's nearly 10 Bunnings snags!",
		},
		{
			name:            "Custom item (exact)",
			count:           7,
			itemName:        "coffee",
			isExactDivision: true,
			expected:        "That's 7 coffees!",
		},
		{
			name:            "Custom item (not exact)",
			count:           7,
			itemName:        "coffee",
			isExactDivision: false,
			expected:        "That's nearly 7 coffees!",
		},
		{
			name:            "Custom item already plural (exact)",
			count:           7,
			itemName:        "coffees",
			isExactDivision: true,
			expected:        "That's 7 coffees!",
		},
		{
			name:            "Custom item already plural (not exact)",
			count:           7,
			itemName:        "coffees",
			isExactDivision: false,
			expected:        "That's nearly 7 coffees!",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FormatResponse(test.count, test.itemName, test.isExactDivision)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProcessMessage(t *testing.T) {
	// Split tests into valid and invalid cases
	validTests := []struct {
		name         string
		text         string
		pricePerItem float64
		expected     string
	}{
		{
			name:         "No dollar values",
			text:         "This has no dollar values",
			pricePerItem: 3.50,
			expected:     "",
		},
		{
			name:         "Single dollar value (exact division)",
			text:         "This costs $35",
			pricePerItem: 3.50,
			expected:     "That's 10 Bunnings snags!",
		},
		{
			name:         "Single dollar value (not exact division)",
			text:         "This costs $34",
			pricePerItem: 3.50,
			expected:     "That's nearly 10 Bunnings snags!",
		},
		{
			name:         "Multiple dollar values (exact division)",
			text:         "This costs $35 and that costs $35",
			pricePerItem: 3.50,
			expected:     "That's 10 Bunnings snags!",
		},
		{
			name:         "Multiple dollar values (not exact division)",
			text:         "This costs $35 and that costs $34",
			pricePerItem: 3.50,
			expected:     "That's nearly 20 Bunnings snags!",
		},
		{
			name:         "Custom price per item (exact division)",
			text:         "This costs $35",
			pricePerItem: 5.00,
			expected:     "That's 7 Bunnings snags!",
		},
		{
			name:         "Custom price per item (not exact division)",
			text:         "This costs $33",
			pricePerItem: 5.00,
			expected:     "That's nearly 7 Bunnings snags!",
		},
		{
			name:         "Very small amount (less than one item)",
			text:         "This costs $2",
			pricePerItem: 3.50,
			expected:     "That wouldn't even buy a single Bunnings snag!",
		},
	}

	// Tests that should return errors
	invalidTests := []struct {
		name         string
		text         string
		pricePerItem float64
	}{
		{
			name:         "Zero price per item",
			text:         "This costs $35",
			pricePerItem: 0,
		},
		{
			name:         "Negative price per item",
			text:         "This costs $35",
			pricePerItem: -1.0,
		},
	}

	// Test valid inputs
	for _, test := range validTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ProcessMessage(test.text, test.pricePerItem)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}

	// Test invalid inputs that should return errors
	for _, test := range invalidTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ProcessMessage(test.text, test.pricePerItem)
			assert.Error(t, err, "Expected error for invalid input")
		})
	}
}
