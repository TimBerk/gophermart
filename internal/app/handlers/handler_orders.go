package handlers

import (
	"TimBerk/gophermart/internal/app/converter"
	"TimBerk/gophermart/pkg/responses"
	"TimBerk/gophermart/pkg/validators"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var errMessage string

	action := "CreateOrder"
	userID, ok := validators.ValidateAuthorization(w, r, action)
	if !ok {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.WriteJSONError(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	orderNumber := string(body)
	logFields := initLogFields(logrus.Fields{"action": action, "user": userID, "order": orderNumber})

	err = validators.ValidateOrderNumber(orderNumber)
	if err != nil {
		errMessage = "failed to validate order number"
		logFields.WithField("error", err).Warning(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusUnprocessableEntity)
		return
	}

	order, err := h.store.GetOrder(h.ctx, orderNumber)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		errMessage = "failed to find order"
		logFields.WithField("error", err).Warning(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return

	} else if errors.Is(err, pgx.ErrNoRows) {
		err := h.store.AddOrder(h.ctx, userID, orderNumber)
		if err != nil {
			errMessage = "failed to create order"
			logFields.WithField("error", err).Warning(errMessage)
			responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		return
	}

	if order.UserID != userID {
		errMessage = "failed to check order: it was uploaded another user"
		logFields.WithField("error", err).Warning(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusConflict)
		return
	}

	logFields.Info("order was uploaded")
	w.WriteHeader(http.StatusOK)

}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	action := "GetOrders"
	userID, ok := validators.ValidateAuthorization(w, r, action)
	if !ok {
		return
	}

	logFields := initLogFields(logrus.Fields{"action": action, "user": userID})

	var errMessage string
	records, err := h.store.GetOrderList(h.ctx, userID)
	if err != nil {
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

	logFields.Info("Return list orders")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRecords)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	action := "GetOrder"
	userID, ok := validators.ValidateAuthorization(w, r, action)
	if !ok {
		return
	}

	orderNumber := chi.URLParam(r, "number")
	logFields := initLogFields(logrus.Fields{"action": action, "user": userID, "order": orderNumber})

	var errMessage string
	err := validators.ValidateOrderNumber(orderNumber)
	if err != nil {
		errMessage = "failed to validate order number"
		logFields.WithField("error", err).Warning(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusUnprocessableEntity)
		return
	}

	order, err := h.store.GetOrder(h.ctx, orderNumber)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		errMessage = "failed to find order"
		logFields.WithField("error", err).Warning(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return

	} else if errors.Is(err, pgx.ErrNoRows) {
		logFields.Info("Not found user order")
		responses.WriteJSONEmpty(w, http.StatusNoContent)
		return
	}

	apiItem := converter.OrderDbToOrderAPI(order)
	jsonRecord, err := easyjson.Marshal(apiItem)
	if err != nil {
		errMessage = "failed to parse order"
		logFields.WithField("error", err).Warning(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRecord)
}
