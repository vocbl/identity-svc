package domain

import (
	"time"
)

type PasswordResetSession struct {
	TokenHash string
	Email     string
	ExpiresAt time.Time
}

func NewPasswordChangeSession(email string, duration time.Duration) (*PasswordResetSession, error) {
	if !ValidateEmail(email) {
		return nil, ErrInvalidEmail
	}

	return &PasswordResetSession{
		Email:     email,
		ExpiresAt: time.Now().UTC().Add(duration),
	}, nil
}

func (s *PasswordResetSession) Enrich(tokenHash string) error {
	if len(tokenHash) != 64 {
		return ErrInvalidTokenHash
	}

	s.TokenHash = tokenHash
	return nil
}
