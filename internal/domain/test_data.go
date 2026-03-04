package domain

import (
	"time"

	ulidutil "github.com/vocbl/shared/utils/ulid"
)

const (
	// Cfg
	TestValidUserVerificationMaxAttempts     = 5
	TestValidUsernameGenerationMaxAttempts   = 3
	TestValidUserVerificationExpireDuration  = time.Minute * 15
	TestValidUserVerificationRestartDuration = time.Minute
	TestValidUserVerificationCleanUpDuration = time.Minute * 40
	TestValidPasswordChangeDuration          = time.Minute * 15

	TestInvalidUserVerificationMaxAttempts     = 1
	TestInvalidUsernameGenerationMaxAttempts   = 1
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

	TestInvalidPassword     = "short"        // Fails length/complexity
	TestInvalidPasswordHash = "invalid-hash" // Fails length/complexity
	TestInvalidEmail        = "not-an-email" // Fails regex
	TestInvalidID           = "not-a-ulid"
	TestInvalidFirstName    = ""                     // Fails required check
	TestInvalidLastName     = ""                     // Fails required check
	TestInvalidUsername     = ""                     // Fails length (if min > 1)
	TestInvalidToken        = "too-short-token"      // Fails domain.ValidateToken
	TestInvalidTokenHash    = "not-a-hex-hash-12345" // Fails DB/Logic lookup
)

var (
	TestValidID                    = ulidutil.NewString()
	TestValidVerificationSessionID = newVerificationSessionID()
	TestValidUserID                = newUserID()
)

func NewValidTestUserVerificationSession() *UserVerificationSession {
	return &UserVerificationSession{
		id:    VerificationSessionID{newID()},
		email: TestValidEmail,
		Creds: UserCreds{
			firstName: TestValidFirstName,
			lastName:  TestValidLastName,
			username:  TestValidUsername,
		},
		createdAt: time.Now().Add(-time.Minute * 3),
	}
}

func NewValidTestUser() *User {
	return &User{
		id:    TestValidUserID,
		email: TestValidEmail,
		Creds: UserCreds{
			firstName: TestValidFirstName,
			lastName:  TestValidLastName,
			username:  TestValidUsername,
		},
		createdAt: time.Now().Add(-time.Minute * 3),
	}
}

func NewValidTestActiveUserVerificationSession() *UserVerificationSession {

	s := NewValidTestUserVerificationSession()

	s.attemptCount++

	expiresAt := time.Now().UTC().Add(TestValidUserVerificationExpireDuration)
	s.expiresAt = &expiresAt

	restartableSince := time.Now().UTC().Add(TestValidUserVerificationRestartDuration)
	s.restartableSince = &restartableSince

	passwordHashPointer := PasswordHash(TestValidPasswordHash)
	s.passwordHash = &passwordHashPointer

	tokenHashPointer := TokenHash(TestValidTokenHash)
	s.tokenHash = &tokenHashPointer

	return s
}

func NewValidTestInactiveUserVerificationSession() *UserVerificationSession {
	s := NewValidTestActiveUserVerificationSession()

	past := time.Now().UTC().Add(-1 * time.Minute)
	s.restartableSince = &past
	s.expiresAt = &past

	return s
}
