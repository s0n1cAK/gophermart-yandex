package jobs

import (
	"context"
	"yandex-diplom/internal/job"

	"go.uber.org/zap"
)

type BalanceJob struct {
}

func (j *BalanceJob) Process(ctx context.Context, svc job.Service, logger *zap.Logger) error {
	err := svc.UpdateMissingBalanceEntries(ctx)
	if err != nil {
		return err
	}
	return nil
}
