package app

import (
	"errors"

	errutil "github.com/vocbl/shared/errors"
)

var (
	// Operation Errors: Act as the "Category" or "Context" for logging and tracing.
	// They describe the high-level intent that failed.
	ErrSessionCreateOp             = errutil.NewOperationSt("verification.session.create")
	ErrSessionStartOp              = errutil.NewOperationSt("verification.session.start")
	ErrSessionDeleteOp             = errutil.NewOperationSt("verification.session.delete")
	ErrSessionCompleteOp           = errutil.NewOperationSt("verification.session.complete")
	ErrSessionRestartOp            = errutil.NewOperationSt("verification.session.restart")
	ErrSessionCleanupOp            = errutil.NewOperationSt("verification.session.cleanup")
	ErrUsernameAvailabilityCheckOp = errutil.NewOperationSt("user.username.check")
	ErrExternalUserCreateOp        = errutil.NewOperationSt("user.external.create")
	ErrExternalIdentityAttachOp    = errutil.NewOperationSt("user.external.link")

	// Business Constraint Errors: These represent "State Conflicts."
	// Application layer translates DB unique constraints into these.
	ErrConflictEmail            = errors.New("email_already_registered")
	ErrConflictUsername         = errors.New("username_already_taken")
	ErrConflictExternalIdentity = errors.New("identity_provider_already_linked")
	ErrUserUnverified           = errors.New("account_verification_pending")
	ErrUserPendingVerification  = errors.New("user_pending_verification")

	// Resource Availability Errors: Represent "Existence" failures.
	ErrNotFoundVerificationSession = errors.New("verification_session_not_found")
	ErrNotFoundUser                = errors.New("user_record_not_found")

	// Internal Fulfillment Errors: The App layer logic reached a dead end.
	ErrUsernameExhaustion = errors.New("username_collision_limit_reached")
)
