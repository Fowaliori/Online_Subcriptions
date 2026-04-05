package errs

import "errors"

var (
	ErrSubscriptionNotFound = errors.New("подписка не найдена")
	ErrServiceNotFound      = errors.New("сервис не найден")
)
