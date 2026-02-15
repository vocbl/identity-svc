package app

import (
	"errors"

	errorutil "github.com/vocbl/shared/errors"
)

var (
	// Operations errors
	ErrCreateVerificationSessionOperation           = errorutil.Operation("failed to create the user")
	ErrCreateVerificationSessionOperationValidation = errorutil.Validation("failed to create the user")
	ErrStartVerificationSessionOperation            = errorutil.Operation("failed to start verification session")
	ErrStartVerificationSessionOperationValidation  = errorutil.Validation("failed to start verification session")
	ErrGetVerificationExpirationTimeOperation       = errorutil.Validation("failed to get verification session expiration time")
	ErrCompleteVerificationOperation                = errorutil.Operation("failed to complete user verification session")
	ErrCheckUsernameAvailabilityOperation           = errorutil.Operation("failed to check username availability")
	ErrRestartVerificationSessionOperation          = errorutil.Operation("failed to restart user verification session")
	ErrCleanUpVerificationSessionsOperation         = errorutil.Operation("failed to clean up verification sessions")
	ErrCreateOauthUserOperation                     = errorutil.Operation("failed to clean up verification sessions")
	ErrCreateOauthUserOperationValidation           = errorutil.Validation("failed to clean up verification sessions")

	// Client errors
	ErrExternalIdentityAlreadyExists       = errors.New("external identity already exists")
	ErrEmailAlreadyExists                  = errors.New("email already exists")
	ErrUsernameTaken                       = errors.New("username is not available")
	ErrUserPendingVerification             = errors.New("user already exists, but is not verified yet")
	ErrInvalidULID                         = errors.New("failed to parse ulid")
	ErrNonExistingVerificationSession      = errors.New("there is no session associated with the provided id")
	ErrVerificationSessionAlreadyCompleted = errors.New("user has already completed verification")
	ErrUserVerificationSessionExpired      = errors.New("session has expired")
	ErrInvalidToken                        = errors.New("invalid token format")
	ErrInvalidUsername                     = errors.New("blank username")
	ErrUserVerificationRestartSince        = errors.New("user session cannot be restarted")
	ErrUserVerificationRestartAttempts     = errors.New("maximum user restart verification attempts have been reached")

	// Service Errors
	ErrUniqueUsernameGeneration = errors.New("failed to generate a unique username")
)
