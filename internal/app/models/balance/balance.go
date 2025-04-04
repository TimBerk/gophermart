package balance

import (
	"TimBerk/gophermart/pkg/validators"
	"fmt"
	"time"
)

//go:generate easyjson -all -snake_case balance.go

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

//easyjson:json
type WithdrawnRequest struct {
	Number string  `json:"order"`
	Sum    float64 `json:"sum"`
}

//easyjson:json
type WithdrawnResponse struct {
	Number    string    `json:"order"`
	Sum       float64   `json:"sum"`
	CreatedAt time.Time `json:"processed_at"`
}

//easyjson:json
type WithdrawnList []WithdrawnResponse

func (w *WithdrawnRequest) Validate() error {
	err := validators.ValidateOrderNumber(w.Number)
	if err != nil {
		return err
	}

	if w.Sum < 0 {
		return fmt.Errorf("sum must be grater or equal 0")
	}
	return nil
}
