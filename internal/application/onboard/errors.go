package app

import (
	"errors"

	errorutil "github.com/vocbl/shared/errors"
)

var (
	// Operation Errors: Act as the "Category" or "Context" for logging and tracing.
	// They describe the high-level intent that failed.
	ErrSessionCreateOp             = errorutil.Operation("verification.session.create")
	ErrSessionStartOp              = errorutil.Operation("verification.session.start")
	ErrSessionDeleteOp             = errorutil.Operation("verification.session.delete")
	ErrSessionCompleteOp           = errorutil.Operation("verification.session.complete")
	ErrSessionRestartOp            = errorutil.Operation("verification.session.restart")
	ErrSessionCleanupOp            = errorutil.Operation("verification.session.cleanup")
	ErrUsernameAvailabilityCheckOp = errorutil.Operation("user.username.check")
	ErrExternalUserCreateOp        = errorutil.Operation("user.external.create")
	ErrExternalIdentityAttachOp    = errorutil.Operation("user.external.link")

	// Business Constraint Errors: These represent "State Conflicts."
	// Application layer translates DB unique constraints into these.
	ErrConflictEmail            = errors.New("email_already_registered")
	ErrConflictUsername         = errors.New("username_already_taken")
	ErrConflictExternalIdentity = errors.New("identity_provider_already_linked")
	ErrUserUnverified           = errors.New("account_verification_pending")

	// Resource Availability Errors: Represent "Existence" failures.
	ErrNotFoundVerificationSession = errors.New("verification_session_not_found")
	ErrNotFoundUser                = errors.New("user_record_not_found")

	// Internal Fulfillment Errors: The App layer logic reached a dead end.
	ErrUsernameExhaustion = errors.New("username_collision_limit_reached")
)
