package gophermart

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"yandex-diplom/internal/auth"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/mocks"
	"yandex-diplom/internal/models"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRegister_Success(t *testing.T) {
	repo := new(mocks.Repository)
	logger, _ := zap.NewDevelopment()
	accURL, _ := url.Parse("http://localhost:8080")

	mart := New(repo, logger, "test", accURL)

	user := models.User{Login: "testuser", Password: "password123"}

	repo.On("CheckUser", mock.Anything, user.Login).Return(false, nil)
	repo.On("RegisterUser", mock.Anything, user).Return(nil)
	repo.On("GetUserByLogin", mock.Anything, user.Login).
		Return(models.User{ID: 1, Login: "testuser"}, nil)

	cookie, err := mart.Register(context.Background(), user)
	require.NoError(t, err)
	require.Equal(t, "Authorization", cookie.Name)
	require.NotEmpty(t, cookie.Value)

	claims, err := auth.ParseJWT(cookie.Value)
	require.NoError(t, err)
	require.Equal(t, float64(1), claims["userID"])
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	repo := new(mocks.Repository)
	logger, _ := zap.NewDevelopment()
	accURL, _ := url.Parse("http://localhost:8080")

	mart := New(repo, logger, "test", accURL)

	user := models.User{Login: "testuser", Password: "password123"}

	repo.On("CheckUser", mock.Anything, user.Login).Return(true, nil)

	_, err := mart.Register(context.Background(), user)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrLoginAlreadyTaken))
}

func TestPutWithdrawl_NotEnoughBalance(t *testing.T) {
	repo := new(mocks.Repository)
	logger, _ := zap.NewDevelopment()
	accURL, _ := url.Parse("http://localhost:8080")

	mart := New(repo, logger, "test", accURL)

	user := models.User{ID: 1, Login: "testuser"}
	withdraw := models.Withdrawal{Order: "79927398713", Sum: 100}

	repo.On("GetBalance", mock.Anything, user.ID).
		Return(models.Balance{Current: 50}, nil)

	err := mart.PutWithdrawl(context.Background(), user, withdraw)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrPaymentRequired))
}

func TestGetWithdrawals_NoContent(t *testing.T) {
	repo := new(mocks.Repository)
	logger, _ := zap.NewDevelopment()
	accURL, _ := url.Parse("http://localhost:8080")

	mart := New(repo, logger, "test", accURL)

	repo.On("GetWithdrawls", mock.Anything, uint64(1)).
		Return([]models.Withdrawal{}, nil)

	_, err := mart.GetWithdrawals(context.Background(), 1)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNoContent))
}
