package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/lib"
	"yandex-diplom/internal/luhn"
	"yandex-diplom/internal/models"
)

func bindUserFromJSON(r *http.Request) (models.User, error) {
	const op = "httpx.BindUserFromJSON"

	r.Body = http.MaxBytesReader(nil, r.Body, 5<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u models.User
	if err := dec.Decode(&u); err != nil {
		return models.User{}, domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInvalidPayload,
		)
	}

	if u.Login == "" || len(u.Password) == 0 {
		return models.User{}, domain.MakeError(
			lib.StandardError(op, errors.New("empty login or password")),
			domain.ErrInvalidPayload,
		)
	}

	u.Login = strings.TrimSpace(u.Login)
	u.Password = strings.TrimSpace(u.Password)

	return u, nil
}

func bindOrderFromPlain(r *http.Request) (models.Order, error) {
	const op = "httpx.bindOrderFromPlain"

	r.Body = http.MaxBytesReader(nil, r.Body, 5<<20)
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return models.Order{}, domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInvalidPayload,
		)
	}

	s := strings.TrimSpace(string(bodyBytes))
	if len(s) == 0 {
		return models.Order{}, domain.MakeError(
			lib.StandardError(op, errors.New("payload is empty")),
			domain.ErrInvalidPayload,
		)
	}

	o := models.Order{Number: s}

	if !luhn.Valid(o.Number) {
		return models.Order{}, domain.MakeError(
			lib.StandardError(op, errors.New("number is invalid")),
			domain.ErrUnprocessableOrder,
		)
	}

	return o, nil
}

func bindWithdrawlFromJSON(r *http.Request) (models.Withdrawal, error) {
	const op = "httpx.BindUserFromJSON"

	r.Body = http.MaxBytesReader(nil, r.Body, 5<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var w models.Withdrawal
	if err := dec.Decode(&w); err != nil {
		return models.Withdrawal{}, domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInvalidPayload,
		)
	}

	if w.Sum <= 0 {
		return models.Withdrawal{}, domain.MakeError(
			lib.StandardError(op, errors.New("sum can't be lower that 1")),
			domain.ErrInvalidPayload,
		)
	}

	return w, nil
}
