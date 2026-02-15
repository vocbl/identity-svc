package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vocbl/users-svc/internal/shared/security"
)

func TestNewPasswordChangeSession(t *testing.T) {
	t.Parallel()

	t.Run("returns error when inputs are invalid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			email    string
			duration time.Duration
			wantErr  error
		}{
			{
				name:     "returns error when email is invalid",
				email:    "not-an-email",
				duration: time.Hour,
				wantErr:  ErrInvalidEmail,
			},
			{
				name:     "returns error when duration is too short",
				email:    "test@example.com",
				duration: time.Second * 30,
				wantErr:  ErrInvalidSessionDuration,
			},
			{
				name:     "returns error when duration is too long",
				email:    "test@example.com",
				duration: time.Minute*15 + 1,
				wantErr:  ErrInvalidSessionDuration,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				session, err := NewPasswordChangeSession(tc.email, tc.duration)
				assert.Nil(t, session)
				assert.ErrorIs(t, err, tc.wantErr)
			})
		}
	})

	t.Run("returns session when inputs are valid", func(t *testing.T) {
		t.Parallel()
		email := "valid@example.com"
		duration := time.Minute * 10

		session, err := NewPasswordChangeSession(email, duration)

		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, email, session.Email)
		assert.WithinDuration(t, time.Now().UTC().Add(duration), session.ExpiresAt, time.Second)
		assert.Empty(t, session.TokenHash)
	})
}

func TestPasswordChangeSession_Enrich(t *testing.T) {
	t.Parallel()

	t.Run("returns error when token hash length is invalid", func(t *testing.T) {
		t.Parallel()
		session := &PasswordResetSession{}

		err := session.Enrich("too-short")

		assert.ErrorIs(t, err, ErrInvalidTokenHash)
		assert.Empty(t, session.TokenHash)
	})

	t.Run("returns nil and sets hash when token hash is valid", func(t *testing.T) {
		t.Parallel()
		session := &PasswordResetSession{}
		validHash := security.TestTokenHash(TokenLenght)

		err := session.Enrich(validHash)

		assert.NoError(t, err)
		assert.Equal(t, validHash, session.TokenHash)
	})
}
