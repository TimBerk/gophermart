package auth

import (
	"TimBerk/gophermart/internal/app/settings/config"
	"TimBerk/gophermart/pkg/responses"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type contextKey string

const (
	UsernameKey contextKey = "username"
	UserIDKey   contextKey = "userID"
)

type JWTRecord struct {
	Username string `json:"username"`
	UserID   int64  `json:"id"`
	jwt.RegisteredClaims
}

func getToken(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	bearerPrefix := "Bearer "
	if authHeader != "" && strings.HasPrefix(authHeader, bearerPrefix) {
		return strings.TrimPrefix(authHeader, bearerPrefix), true
	}

	cookie, err := r.Cookie("token")
	if err == nil {
		return cookie.Value, true
	}

	return "", false
}

func Authentication(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/user/register" || r.URL.Path == "/api/user/login" {
				next.ServeHTTP(w, r) // Skip authentication
				return
			}

			var errMessage string

			tokenString, ok := getToken(r)
			if !ok || tokenString == "" {
				errMessage = "User not authorized"
				logrus.WithFields(logrus.Fields{"action": "M.Authentication"}).Error(errMessage)
				responses.WriteJSONError(w, errMessage, http.StatusUnauthorized)
				return
			}

			claims := &JWTRecord{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return cfg.KeyJWT, nil
			})

			if err != nil {
				errMessage = "Failed parse token"
				logrus.WithFields(logrus.Fields{"action": "M.Authentication", "error": err}).Error(errMessage)
				responses.WriteJSONError(w, errMessage, http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				errMessage = "Invalid token"
				logrus.WithFields(logrus.Fields{"action": "M.Authentication", "error": err}).Error(errMessage)
				responses.WriteJSONError(w, errMessage, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
