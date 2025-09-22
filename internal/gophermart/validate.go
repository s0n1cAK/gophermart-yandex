package gophermart

import (
	"fmt"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/models"
)

func ValidateUser(user models.User) error {
	if user.Password == "" || user.Login == "" {
		return domain.MakeError(fmt.Errorf("login and password must be non-empty"), domain.ErrInvalidPayload)
	}

	if len(user.Login) <= 3 {
		return domain.MakeError(fmt.Errorf("login must be longer than 3 characters"), domain.ErrInvalidPayload)
	}

	if len(user.Password) <= 3 {
		return domain.MakeError(fmt.Errorf("password must be longer than 3 characters"), domain.ErrInvalidPayload)
	}
	return nil
}
