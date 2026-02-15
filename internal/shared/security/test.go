package security

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPasswordHash(t *testing.T) string {
	password := make([]byte, 15)
	_, err := rand.Read(password)
	require.NoError(t, err)

	passwordHash, err := HashPassword(string(password))
	require.NoError(t, err)

	return passwordHash
}

func TestTokenHash(lenth int) string {
	_, tokenHash := NewTokenGenerator(lenth)()
	return tokenHash
}

func TestToken(lenth int) string {
	token, _ := NewTokenGenerator(lenth)()
	return token
}
