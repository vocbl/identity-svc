package domain

import (
	"fmt"
	"net/mail"
	"time"
	"unicode"
)

const (
	TokenLenght = 64
)

// --- Existing Validations ---

func ValidateEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func ValidatePassword(password string) bool {
	if len(password) < 6 {
		return false
	}

	var hasUpper, hasDigit bool
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if hasUpper && hasDigit {
			return true
		}
	}
	return false
}

func ValidatePasswordHash(passwordHash string) error {
	const expected = 60
	actual := len(passwordHash)
	if actual == expected {
		return nil
	}
	return fmt.Errorf("invalid password hash length: expected %d, got %d", expected, actual)
}

// --- Token & Session Validations ---

func ValidateAccessTokenDuration(duration time.Duration) error {
	// Access tokens are stateless and should be short-lived
	min, max := 1*time.Minute, 2*time.Hour
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid access token duration: expected between %v and %v, got %v", min, max, duration)
}

func ValidateRefreshTokenDuration(duration time.Duration) error {
	// Refresh tokens are stateful and can live longer
	min, max := 1*time.Hour, 365*24*time.Hour
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid refresh token duration: expected between %v and %v, got %v", min, max, duration)
}

func ValidateUserVerificationSessionDuration(duration time.Duration) error {
	min, max := 10*time.Minute, 24*time.Hour
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid user verification session duration: expected between %v and %v, got %v", min, max, duration)
}

func ValidateUserVerificationSessionRestartDuration(duration time.Duration) error {
	min, max := time.Minute, 10*time.Minute
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid user verification session restart duration: expected between %v and %v, got %v", min, max, duration)
}

func ValidateUserVerificationSessionCleanUpDuration(duration time.Duration) error {
	min, max := 30*time.Minute, 24*time.Hour
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid user verification session cleanup duration: expected between %v and %v, got %v", min, max, duration)
}

func ValidatePasswordResetSessionDuration(duration time.Duration) error {
	min, max := time.Minute, 10*time.Minute
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid password reset session duration: expected between %v and %v, got %v", min, max, duration)
}

// --- Helper Validations ---

func ValidateUserVerificationSessionMaxAttempts(attempts int) error {
	min, max := 2, 9
	if attempts >= min && attempts <= max {
		return nil
	}
	return fmt.Errorf("invalid user verification session max attempts: expected between %d and %d, got %d", min, max, attempts)
}

func ValidateTokenHash(tokenHash string) error {
	const expected = 64
	actual := len(tokenHash)
	if actual == expected {
		return nil
	}
	return fmt.Errorf("invalid token hash length: expected %d, got %d", expected, actual)
}

func ValidateToken(token string) error {
	const expected = 64
	actual := len(token)
	if actual == expected {
		return nil
	}
	return fmt.Errorf("invalid token length: expected %d, got %d", expected, actual)
}
