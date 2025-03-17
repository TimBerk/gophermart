package handlers

import (
	"TimBerk/gophermart/internal/app/models/auth"
	"TimBerk/gophermart/internal/app/settings/config"
	"TimBerk/gophermart/pkg/responses"
	"TimBerk/gophermart/pkg/secure"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type JWTRecord struct {
	Username string `json:"username"`
	UserId   int64  `json:"id"`
	jwt.RegisteredClaims
}

func generateToken(cfg *config.Config, userData auth.RequestData, user_id int64) (time.Time, string, error) {
	durationTime := time.Duration(cfg.ExpireJWT) * time.Minute
	expirationTime := time.Now().Add(durationTime)
	claims := &JWTRecord{
		Username: userData.Username,
		UserId:   user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	generatedToken, err := token.SignedString(cfg.KeyJWT)
	return expirationTime, generatedToken, err
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var userData auth.RequestData
	var errMessage string

	action := "Register"

	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		errMessage = "failed to parse request data"
		logrus.WithFields(logrus.Fields{"action": action, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusBadRequest)
		return
	}

	err = userData.Validate()
	if err != nil {
		errMessage = "failed to validate request data"
		logrus.WithFields(logrus.Fields{"action": action, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusBadRequest)
		return
	}

	user_id, err := h.store.CheckUser(h.ctx, userData.Username)
	if err != nil {
		errMessage = "failed to find user"
		logrus.WithFields(logrus.Fields{"action": action, "user": userData.Username, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}
	if user_id != 0 {
		errMessage = "user was registered"
		logrus.WithFields(logrus.Fields{"action": action, "user": userData.Username, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusConflict)
		return
	}

	hashedPassword, err := secure.HashPassword(userData.Password)
	if err != nil {
		errMessage = "failed to prepare password"
		logrus.WithFields(logrus.Fields{"action": action, "user": userData.Username, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}
	user_id, err = h.store.AddUser(h.ctx, userData.Username, hashedPassword)
	if err != nil {
		errMessage = "failed to register user"
		logrus.WithFields(logrus.Fields{"action": action, "user": userData.Username, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	expirationTime, tokenString, err := generateToken(h.cfg, userData, user_id)
	if err != nil {
		errMessage = "failed to generate token"
		logrus.WithFields(logrus.Fields{"action": action, "user": userData.Username, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	responses.WriteJSONToken(w, tokenString)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var userData auth.RequestData
	var errMessage string

	action := "Login"

	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		errMessage = "failed to parse request data"
		logrus.WithFields(logrus.Fields{"action": action, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusBadRequest)
		return
	}

	err = userData.Validate()
	if err != nil {
		errMessage = "failed to validate request data"
		logrus.WithFields(logrus.Fields{"action": action, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUser(h.ctx, userData.Username)
	if err != nil {
		errMessage = "failed to find user"
		logrus.WithFields(logrus.Fields{"action": action, "user": userData.Username, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	if !secure.CheckPasswordHash(userData.Password, user.PasswordHash) {
		errMessage = "incorrect pair username and password"
		logrus.WithFields(logrus.Fields{"action": action, "user": userData.Username, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusUnauthorized)
		return
	}

	expirationTime, tokenString, err := generateToken(h.cfg, userData, user.ID)
	if err != nil {
		errMessage = "failed to generate token"
		logrus.WithFields(logrus.Fields{"action": action, "user": userData.Username, "error": err}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	responses.WriteJSONToken(w, tokenString)
}
