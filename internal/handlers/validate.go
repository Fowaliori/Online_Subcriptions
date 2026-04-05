package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	errNegativePrice  = errors.New("цена должна быть >= 0")
	errEndBeforeStart = errors.New("конец периода раньше начала")
)

func parseSubscriptionID(id string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(id))
}

func parseUserID(s string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(s))
}

func validatePrice(p int) error {
	if p < 0 {
		return errNegativePrice
	}
	return nil
}

func validateSubscriptionMonthRange(start time.Time, end *time.Time) error {
	if end == nil {
		return nil
	}
	if end.Before(start) {
		return errEndBeforeStart
	}
	return nil
}
