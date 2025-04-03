package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOrderNumber(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "valid order number 1",
			input:       "49927398716",
			expectedErr: "",
		},
		{
			name:        "valid order number 2",
			input:       "4532015112830366",
			expectedErr: "",
		},
		{
			name:        "valid order number 3",
			input:       "79927398713",
			expectedErr: "",
		},

		// Invalid Luhn numbers
		{
			name:        "invalid luhn number 1",
			input:       "49927398717",
			expectedErr: "invalid number",
		},
		{
			name:        "invalid luhn number 2",
			input:       "1234567812345678",
			expectedErr: "invalid number",
		},

		// Edge cases
		{
			name:        "empty string",
			input:       "",
			expectedErr: "empty required value",
		},
		{
			name:        "non-numeric string",
			input:       "ABCDEFG",
			expectedErr: "incorrect number",
		},
		{
			name:        "number with spaces",
			input:       " 49927398716 ",
			expectedErr: "incorrect number",
		},
		{
			name:        "number with special chars",
			input:       "4992-7398-716",
			expectedErr: "incorrect number",
		},
		{
			name:        "single digit",
			input:       "7",
			expectedErr: "invalid number", // Luhn check will fail
		},
		{
			name:        "very long number",
			input:       "1234567890123456789012345678901234567890",
			expectedErr: "incorrect number", // Too long for int64
		},
		{
			name:        "negative number",
			input:       "-49927398716",
			expectedErr: "invalid number",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOrderNumber(tc.input)

			if tc.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}
