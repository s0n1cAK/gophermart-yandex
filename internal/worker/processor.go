package worker

import (
	"context"
	"math/rand"
	"time"
	"yandex-diplom/internal/job"
	"yandex-diplom/internal/job/jobs"
	"yandex-diplom/internal/models"

	"go.uber.org/zap"
)

type Workers struct {
	ctx    context.Context
	logger *zap.Logger
	svc    job.Service
	jobCh  chan job.Job
}

func InitWorkers(ctx context.Context, workerCount int, svc job.Service, jobCh chan job.Job) Workers {
	for i := range workerCount {
		go worker(ctx, i, svc, jobCh)
	}
	return Workers{ctx: ctx, logger: svc.GetLogger(), svc: svc, jobCh: jobCh}
}

func worker(ctx context.Context, id int, svc job.Service, in <-chan job.Job) {
	logger := svc.GetLogger()
	logger.Debug("Starting worker", zap.Int("wid", id))
	for {
		select {
		case <-ctx.Done():
			logger.Debug("Stopping worker", zap.Int("wid", id))
			return
		case job := <-in:
			err := job.Process(ctx, svc, logger)
			if err != nil {
				logger.Warn("Job processing failed", zap.Int("wid", id), zap.Error(err))
			}
		}
	}
}

type OrderConfig struct {
	BatchSize               int
	FetchNewInterval        time.Duration
	FetchProccesingInterval time.Duration
}

type BalanceConfig struct {
	FetchInterval time.Duration
}

func (w Workers) StartOrderProcessor(cfg OrderConfig) {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}

	if cfg.FetchNewInterval <= 0 {
		cfg.FetchNewInterval = 5 * time.Second
	}

	if cfg.FetchProccesingInterval <= 0 {
		cfg.FetchProccesingInterval = 3 * time.Second
	}

	w.logger.Info("Order processor config", zap.Int("BatchSize", cfg.BatchSize), zap.Duration("FetchNewInterval", cfg.FetchNewInterval), zap.Duration("FetchProccesingInterval", cfg.FetchProccesingInterval))

	throttler := job.NewThrottler()

	go func() {
		ticker := time.NewTicker(cfg.FetchNewInterval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-ticker.C:
				orders, err := w.svc.FetchNewOrders(w.ctx, cfg.BatchSize)
				if err != nil {
					w.logger.Warn("[order-processor] failed to get new orders", zap.Error(err))
					jitterSleep(cfg.FetchNewInterval)
					continue
				}
				putOrdersInChan(w, orders, throttler)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(cfg.FetchProccesingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-ticker.C:
				orders, err := w.svc.FetchProccesingOrders(w.ctx, cfg.BatchSize)
				if err != nil {
					w.logger.Warn("[order-processor] failed to get processing orders", zap.Error(err))
					jitterSleep(cfg.FetchProccesingInterval)
					continue
				}
				putOrdersInChan(w, orders, throttler)
			}
		}
	}()

}

func jitterSleep(base time.Duration) {
	j := time.Duration(rand.Int63n(int64(base / 2)))
	time.Sleep(base/2 + j)
}

func putOrdersInChan(w Workers, orders []models.Order, throttler *job.Throttler) {
	for _, o := range orders {
		select {
		case <-w.ctx.Done():
			return
		case w.jobCh <- &jobs.OrderJob{Order: o, Throttler: throttler}:
		default:
			w.logger.Warn("[order-processor] job channel full, skipping order", zap.String("order", o.Number))
		}
	}
}

func (w Workers) StartBalanceProcessor(cfg BalanceConfig) {
	if cfg.FetchInterval <= 0 {
		cfg.FetchInterval = 30 * time.Second
	}

	w.logger.Info("Balance processor config", zap.Duration("FetchInterval", cfg.FetchInterval))

	go func() {
		ticker := time.NewTicker(cfg.FetchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-ticker.C:
				select {
				case <-w.ctx.Done():
					return
				case w.jobCh <- &jobs.BalanceJob{}:
				default:
					w.logger.Warn("[balance-processor] job channel full, skipping job")
					continue
				}
			}
		}
	}()
}
