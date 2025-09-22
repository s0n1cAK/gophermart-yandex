package job

import (
	"context"
	"yandex-diplom/internal/models"

	"go.uber.org/zap"
)

type Job interface {
	Process(ctx context.Context, svc Service, logger *zap.Logger) error
}

type Service interface {
	UpdateOrderProcessed(ctx context.Context, order models.Order, points float64) error
	UpdateOrderInvalid(ctx context.Context, order models.Order) error
	FetchNewOrders(ctx context.Context, limit int) ([]models.Order, error)
	FetchProccesingOrders(ctx context.Context, limit int) ([]models.Order, error)
	UpdateBalanceEntries(ctx context.Context, order models.Order) error
	UpdateMissingBalanceEntries(ctx context.Context) error
	GetOrderFromAccurual(ctx context.Context, number string) (models.Order, error)
	GetLogger() *zap.Logger
}
