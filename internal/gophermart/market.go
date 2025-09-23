package gophermart

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"yandex-diplom/internal/auth"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/luhn"
	"yandex-diplom/internal/models"

	"go.uber.org/zap"
)

type User interface {
	Register(ctx context.Context, user models.User) (http.Cookie, error)
	Login(ctx context.Context, user models.User) (http.Cookie, error)
	PutOrder(ctx context.Context, login string, order models.Order) error
	GetOrders(ctx context.Context, login string) ([]models.Order, error)
	GetWithdrawals(ctx context.Context, userID uint64) ([]models.Withdrawal, error)
	UpdateOrderProcessed(ctx context.Context, order models.Order, points float64) error
	UpdateOrderInvalid(ctx context.Context, order models.Order) error
	FetchNewOrders(ctx context.Context, limit int) ([]models.Order, error)
	FetchProccesingOrders(ctx context.Context, limit int) ([]models.Order, error)
	UpdateBalanceEntries(ctx context.Context, order models.Order) error
	UpdateMissingBalanceEntries(ctx context.Context) error
	GetBalance(ctx context.Context, user models.User) (models.Balance, error)
	PutWithdrawl(ctx context.Context, user models.User, Withdrawal models.Withdrawal) error
}

type System interface {
	WriteError(w http.ResponseWriter, err error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
	GetOrderFromAccurual(ctx context.Context, number string) (models.Order, error)
	GetLogger() *zap.Logger
}

type Service interface {
	User
	System
}

type Reposiroty interface {
	RegisterUser(ctx context.Context, user models.User) error
	ValidateUser(ctx context.Context, user models.User) error
	CreateOrder(ctx context.Context, userLogin string, order models.Order) error
	GetUserByLogin(ctx context.Context, userLogin string) (models.User, error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
	CheckUser(ctx context.Context, login string) (bool, error)
	CheckOrder(ctx context.Context, user string, order models.Order) error
	GetOrders(ctx context.Context, user string) ([]models.Order, error)
	UpdateOrderProcessed(ctx context.Context, order string, points float64) error
	UpdateOrderInvalid(ctx context.Context, order string) error
	FetchNewOrders(ctx context.Context, limit int) ([]models.Order, error)
	FetchProccesingOrders(ctx context.Context, limit int) ([]models.Order, error)
	UpdateBalanceEntries(ctx context.Context, order models.Order) error
	UpdateMissingBalanceEntries(ctx context.Context) error
	UpdateBalance(ctx context.Context, userID uint64) error
	GetBalance(ctx context.Context, userID uint64) (models.Balance, error)
	UpdateWithdrawlEntries(ctx context.Context, userID uint64, withdraw models.Withdrawal) error
	GetWithdrawls(ctx context.Context, userID uint64) ([]models.Withdrawal, error)
}

type Mart struct {
	db          Reposiroty
	log         *zap.Logger
	Environment string
	accurual    *url.URL
	client      *http.Client
}

func New(db Reposiroty, logger *zap.Logger, env string, accural *url.URL) Service {
	return &Mart{db: db, log: logger, Environment: env, accurual: accural,
		client: &http.Client{Timeout: 10 * time.Second}}
}

func (m *Mart) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	op := "gophermart.GetUserByID"

	user, err := m.db.GetUserByID(ctx, id)
	if err != nil {
		return models.User{}, domain.Wrap(op, err)
	}

	return user, nil
}

func (m *Mart) Register(ctx context.Context, user models.User) (http.Cookie, error) {
	op := "gophermart.Register"

	if err := ValidateUser(user); err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	exist, err := m.db.CheckUser(ctx, user.Login)
	if err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	if exist {
		return http.Cookie{}, domain.Wrap(op, domain.MakeError(fmt.Errorf("user already exist"), domain.ErrLoginAlreadyTaken))
	}

	err = m.db.RegisterUser(ctx, user)
	if err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	dbUser, err := m.db.GetUserByLogin(ctx, user.Login)
	if err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	token, err := auth.CreateJWTToken(dbUser.ID)
	if err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(auth.TTL),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie, nil
}

func (m *Mart) Login(ctx context.Context, user models.User) (http.Cookie, error) {
	op := "gophermart.Login"

	if err := ValidateUser(user); err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	err := m.db.ValidateUser(ctx, user)
	if err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	dbUser, err := m.db.GetUserByLogin(ctx, user.Login)
	if err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	token, err := auth.CreateJWTToken(dbUser.ID)
	if err != nil {
		return http.Cookie{}, domain.Wrap(op, err)
	}

	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(auth.TTL),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie, nil
}

func (m *Mart) PutOrder(ctx context.Context, login string, order models.Order) error {
	op := "gophermart.PutOrder"

	valid := luhn.Valid(order.Number)
	if !valid {
		return domain.Wrap(op, domain.MakeError(fmt.Errorf("invalid order number"), domain.ErrInvalidPayload))
	}

	err := m.db.CheckOrder(ctx, login, order)
	if err != nil {
		return domain.Wrap(op, err)
	}

	err = m.db.CreateOrder(ctx, login, order)
	if err != nil {
		return domain.Wrap(op, err)
	}

	return nil
}

func (m *Mart) GetOrders(ctx context.Context, login string) ([]models.Order, error) {
	op := "gophermart.GetOrders"

	orders, err := m.db.GetOrders(ctx, login)
	if err != nil {
		return []models.Order{}, domain.Wrap(op, err)
	}

	return orders, nil
}

func (m *Mart) GetBalance(ctx context.Context, user models.User) (models.Balance, error) {
	op := "gophermart.GetBalance"

	balance, err := m.db.GetBalance(ctx, user.ID)
	if err != nil {
		return models.Balance{}, domain.Wrap(op, err)
	}

	return balance, nil
}

func (m *Mart) PutWithdrawl(ctx context.Context, user models.User, Withdrawal models.Withdrawal) error {
	op := "gophermart.PutWithdrawl"

	valid := luhn.Valid(Withdrawal.Order)
	if !valid {
		return domain.Wrap(op, domain.MakeError(fmt.Errorf("invalid order number"), domain.ErrUnprocessableOrder))
	}

	balance, err := m.db.GetBalance(ctx, user.ID)
	if err != nil {
		return domain.Wrap(op, err)
	}

	if balance.Current < Withdrawal.Sum {
		return domain.Wrap(op, domain.MakeError(fmt.Errorf("the current balance is lower than the amount indicated"), domain.ErrPaymentRequired))
	}

	err = m.db.UpdateWithdrawlEntries(ctx, user.ID, Withdrawal)
	if err != nil {
		return domain.Wrap(op, err)
	}

	return nil
}

func (m *Mart) GetWithdrawals(ctx context.Context, userID uint64) ([]models.Withdrawal, error) {
	op := "gophermart.GetWithdrawals"

	withdrawals, err := m.db.GetWithdrawls(ctx, userID)
	if err != nil {
		return []models.Withdrawal{}, domain.Wrap(op, err)
	}

	if len(withdrawals) <= 0 {
		return []models.Withdrawal{}, domain.Wrap(op, domain.MakeError(fmt.Errorf("there is no record of withdrawal"), domain.ErrNoContent))
	}

	return withdrawals, nil
}

func (m *Mart) GetLogger() *zap.Logger {
	return m.log
}

func (m *Mart) UpdateOrderInvalid(ctx context.Context, order models.Order) error {
	return m.db.UpdateOrderInvalid(ctx, order.Number)
}

func (m *Mart) UpdateOrderProcessed(ctx context.Context, order models.Order, points float64) error {
	return m.db.UpdateOrderProcessed(ctx, order.Number, points)
}

func (m *Mart) FetchNewOrders(ctx context.Context, limit int) ([]models.Order, error) {
	return m.db.FetchNewOrders(ctx, limit)
}

func (m *Mart) FetchProccesingOrders(ctx context.Context, limit int) ([]models.Order, error) {
	return m.db.FetchProccesingOrders(ctx, limit)
}

func (m *Mart) UpdateBalanceEntries(ctx context.Context, order models.Order) error {
	return m.db.UpdateBalanceEntries(ctx, order)
}

func (m *Mart) UpdateMissingBalanceEntries(ctx context.Context) error {
	return m.db.UpdateMissingBalanceEntries(ctx)
}
