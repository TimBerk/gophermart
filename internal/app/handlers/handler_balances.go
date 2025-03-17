package handlers

import (
	model "TimBerk/gophermart/internal/app/models/balance"
	"TimBerk/gophermart/pkg/responses"
	"TimBerk/gophermart/pkg/validators"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	action := "GetBalance"
	userID, ok := validators.ValidateAuthorization(w, r, action)
	if !ok {
		return
	}

	logFields := initLogFields(logrus.Fields{"action": action, "user": userID})

	var errMessage string
	balance, err := h.store.GetBalance(h.ctx, userID)
	if err != nil {
		errMessage = "failed to find balance"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusUnauthorized)
		return
	}

	jsonRecord, err := easyjson.Marshal(balance)
	if err != nil {
		errMessage = "failed to parse balance"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRecord)

}

func (h *Handler) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	action := "WithdrawBalance"

	userID, ok := validators.ValidateAuthorization(w, r, action)
	if !ok {
		return
	}

	logFields := initLogFields(logrus.Fields{"action": action, "user": userID})

	var errMessage string
	balance, err := h.store.GetBalance(h.ctx, userID)
	if err != nil {
		errMessage = "failed to get balance"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusUnauthorized)
		return
	}
	if balance.Current <= 0.00 {
		errMessage = "failed to use balance: it's empty"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusPaymentRequired)
		return
	}

	var requestData model.WithdrawnRequest
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		errMessage = "failed to parse request data"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	err = requestData.Validate()
	if err != nil {
		errMessage = "failed to validate request data"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	if balance.Current-float64(requestData.Sum) < 0.00 {
		errMessage = "failed to use balance: it's less than sum"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusPaymentRequired)
		return
	}

	order, err := h.store.GetOrder(h.ctx, requestData.Number)
	if err != nil || order.UserID != userID {
		errMessage = "failed to find order"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusUnprocessableEntity)
		return
	}

	// Begin a new transaction
	tx, err := h.store.BeginTx(h.ctx)
	if err != nil {
		errMessage = "failed to update order"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(h.ctx); rbErr != nil {
				logFields.WithField("error", rbErr).Error("failed to rollback transaction")
			}
		}
	}()

	err = h.store.UpdateOrderBalance(h.ctx, tx, requestData.Number, requestData.Sum)
	if err != nil {
		errMessage = "failed to update order"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	// Возможно перенести логику в триггер
	err = h.store.UpdateBalance(h.ctx, tx, userID, requestData.Sum)
	if err != nil {
		errMessage = "failed to update user balance"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(h.ctx); err != nil {
		errMessage = "failed to commit transaction"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
	}

}

func (h *Handler) GetWithdraw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var errMessage string
	action := "GetWithdraw"

	userID, ok := validators.ValidateAuthorization(w, r, action)
	if !ok {
		return
	}

	logFields := initLogFields(logrus.Fields{"action": action, "user": userID})

	records, err := h.store.GetOrderWithdrawals(h.ctx, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		errMessage = "failed to find orders"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}
	if len(records) == 0 {
		logFields.Info("Not found user orders")
		responses.WriteJSONEmpty(w, http.StatusNoContent)
		return
	}

	jsonRecords, err := easyjson.Marshal(records)
	if err != nil {
		errMessage = "failed to parse orders"
		logFields.WithField("error", err).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRecords)
}
