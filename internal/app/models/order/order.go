package order

import (
	"time"
)

//go:generate easyjson -all -snake_case order.go

type OrderDetailResponse struct {
	Number  string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

//easyjson:json
type OrderResponse struct {
	Number    string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   *float64  `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at"`
}

//easyjson:json
type OrderListResponse []OrderResponse

//easyjson:json
type OrderAccrualRegister struct {
	Number string `json:"order"`
}

//easyjson:json
type OrderAccrual struct {
	Number string `json:"order"`
	Status string `json:"status"`
}
