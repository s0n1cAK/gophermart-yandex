package auth

import (
	"context"
	"net/http"

	"yandex-diplom/internal/models"
)

type contextKey string

const userKey = contextKey("user")

func GetUserFromContext(ctx context.Context) *models.User {
	val := ctx.Value(userKey)
	if user, ok := val.(*models.User); ok {
		return user
	}
	return nil
}

type UserProvider interface {
	GetUserByID(ctx context.Context, id int64) (models.User, error)
}

func Middleware(userProvider UserProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("Authorization")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := ParseJWT(cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			rawID, ok := claims["userID"]
			if !ok {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			floatID, ok := rawID.(float64)
			if !ok {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			userID := int64(floatID)

			user, err := userProvider.GetUserByID(r.Context(), userID)
			if err != nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
