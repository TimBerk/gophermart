package rate_limit

import (
	"TimBerk/gophermart/internal/app/middlewares/auth"
	"TimBerk/gophermart/pkg/responses"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[int64]*rate.Limiter
	mu       sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[int64]*rate.Limiter),
	}
}

func (u *RateLimiter) GetLimiter(userID int64) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	if limiter, exists := u.limiters[userID]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Every(time.Second), 10)
	u.limiters[userID] = limiter
	return limiter
}

func RateLimit(rateLimiter *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/orders/{number}" {
				next.ServeHTTP(w, r) // Skip authentication
				return
			}

			var errMessage string

			rCtx := r.Context()
			userID, ok := rCtx.Value(auth.UserIDKey).(int64)
			if !ok || userID == 0 {
				errMessage = "User not authorized"
				logrus.WithFields(logrus.Fields{"action": "M.RateLimit", "user": userID, "error": ok}).Error(errMessage)
				responses.WriteJSONError(w, "User not authorized", http.StatusUnauthorized)
				return
			}

			limiter := rateLimiter.GetLimiter(userID)

			if !limiter.Allow() {
				errMessage = "Too many requests"
				logrus.WithFields(logrus.Fields{"action": "M.RateLimit", "user": userID, "error": ok}).Error(errMessage)
				//responses.WriteJSONError(w, errMessage, http.StatusTooManyRequests)
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
