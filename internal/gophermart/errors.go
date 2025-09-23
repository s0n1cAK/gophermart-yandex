package gophermart

import (
	"errors"
	"net/http"
	"yandex-diplom/internal/domain"
)

func (m *Mart) WriteError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrInvalidPayload),
		errors.Is(err, domain.ErrUserNotFound):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrLoginAlreadyTaken),
		errors.Is(err, domain.ErrOrderCreatedByOtherUser):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrInvalidCredentials):
		status = http.StatusUnauthorized
	case errors.Is(err, domain.ErrUnprocessableOrder):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrOrderCreatedByUser):
		status = http.StatusOK
	case errors.Is(err, domain.ErrNoContent):
		status = http.StatusNoContent
	case errors.Is(err, domain.ErrPaymentRequired):
		status = http.StatusPaymentRequired
	}

	http.Error(w, "", status)

	var derr domain.Error
	errors.As(err, &derr)

	if m.Environment == "test" || m.Environment == "dev" {
		if derr.AppErr != nil {
			m.log.Info(derr.AppErr.Error())
		}
	}
}
