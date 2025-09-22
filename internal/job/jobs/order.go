package jobs

import (
	"context"
	"errors"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/job"
	"yandex-diplom/internal/models"

	"go.uber.org/zap"
)

type OrderJob struct {
	Order     models.Order
	Throttler *job.Throttler
}

func (j *OrderJob) Process(ctx context.Context, svc job.Service, logger *zap.Logger) error {
	if j.Throttler.IsPaused() {
		logger.Debug("[OrderJob] throttled, skipping", zap.String("order", j.Order.Number))
		return nil
	}

	ext, err := svc.GetOrderFromAccurual(ctx, j.Order.Number)
	if err != nil {
		if e := new(domain.TooManyRequestsError); errors.As(err, &e) {
			logger.Info("[OrderJob] 429 received", zap.String("order", j.Order.Number))
			j.Throttler.Pause(e.RetryAfter)
			return nil
		}
		if errors.Is(err, domain.ErrNoContent) {
			logger.Debug("[OrderJob] not created in accurual", zap.String("order", j.Order.Number))
			return nil
		}
		return err
	}

	switch ext.Status {
	case "REGISTERED", "PROCESSING":
		return nil
	case "INVALID":
		return svc.UpdateOrderInvalid(ctx, j.Order)
	case "PROCESSED":
		err = svc.UpdateOrderProcessed(ctx, j.Order, ext.Accrual)
		if err != nil {
			return err
		}
		err = svc.UpdateBalanceEntries(ctx, j.Order)
		if err != nil {
			return err
		}
		return nil
	default:
		logger.Warn("Unknown status", zap.String("status", ext.Status))
	}

	return nil
}
