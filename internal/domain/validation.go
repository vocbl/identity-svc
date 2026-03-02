package domain

import (
	"fmt"
	"time"
)

const (
	TokenLenght = 64
)

// --- Existing Validations ---

// --- Token & Session Validations ---

func validateAccessTokenDuration(duration time.Duration) error {
	// Access tokens are stateless and should be short-lived
	min, max := 1*time.Minute, 2*time.Hour
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid access token duration: expected between %v and %v, got %v", min, max, duration)
}

func validateRefreshTokenDuration(duration time.Duration) error {
	// Refresh tokens are stateful and can live longer
	min, max := 1*time.Hour, 365*24*time.Hour
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid refresh token duration: expected between %v and %v, got %v", min, max, duration)
}

// --- Helper Validations ---

func validateTokenHash(tokenHash string) error {
	const expected = 64
	actual := len(tokenHash)
	if actual == expected {
		return nil
	}
	return fmt.Errorf("invalid token hash length: expected %d, got %d", expected, actual)
}

func validateToken(token string) error {
	const expected = 64
	actual := len(token)
	if actual == expected {
		return nil
	}
	return fmt.Errorf("invalid token length: expected %d, got %d", expected, actual)
}
