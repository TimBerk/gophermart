package validators

import (
	"TimBerk/gophermart/internal/app/middlewares/auth"
	"TimBerk/gophermart/pkg/responses"
	"TimBerk/gophermart/pkg/secure"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func ValidateOrderNumber(number string) error {
	if number == "" {
		return fmt.Errorf("empty required value")
	}

	intNumber, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return fmt.Errorf("incorrect number")
	}

	if isValid := secure.CheckLuhn(intNumber); !isValid {
		return fmt.Errorf("invalid number")
	}
	return nil
}

func ValidateAuthorization(w http.ResponseWriter, r *http.Request, action string) (int64, bool) {
	ctx := r.Context()
	userID, ok := ctx.Value(auth.UserIDKey).(int64)
	if !ok {
		errMessage := "User is not authorized"
		logrus.WithFields(logrus.Fields{"action": action, "error": ok}).Error(errMessage)
		responses.WriteJSONError(w, errMessage, http.StatusUnauthorized)
	}
	return userID, ok
}
