package gophermart

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/models"
)

type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}

func (m *Mart) GetOrderFromAccurual(ctx context.Context, number string) (models.Order, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u := *m.accurual
	u.Path = "/api/orders/" + number

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return models.Order{}, domain.MakeError(
			fmt.Errorf("build request: %w", err), domain.ErrInternal)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return models.Order{}, domain.MakeError(
			fmt.Errorf("request.GetOrderFromAccurual request to accrual server failed: %w", err),
			domain.ErrInternal)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var ext struct {
			Order   string  `json:"order"`
			Status  string  `json:"status"`
			Accrual float32 `json:"accrual,omitempty"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&ext); err != nil {
			return models.Order{}, domain.MakeError(
				fmt.Errorf("request.GetOrderFromAccurual can't parse answer from accrual server: %w", err),
				domain.ErrInternal)
		}

		return models.Order{
			Number:  ext.Order,
			Status:  ext.Status,
			Accrual: float64(ext.Accrual),
		}, nil

	case http.StatusTooManyRequests:
		retry := time.Minute
		if val := resp.Header.Get("Retry-After"); val != "" {
			if seconds, err := strconv.Atoi(val); err == nil {
				retry = time.Duration(seconds) * time.Second
			}
		}
		return models.Order{}, domain.MakeError(
			fmt.Errorf("request.GetOrderFromAccurual too many requests to accrual"),
			&domain.TooManyRequestsError{RetryAfter: retry})

	case http.StatusNoContent:
		return models.Order{}, domain.MakeError(
			fmt.Errorf("request.GetOrderFromAccurual order doesn't exist at accrual"),
			domain.ErrNoContent)

	default:
		return models.Order{}, domain.MakeError(
			fmt.Errorf("request.GetOrderFromAccurual unexpected status %d", resp.StatusCode),
			domain.ErrInternal)
	}
}
