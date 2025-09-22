package auth

import (
	"os"
	"time"
	"yandex-diplom/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

const TTL = time.Hour * 24

func CreateJWTToken(userID uint64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(TTL).Unix(),
	})

	tokenString, err := token.SignedString([]byte(GetSecret()))
	if err != nil {
		return "", domain.MakeError(err, domain.ErrInternal)
	}

	return tokenString, nil
}

func ParseJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return GetSecret(), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	return claims, nil
}

func GetSecret() []byte {
	secret, exists := os.LookupEnv("SECRET")

	if !exists || secret == "" {
		secret = "defaultsecret"
	}

	return []byte(secret)
}
