package jobs

import (
	"context"
	"errors"
	"testing"
	"time"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/job"
	"yandex-diplom/internal/mocks"
	"yandex-diplom/internal/models"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderJob_Throttled(t *testing.T) {
	th := job.NewThrottler()
	th.Pause(time.Minute)

	j := OrderJob{Order: models.Order{Number: "1"}, Throttler: th}
	mockSvc := new(mocks.MockService)

	err := j.Process(context.Background(), mockSvc, mockSvc.GetLogger())
	require.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestOrderJob_TooManyRequests(t *testing.T) {
	th := job.NewThrottler()
	j := OrderJob{Order: models.Order{Number: "1"}, Throttler: th}
	mockSvc := new(mocks.MockService)

	mockSvc.On("GetOrderFromAccurual", mock.Anything, "1").
		Return(models.Order{}, domain.MakeError(errors.New("rate limit"), &domain.TooManyRequestsError{RetryAfter: time.Second}))

	err := j.Process(context.Background(), mockSvc, mockSvc.GetLogger())
	require.NoError(t, err)
	require.True(t, th.IsPaused())
}

func TestOrderJob_NoContent(t *testing.T) {
	th := job.NewThrottler()
	j := OrderJob{Order: models.Order{Number: "1"}, Throttler: th}
	mockSvc := new(mocks.MockService)

	mockSvc.On("GetOrderFromAccurual", mock.Anything, "1").
		Return(models.Order{}, domain.MakeError(errors.New("not found"), domain.ErrNoContent))

	err := j.Process(context.Background(), mockSvc, mockSvc.GetLogger())
	require.NoError(t, err)
}

func TestOrderJob_InvalidStatus(t *testing.T) {
	th := job.NewThrottler()
	j := OrderJob{Order: models.Order{Number: "1"}, Throttler: th}
	mockSvc := new(mocks.MockService)

	mockSvc.On("GetOrderFromAccurual", mock.Anything, "1").
		Return(models.Order{Number: "1", Status: "INVALID"}, nil)
	mockSvc.On("UpdateOrderInvalid", mock.Anything, mock.Anything).Return(nil)

	err := j.Process(context.Background(), mockSvc, mockSvc.GetLogger())
	require.NoError(t, err)
	mockSvc.AssertCalled(t, "UpdateOrderInvalid", mock.Anything, mock.Anything)
}

func TestOrderJob_Processed(t *testing.T) {
	th := job.NewThrottler()
	j := OrderJob{Order: models.Order{Number: "1"}, Throttler: th}
	mockSvc := new(mocks.MockService)

	mockSvc.On("GetOrderFromAccurual", mock.Anything, "1").
		Return(models.Order{Number: "1", Status: "PROCESSED", Accrual: 10}, nil)
	mockSvc.On("UpdateOrderProcessed", mock.Anything, mock.Anything, 10.0).Return(nil)
	mockSvc.On("UpdateBalanceEntries", mock.Anything, mock.Anything).Return(nil)

	err := j.Process(context.Background(), mockSvc, mockSvc.GetLogger())
	require.NoError(t, err)
}
