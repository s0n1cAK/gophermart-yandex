package gophermart

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"yandex-diplom/internal/domain"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestMartForWriteError(env string) *Mart {
	logger, _ := zap.NewDevelopment()
	return &Mart{
		log:         logger,
		Environment: env,
	}
}

func TestWriteError_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid payload", domain.ErrInvalidPayload, http.StatusBadRequest},
		{"user not found", domain.ErrUserNotFound, http.StatusBadRequest},
		{"login already taken", domain.ErrLoginAlreadyTaken, http.StatusConflict},
		{"order created by other", domain.ErrOrderCreatedByOtherUser, http.StatusConflict},
		{"invalid credentials", domain.ErrInvalidCredentials, http.StatusUnauthorized},
		{"unprocessable order", domain.ErrUnprocessableOrder, http.StatusUnprocessableEntity},
		{"order created by user", domain.ErrOrderCreatedByUser, http.StatusOK},
		{"no content", domain.ErrNoContent, http.StatusNoContent},
		{"payment required", domain.ErrPaymentRequired, http.StatusPaymentRequired},
		{"unknown error", errors.New("something"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			m := newTestMartForWriteError("test")

			m.WriteError(rec, tt.err)

			require.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
