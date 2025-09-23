package gophermart

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
	"yandex-diplom/internal/domain"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestMart(serverURL string) *Mart {
	u, _ := url.Parse(serverURL)
	logger, _ := zap.NewDevelopment()
	return &Mart{
		log:      logger,
		accurual: u,
		client:   &http.Client{},
	}
}

func TestGetOrderFromAccurual_OK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/orders/123", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"order":"123","status":"PROCESSED","accrual":42.5}`))
	}))
	defer ts.Close()

	m := newTestMart(ts.URL)

	order, err := m.GetOrderFromAccurual(context.Background(), "123")
	require.NoError(t, err)
	require.Equal(t, "123", order.Number)
	require.Equal(t, "PROCESSED", order.Status)
	require.Equal(t, 42.5, order.Accrual)
}

func TestGetOrderFromAccurual_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer ts.Close()

	m := newTestMart(ts.URL)

	_, err := m.GetOrderFromAccurual(context.Background(), "123")
	require.Error(t, err)
	require.True(t, domain.GetAppErr(err) != nil)
}

func TestGetOrderFromAccurual_TooManyRequests_DefaultRetry(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	m := newTestMart(ts.URL)

	_, err := m.GetOrderFromAccurual(context.Background(), "123")
	require.Error(t, err)

	var derr domain.Error
	require.True(t, errors.As(err, &derr))

	var tooMany *domain.TooManyRequestsError
	require.True(t, errors.As(err, &tooMany))
	require.Equal(t, time.Minute, tooMany.RetryAfter)
}

func TestGetOrderFromAccurual_TooManyRequests_CustomRetry(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	m := newTestMart(ts.URL)

	_, err := m.GetOrderFromAccurual(context.Background(), "123")
	require.Error(t, err)

	var tooMany *domain.TooManyRequestsError
	require.True(t, errors.As(err, &tooMany))
	require.Equal(t, 5*time.Second, tooMany.RetryAfter)
}

func TestGetOrderFromAccurual_NoContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	m := newTestMart(ts.URL)

	_, err := m.GetOrderFromAccurual(context.Background(), "123")
	require.Error(t, err)
	require.True(t, domain.GetAppErr(err) != nil)
	require.True(t, errors.Is(err, domain.ErrNoContent))
}

func TestGetOrderFromAccurual_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	m := newTestMart(ts.URL)

	_, err := m.GetOrderFromAccurual(context.Background(), "123")
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInternal))
}
