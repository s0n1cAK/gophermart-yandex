package httpx

import (
	"encoding/json"
	"net/http"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/lib"
	"yandex-diplom/internal/models"
)

func responseJSONFromOrders(w http.ResponseWriter, o []models.Order) error {
	const op = "httpx.bindJSONFromOrders"

	payload, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInternal,
		)
	}

	if _, err := w.Write(payload); err != nil {
		return domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInternal,
		)
	}
	return nil
}

func responseJSONBalance(w http.ResponseWriter, b models.Balance) error {
	const op = "httpx.bindJSONFromOrders"

	payload, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInternal,
		)
	}

	if _, err := w.Write(payload); err != nil {
		return domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInternal,
		)
	}
	return nil
}

func responseJSONWithdrawals(w http.ResponseWriter, p []models.Withdrawal) error {
	const op = "httpx.responseJSONWithdrawals"

	payload, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInternal,
		)
	}

	if _, err := w.Write(payload); err != nil {
		return domain.MakeError(
			lib.StandardError(op, err),
			domain.ErrInternal,
		)
	}
	return nil
}
