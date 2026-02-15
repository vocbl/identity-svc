package app

import (
	"errors"

	errutil "github.com/vocbl/shared/errors"
)

var (

	//Operations
	ErrUsernameAvailabilityCheckOp   = errutil.NewOperationSt("OnboardService.CheckUsernameAvailability")
	ErrOnboardOp                     = errutil.NewOperationSt("OnboardService.Onboard")
	ErrOnboardFromExternalIdentityOp = errutil.NewOperationSt("OnboardService.OnboardFromExternalIdentiry")
	ErrCancelOnboardingOp            = errutil.NewOperationSt("OnboardService.CancelOnboarding")
	ErrRestartVerificationOp         = errutil.NewOperationSt("OnboardService.RestartVerification")
	ErrCompleteVerificationOp        = errutil.NewOperationSt("OnboardService.CompleteVerification")
	ErrProcessIntegrationUpdateOp    = errutil.NewOperationSt("IntegrationService.ProcessUpdate")

	// Business Constraint Errors: These represent "State Conflicts."
	// Application layer translates DB unique constraints into these.
	ErrConflictEmail            = errors.New("email_already_registered")
	ErrConflictUsername         = errors.New("username_already_taken")
	ErrConflictExternalIdentity = errors.New("identity_provider_already_linked")
	ErrUserUnverified           = errors.New("account_verification_pending")
	ErrUserPendingVerification  = errors.New("user_pending_verification")

	// Resource Availability Errors: Represent "Existence" failures.
	ErrNotFoundSession = errors.New("onboard_session_not_found")
	ErrNotFoundUser    = errors.New("user_record_not_found")

	// Internal Fulfillment Errors: The App layer logic reached a dead end.
	ErrUsernameGenerationExhaustion = errors.New("username_collision_limit_reached")
)
