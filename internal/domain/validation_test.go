package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidatePassword(t *testing.T) {
	t.Run("returns true when password meets all requirements", func(t *testing.T) {
		assert.True(t, ValidatePassword("Pass12"))
	})

	t.Run("returns false when password is too short", func(t *testing.T) {
		assert.False(t, ValidatePassword("P1as"))
	})

	t.Run("returns false when password is missing uppercase", func(t *testing.T) {
		assert.False(t, ValidatePassword("pass123"))
	})

	t.Run("returns false when password is missing a digit", func(t *testing.T) {
		assert.False(t, ValidatePassword("Password"))
	})

	t.Run("returns false when password is only digits", func(t *testing.T) {
		assert.False(t, ValidatePassword("12345678"))
	})

	t.Run("returns false when password is only uppercase", func(t *testing.T) {
		assert.False(t, ValidatePassword("PASSWORDS"))
	})

	t.Run("returns false when password is empty", func(t *testing.T) {
		assert.False(t, ValidatePassword(""))
	})

	t.Run("returns true with special characters and requirements", func(t *testing.T) {
		assert.True(t, ValidatePassword("!@#P1as"))
	})

	t.Run("returns true with non-ASCII uppercase", func(t *testing.T) {
		assert.True(t, ValidatePassword("Öaaaa1"))
	})
}

func TestValidateEmail(t *testing.T) {
	t.Run("returns true when email is valid", func(t *testing.T) {
		assert.True(t, ValidateEmail("user@example.com"))
	})

	t.Run("returns false when email is missing @", func(t *testing.T) {
		assert.False(t, ValidateEmail("userexample.com"))
	})

	t.Run("returns false when email has extra text in brackets", func(t *testing.T) {
		assert.False(t, ValidateEmail("John Doe <john@example.com>"))
	})

	t.Run("returns false when email is empty", func(t *testing.T) {
		assert.False(t, ValidateEmail(""))
	})
}

func TestValidatePasswordHash(t *testing.T) {
	t.Parallel()
	t.Run("returns true when bcrypt hash is 60 chars", func(t *testing.T) {
		// Typical bcrypt length is 60
		hash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgNo3G1Xpux1pD.34rXIDuS7p3u."
		assert.NoError(t, ValidatePasswordHash(hash))
	})

	t.Run("returns false when hash length is incorrect", func(t *testing.T) {
		assert.Error(t, ValidatePasswordHash("short_hash"))
		assert.Error(t, ValidatePasswordHash(string(make([]byte, 61))))
	})
}

func TestValidateTokenHash(t *testing.T) {
	t.Parallel()
	t.Run("returns true when SHA256 hex hash is 64 chars", func(t *testing.T) {
		assert.NoError(t, ValidateTokenHash(TestValidTokenHash))
	})

	t.Run("returns false when token hash length is incorrect", func(t *testing.T) {
		assert.Error(t, ValidateTokenHash("too-short"))
	})
}

func TestValidateToken(t *testing.T) {
	t.Parallel()
	t.Run("returns true when token is exactly 64 chars", func(t *testing.T) {
		assert.NoError(t, ValidateToken(TestValidToken))
	})

	t.Run("returns false when token length is incorrect", func(t *testing.T) {
		assert.Error(t, ValidateToken("invalid_length_token"))
	})
}

func TestValidateUserVerificationSessionDuration(t *testing.T) {
	t.Parallel()
	t.Run("returns true for valid range (1m to 24h)", func(t *testing.T) {
		assert.NoError(t, ValidateUserVerificationSessionDuration(time.Minute))
		assert.NoError(t, ValidateUserVerificationSessionDuration(time.Hour))
		assert.NoError(t, ValidateUserVerificationSessionDuration(24*time.Hour))
	})

	t.Run("returns false for values outside range", func(t *testing.T) {
		assert.Error(t, ValidateUserVerificationSessionDuration(59*time.Second))
		assert.Error(t, ValidateUserVerificationSessionDuration(25*time.Hour))
	})
}

func TestValidatePasswordResetSessionDuration(t *testing.T) {
	t.Parallel()
	t.Run("returns true for valid range (1m to 10m)", func(t *testing.T) {
		assert.NoError(t, ValidatePasswordResetSessionDuration(time.Minute))
		assert.NoError(t, ValidatePasswordResetSessionDuration(10*time.Minute))
	})

	t.Run("returns false for values outside range", func(t *testing.T) {
		assert.Error(t, ValidatePasswordResetSessionDuration(59*time.Second))
		assert.Error(t, ValidatePasswordResetSessionDuration(11*time.Minute))
	})
}
