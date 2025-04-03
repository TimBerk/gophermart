package responses

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

type JSONErrorResponse struct {
	Error string `json:"error"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	errorResponse := JSONErrorResponse{Error: message}

	logrus.WithFields(logrus.Fields{
		"header":     w.Header(),
		"message":    message,
		"statusCode": statusCode,
	}).Error("JSON Error")

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(errorResponse)
}

func WriteJSONEmpty(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
}

func WriteJSONToken(w http.ResponseWriter, token string) {
	w.WriteHeader(http.StatusOK)
	tokenResponse := TokenResponse{Token: token}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(tokenResponse)
}
