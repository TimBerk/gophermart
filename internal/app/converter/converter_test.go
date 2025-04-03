package converter

import (
	model "TimBerk/gophermart/internal/app/models/order"
	"TimBerk/gophermart/internal/app/store"
	"TimBerk/gophermart/pkg/utils"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrderDBToOrderAPI(t *testing.T) {
	testCases := map[string]struct {
		input    store.OrderRecord
		expected model.OrderDetailResponse
	}{
		"complete record with accrual": {
			input: store.OrderRecord{
				ID:      1,
				UserID:  100,
				Order:   "ORDER123",
				Status:  store.Processed,
				Accrual: sql.NullFloat64{Float64: 150.75, Valid: true},
			},
			expected: model.OrderDetailResponse{
				Number:  "ORDER123",
				Status:  "PROCESSED",
				Accrual: utils.PtrFloat64(150.75),
			},
		},
		"record without accrual": {
			input: store.OrderRecord{
				ID:      2,
				UserID:  100,
				Order:   "ORDER456",
				Status:  store.New,
				Accrual: sql.NullFloat64{Valid: false},
			},
			expected: model.OrderDetailResponse{
				Number:  "ORDER456",
				Status:  "NEW",
				Accrual: nil,
			},
		},
		"zero value record": {
			input: store.OrderRecord{
				ID:      0,
				UserID:  0,
				Order:   "",
				Status:  store.Status(""),
				Accrual: sql.NullFloat64{Valid: false},
			},
			expected: model.OrderDetailResponse{
				Number:  "",
				Status:  "",
				Accrual: nil,
			},
		},
		"record with zero accrual": {
			input: store.OrderRecord{
				ID:      3,
				UserID:  100,
				Order:   "ORDER789",
				Status:  store.Processing,
				Accrual: sql.NullFloat64{Float64: 0, Valid: true},
			},
			expected: model.OrderDetailResponse{
				Number:  "ORDER789",
				Status:  "PROCESSING",
				Accrual: utils.PtrFloat64(0),
			},
		},
		"record with negative accrual": {
			input: store.OrderRecord{
				ID:      4,
				UserID:  100,
				Order:   "ORDER987",
				Status:  store.Invalid,
				Accrual: sql.NullFloat64{Float64: -50.25, Valid: true},
			},
			expected: model.OrderDetailResponse{
				Number:  "ORDER987",
				Status:  "INVALID",
				Accrual: utils.PtrFloat64(-50.25),
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			result := OrderDBToOrderAPI(tc.input)

			assert.Equal(t, tc.expected, result)
			if tc.input.Accrual.Valid {
				assert.NotNil(t, result.Accrual)
				assert.Equal(t, tc.input.Accrual.Float64, *result.Accrual)
			} else {
				assert.Nil(t, result.Accrual)
			}
		})
	}
}
