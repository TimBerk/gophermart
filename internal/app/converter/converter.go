package converter

import (
	model "TimBerk/gophermart/internal/app/models/order"
	"TimBerk/gophermart/internal/app/store"
)

func OrderDBToOrderAPI(record store.OrderRecord) model.OrderDetailResponse {
	item := model.OrderDetailResponse{
		Number: record.Order,
		Status: string(record.Status),
	}

	if record.Accrual.Valid {
		item.Accrual = &record.Accrual.Float64
	} else {
		item.Accrual = nil
	}

	return item
}
