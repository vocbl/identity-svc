package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

type TokenGenerator func() (string, string)

func NewTokenGenerator(length int) TokenGenerator {
	return func() (string, string) {
		secret := make([]byte, length/2)
		rand.Read(secret)
		token := hex.EncodeToString(secret)

		return token, HashToken(token)
	}
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
