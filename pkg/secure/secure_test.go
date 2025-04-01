package secure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckLuhn(t *testing.T) {
	testCases := []struct {
		name     string
		number   int64
		expected bool
	}{
		{
			name:     "valid 1",
			number:   49927398716,
			expected: true,
		},
		{
			name:     "valid 2",
			number:   4532015112830366,
			expected: true,
		},
		{
			name:     "invalid 1",
			number:   49927398717,
			expected: false,
		},
		{
			name:     "invalid 2",
			number:   1234567812345678,
			expected: false,
		},
		{
			name:     "single digit valid",
			number:   7,
			expected: false,
		},
		{
			name:     "negative number",
			number:   -49927398716,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CheckLuhn(tc.number)
			assert.Equal(t, tc.expected, result, "Number: %d", tc.number)
		})
	}
}
