package auth

import (
	"TimBerk/gophermart/internal/app/settings/config"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mockConfig(jwtKey string) *config.Config {
	return &config.Config{
		KeyJWT: []byte(jwtKey),
	}
}

func TestGetToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
		expectedOk    bool
	}{
		{
			name:          "valid bearer token",
			authHeader:    "Bearer valid.token.here",
			expectedToken: "valid.token.here",
			expectedOk:    true,
		},
		{
			name:          "empty header",
			authHeader:    "",
			expectedToken: "",
			expectedOk:    false,
		},
		{
			name:          "malformed header",
			authHeader:    "Bearervalid.token.here",
			expectedToken: "",
			expectedOk:    false,
		},
		{
			name:          "different auth type",
			authHeader:    "Basic dXNlcjpwYXNz",
			expectedToken: "",
			expectedOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			token, ok := getToken(req)
			assert.Equal(t, tt.expectedToken, token)
			assert.Equal(t, tt.expectedOk, ok)
		})
	}
}

func TestAuthenticationMiddleware(t *testing.T) {
	validKey := "secret-key"
	invalidKey := "wrong-key"

	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &JWTRecord{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	})
	validTokenString, _ := validToken.SignedString([]byte(validKey))

	tests := []struct {
		name            string
		token           string
		config          *config.Config
		expectedStatus  int
		shouldCallNext  bool
		expectedMessage string
	}{
		{
			name:           "Valid token",
			token:          validTokenString,
			config:         mockConfig(validKey),
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:            "Empty token",
			token:           "",
			config:          mockConfig(validKey),
			expectedStatus:  http.StatusUnauthorized,
			shouldCallNext:  false,
			expectedMessage: "User not authorized",
		},
		{
			name:            "Wrong signing key",
			token:           validTokenString,
			config:          mockConfig(invalidKey),
			expectedStatus:  http.StatusUnauthorized,
			shouldCallNext:  false,
			expectedMessage: "Failed parse token",
		},
		{
			name:            "No Authorization header",
			token:           "",
			config:          mockConfig(validKey),
			expectedStatus:  http.StatusUnauthorized,
			shouldCallNext:  false,
			expectedMessage: "User not authorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rr := httptest.NewRecorder()

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := Authentication(tt.config)
			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.shouldCallNext, nextCalled, "Next handler call mismatch")
			assert.Equal(t, tt.expectedStatus, rr.Code, "HTTP status code mismatch")
			if tt.expectedMessage != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedMessage, "Error message mismatch")
			}
		})
	}
}
