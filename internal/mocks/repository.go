package mocks

import (
	"context"
	"yandex-diplom/internal/models"

	"github.com/stretchr/testify/mock"
)

type Repository struct {
	mock.Mock
}

func (m *Repository) RegisterUser(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *Repository) ValidateUser(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *Repository) CreateOrder(ctx context.Context, userLogin string, order models.Order) error {
	args := m.Called(ctx, userLogin, order)
	return args.Error(0)
}

func (m *Repository) GetUserByLogin(ctx context.Context, userLogin string) (models.User, error) {
	args := m.Called(ctx, userLogin)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *Repository) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *Repository) CheckUser(ctx context.Context, login string) (bool, error) {
	args := m.Called(ctx, login)
	return args.Bool(0), args.Error(1)
}

func (m *Repository) CheckOrder(ctx context.Context, user string, order models.Order) error {
	args := m.Called(ctx, user, order)
	return args.Error(0)
}

func (m *Repository) GetOrders(ctx context.Context, user string) ([]models.Order, error) {
	args := m.Called(ctx, user)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *Repository) UpdateOrderProcessed(ctx context.Context, order string, points float64) error {
	args := m.Called(ctx, order, points)
	return args.Error(0)
}

func (m *Repository) UpdateOrderInvalid(ctx context.Context, order string) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *Repository) FetchNewOrders(ctx context.Context, limit int) ([]models.Order, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *Repository) FetchProccesingOrders(ctx context.Context, limit int) ([]models.Order, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *Repository) UpdateBalanceEntries(ctx context.Context, order models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *Repository) UpdateMissingBalanceEntries(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *Repository) UpdateBalance(ctx context.Context, userID uint64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *Repository) GetBalance(ctx context.Context, userID uint64) (models.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(models.Balance), args.Error(1)
}

func (m *Repository) UpdateWithdrawlEntries(ctx context.Context, userID uint64, withdraw models.Withdrawal) error {
	args := m.Called(ctx, userID, withdraw)
	return args.Error(0)
}

func (m *Repository) GetWithdrawls(ctx context.Context, userID uint64) ([]models.Withdrawal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Withdrawal), args.Error(1)
}
