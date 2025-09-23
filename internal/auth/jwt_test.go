package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAndParseJWT(t *testing.T) {
	os.Setenv("SECRET", "testsecret")
	defer os.Unsetenv("SECRET")

	token, err := CreateJWTToken(42)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseJWT(token)
	assert.NoError(t, err)
	assert.Equal(t, float64(42), claims["userID"])
}

func TestParseJWT_InvalidSignature(t *testing.T) {
	os.Setenv("SECRET", "secret1")
	token, err := CreateJWTToken(99)
	assert.NoError(t, err)

	os.Setenv("SECRET", "secret2")

	claims, err := ParseJWT(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestGetSecret_Default(t *testing.T) {
	os.Unsetenv("SECRET")

	secret := GetSecret()
	assert.Equal(t, []byte("defaultsecret"), secret)
}
