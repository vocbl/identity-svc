package domain

import (
	"database/sql"
	"testing"
	"time"

	ulidutil "github.com/vocbl/shared/utils/ulid"
	"github.com/vocbl/users-svc/internal/shared/security"
)

const (
	// Cfg
	TestValidUserVerificationMaxAttempts     = 5
	TestValidUserVerificationDuration        = time.Minute * 15
	TestValidUserVerificationRestartDuration = time.Minute
	TestValidUserVerificationCleanUpDuration = time.Minute * 40
	TestValidPasswordChangeDuration          = time.Minute * 15

	TestInvalidUserVerificationMaxAttempts     = 1
	TestInvalidUserVerificationDuration        = time.Minute
	TestInvalidUserVerificationRestartDuration = time.Second
	TestInvalidUserVerificationCleanUpDuration = time.Minute
	TestInvalidPasswordChangeDuration          = time.Minute

	// User
	TestValidPassword     = "TestPassword123"
	TestValidPasswordHash = "$2a$10$8K1p/a0DX1.A8At9.S8ObeS8ObeS8ObeS8ObeS8ObeS8ObeS8ObeS"
	TestValidEmail        = "valid@email.test"
	TestValidFirstName    = "John"
	TestValidLastName     = "Doe"
	TestValidUsername     = "TestJohn"
	TestValidToken        = "this_is_a_very_long_valid_token_string_that_passes_validation_64"
	TestValidTokenHash    = "8b2330a84d4117b43a9b1399479b12f60233f278d6556e9c3e414c2b9a7b7746"

	TestInvalidPassword     = "short"                // Fails length/complexity
	TestInvalidPasswordHash = "invalid-hash"         // Fails length/complexity
	TestInvalidEmail        = "not-an-email"         // Fails regex
	TestInvalidFirstName    = ""                     // Fails required check
	TestInvalidLastName     = ""                     // Fails required check
	TestInvalidUsername     = ""                     // Fails length (if min > 1)
	TestInvalidToken        = "too-short-token"      // Fails domain.ValidateToken
	TestInvalidTokenHash    = "not-a-hex-hash-12345" // Fails DB/Logic lookup
)

func NewValidTestUserVerificationSession(t *testing.T) UserVerificationSession {
	return UserVerificationSession{
		SessionID: ulidutil.NewULID(),
		TokenHash: security.TestTokenHash(TokenLenght),
		Email:     TestValidEmail,
		PasswordHash: sql.NullString{
			String: security.TestPasswordHash(t),
			Valid:  true,
		},
		Creds: UserCreds{
			FirstName: TestValidFirstName,
			LastName:  TestValidLastName,
			Username:  TestValidUsername,
		},
		Duration:  TestValidUserVerificationDuration,
		CreatedAt: time.Now().Add(-time.Minute * 3),
	}
}

func TestValidVerificationPolicy() VerificationPolicy {
	return VerificationPolicy{
		MaxAttempts:     TestValidUserVerificationMaxAttempts,
		Duration:        TestValidUserVerificationDuration,
		RestartDuration: TestValidUserVerificationRestartDuration,
		CleanUpDuration: TestValidUserVerificationCleanUpDuration,
	}
}
