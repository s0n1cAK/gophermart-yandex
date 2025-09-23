package mocks

import (
	"context"
	"yandex-diplom/internal/models"

	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) UpdateOrderProcessed(ctx context.Context, order models.Order, points float64) error {
	args := m.Called(ctx, order, points)
	return args.Error(0)
}

func (m *MockService) UpdateOrderInvalid(ctx context.Context, order models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockService) FetchNewOrders(ctx context.Context, limit int) ([]models.Order, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockService) FetchProccesingOrders(ctx context.Context, limit int) ([]models.Order, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockService) UpdateBalanceEntries(ctx context.Context, order models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockService) UpdateMissingBalanceEntries(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockService) GetOrderFromAccurual(ctx context.Context, number string) (models.Order, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(models.Order), args.Error(1)
}

func (m *MockService) GetLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}
