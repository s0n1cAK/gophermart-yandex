package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"yandex-diplom/internal/models"

	"github.com/stretchr/testify/assert"
)

type mockUserProvider struct {
	user *models.User
	err  error
}

func (m *mockUserProvider) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	if m.err != nil {
		return models.User{}, m.err
	}
	return *m.user, nil
}

func TestMiddleware_NoCookie(t *testing.T) {
	provider := &mockUserProvider{}
	middleware := Middleware(provider)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("handler should not be called")
	})).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_ValidToken(t *testing.T) {
	token, err := CreateJWTToken(123)
	assert.NoError(t, err)

	provider := &mockUserProvider{
		user: &models.User{ID: 123, Login: "TestUser"},
	}

	middleware := Middleware(provider)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "Authorization", Value: token})
	w := httptest.NewRecorder()

	var gotUser *models.User
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = GetUserFromContext(r.Context())
	})).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, gotUser)
	assert.Equal(t, uint64(123), gotUser.ID)
	assert.Equal(t, "TestUser", gotUser.Login)
}
