package domain

import (
	"errors"
	"fmt"
	"time"
	"yandex-diplom/internal/lib"
)

var (
	ErrInternal                = errors.New("internal error")
	ErrServiceUnavailable      = errors.New("service unavailable")
	ErrInvalidPayload          = errors.New("invalid payload")
	ErrLoginAlreadyTaken       = errors.New("user with that login already created")
	ErrOrderAlreadyExists      = errors.New("order already created")
	ErrInvalidCredentials      = errors.New("invalid login or password")
	ErrUserNotFound            = errors.New("unknowe user")
	ErrUnprocessableOrder      = errors.New("unprocessable order number")
	ErrOrderCreatedByUser      = errors.New("order aldready created by user")
	ErrOrderCreatedByOtherUser = errors.New("order aldready created by other user")
	ErrNoContent               = errors.New("no content")
	ErrPaymentRequired         = errors.New("payment required")
)

type TooManyRequestsError struct {
	RetryAfter time.Duration
}

func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("429 Too Many Requests — retry after %s", e.RetryAfter)
}

type Error struct {
	AppErr error
	SrvErr error
}

func (e Error) Error() string {
	if e.SrvErr != nil {
		return e.SrvErr.Error()
	}

	if e.AppErr != nil {
		return e.AppErr.Error()
	}
	return "unknown error"
}

func (e Error) Unwrap() error {
	return e.SrvErr
}

func MakeError(apperr error, srverr error) Error {
	return Error{AppErr: apperr, SrvErr: srverr}
}

func Wrap(op string, err error) error {
	var derr Error
	if errors.As(err, &derr) {
		return Error{
			AppErr: lib.StandardError(op, derr.AppErr),
			SrvErr: derr.SrvErr,
		}
	}
	return Error{
		AppErr: lib.StandardError(op, err),
		SrvErr: ErrInternal,
	}
}

func GetAppErr(err error) error {
	var derr Error
	if errors.As(err, &derr) {
		return derr.AppErr
	}
	return nil
}
