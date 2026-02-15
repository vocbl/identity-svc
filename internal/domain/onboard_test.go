package domain

import (
	"database/sql"
	"testing"
	"time"

	"github.com/oklog/ulid"
	"github.com/stretchr/testify/assert"
	ulidtest "github.com/vocbl/shared/test/ulid"
	ulidutil "github.com/vocbl/shared/utils/ulid"
	"github.com/vocbl/users-svc/internal/shared/security"
)

func TestUserCreds_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input UserCreds
		errs  []error
	}{
		{
			name: "returns nil when creds are valid",
			input: UserCreds{
				FirstName: "John",
				LastName:  "Doe",
				Username:  "johndoe123",
			},
		},
		{
			name: "returns slice of one error when first name is invalid",
			input: UserCreds{
				LastName: "Doe",
				Username: "johndoe123",
			},
			errs: []error{ErrInvalidFirstName},
		},
		{
			name:  "returns slice of three errors when creds are blank",
			input: UserCreds{},
			errs:  []error{ErrInvalidFirstName, ErrInvalidLastName, ErrInvalidUsername},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			errs := tc.input.Validate()
			for _, testErr := range tc.errs {
				assert.Contains(t, errs, testErr)
			}
		})
	}
}

func TestUserVerificationSession_Prepare(t *testing.T) {
	t.Parallel()

	var session UserVerificationSession
	tokenHash := security.TestTokenHash(TokenLenght)
	dur := 10 * time.Minute
	restartableSince := time.Now()

	session.Enrich(tokenHash, dur, restartableSince)
	assert.Equal(t, tokenHash, session.TokenHash)
	assert.Equal(t, dur, session.Duration)
	assert.Equal(t, restartableSince, session.RestartableSince)
	assert.NotEqual(t, ulid.ULID{}, session.SessionID)
}

func TestUserVerificationSession_IsRestartable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		restartableSince time.Time
		expected         bool
	}{
		{
			name:             "returns false when restart duration has not passed",
			restartableSince: time.Now().Add(time.Minute),
			expected:         false,
		},
		{
			name:             "returns true when restart duration has passed",
			restartableSince: time.Now(),
			expected:         true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			session := UserVerificationSession{
				RestartableSince: tc.restartableSince,
			}
			assert.Equal(t, tc.expected, session.IsRestartable())
		})
	}
}
func TestUserVerificationSession_Validate(t *testing.T) {
	t.Parallel()

	validCreds := UserCreds{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
	}

	tests := []struct {
		name        string
		session     UserVerificationSession
		password    string
		expectedLen int
		errs        []error
	}{
		{
			name: "returns nil when all fields are valid",
			session: UserVerificationSession{
				Email: "test@example.com",
				Creds: validCreds,
			},
			password:    TestValidPassword,
			expectedLen: 0,
		},
		{
			name: "returns slice of one error when email is invalid",
			session: UserVerificationSession{
				Email: "invalid-email",
				Creds: validCreds,
			},
			password:    TestValidPassword,
			expectedLen: 1,
			errs:        []error{ErrInvalidEmail},
		},
		{
			name: "returns slice of one error when password is too short",
			session: UserVerificationSession{
				Email: "test@example.com",
				Creds: validCreds,
			},
			password:    "P1as",
			expectedLen: 1,
			errs:        []error{ErrInvalidPassword},
		},
		{
			name: "returns slice of one error when password lacks uppercase",
			session: UserVerificationSession{
				Email: "test@example.com",
				Creds: validCreds,
			},
			password:    "password123",
			expectedLen: 1,
			errs:        []error{ErrInvalidPassword},
		},
		{
			name: "returns slice of one error when password lacks a digit",
			session: UserVerificationSession{
				Email: "test@example.com",
				Creds: validCreds,
			},
			password:    "Password",
			expectedLen: 1,
			errs:        []error{ErrInvalidPassword},
		},
		{
			name: "returns slice of one error when password is empty",
			session: UserVerificationSession{
				Email: "test@example.com",
				Creds: validCreds,
			},
			password:    "",
			expectedLen: 1,
			errs:        []error{ErrInvalidPassword},
		},
		{
			name: "returns slice of three errors when nested credentials are blank",
			session: UserVerificationSession{
				Email: "test@example.com",
				Creds: UserCreds{},
			},
			password:    TestValidPassword,
			expectedLen: 3,
			errs:        []error{ErrInvalidFirstName, ErrInvalidLastName, ErrInvalidUsername},
		},
		{
			name:        "returns slice of five errors when everything is invalid",
			session:     UserVerificationSession{},
			expectedLen: 5,
			errs:        []error{ErrInvalidFirstName, ErrInvalidLastName, ErrInvalidUsername, ErrInvalidEmail, ErrInvalidPassword},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errs := tc.session.Validate(tc.password)

			assert.Len(t, errs, tc.expectedLen)
			for _, testedErr := range tc.errs {
				assert.Contains(t, errs, testedErr)
			}
		})
	}
}

func TestUserVerificationSession_ToUser(t *testing.T) {
	t.Parallel()

	session := UserVerificationSession{
		Email: "tester@example.com",
		Creds: UserCreds{
			FirstName: "Jane",
			LastName:  "Doe",
			Username:  "janedoe",
		},

		PasswordHash: sql.NullString{
			String: security.TestPasswordHash(t),
			Valid:  true,
		},
	}

	userID := ulidutil.NewULID()
	user := session.ToUser(userID)

	assert.NotNil(t, user)
	assert.Equal(t, session.Email, user.Email)
	assert.Equal(t, session.Creds, user.Creds)
	assert.Equal(t, session.PasswordHash, user.PasswordHash)
	ulidtest.NotNil(t, user.ID)
}

func TestVerificationPolicy_Validate(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for valid config", testVerificationCfg_Validate_Success)
	t.Run("returns error for invalid duration", testVerificationCfg_Validate_InvalidDuration)
	t.Run("returns error for invalid restart duration", testVerificationCfg_Validate_InvalidRestartDuration)
	t.Run("returns error for invalid cleanup duration", testVerificationCfg_Validate_InvalidCleanupDuration)
	t.Run("returns error for invalid max attempts", testVerificationCfg_Validate_InvalidMaxAttempts)
}

func testVerificationCfg_Validate_Success(t *testing.T) {
	t.Parallel()
	cfg := TestValidVerificationPolicy()
	assert.NoError(t, cfg.Validate())
}

func testVerificationCfg_Validate_InvalidDuration(t *testing.T) {
	t.Parallel()
	cfg := TestValidVerificationPolicy()
	cfg.Duration = TestInvalidUserVerificationDuration
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration")
}

func testVerificationCfg_Validate_InvalidRestartDuration(t *testing.T) {
	t.Parallel()
	cfg := TestValidVerificationPolicy()
	cfg.RestartDuration = TestInvalidUserVerificationRestartDuration
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid restart duration")
}

func testVerificationCfg_Validate_InvalidCleanupDuration(t *testing.T) {
	t.Parallel()
	cfg := TestValidVerificationPolicy()
	cfg.CleanUpDuration = TestInvalidUserVerificationCleanUpDuration
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid clean up duration")
}

func testVerificationCfg_Validate_InvalidMaxAttempts(t *testing.T) {
	t.Parallel()
	cfg := TestValidVerificationPolicy()
	cfg.MaxAttempts = TestInvalidUserVerificationMaxAttempts
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid max attempts")
}

//
//
//
//
