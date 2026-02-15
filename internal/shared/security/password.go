package security

import "golang.org/x/crypto/bcrypt"

type PasswordHasher func(password string) (string, error)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashedPassword), err
}

func VerifyPassword(password, hasedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hasedPassword), []byte(password)) == nil
}
