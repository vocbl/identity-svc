package onboard

import "errors"

var (

	// --- Verification ---
	ErrVerificationExpired        = errors.New("verification_expired")
	ErrVerificationAttemptLimit   = errors.New("verification_attempt_limit_reached")
	ErrVerificationNotRestartable = errors.New("verification_restart_not_allowed")

	ErrVerificationTokenMismatch    = errors.New("verification_token_mismatch")
	ErrVerificationAlreadyCompleted = errors.New("verification_already_completed")

	ErrInvalidVerificationToken = errors.New("verification_token_invalid")

	// --- Verification Policy ---
	ErrInvalidVerificationDuration        = errors.New("policy_verification_duration_invalid")
	ErrInvalidVerificationRestartDuration = errors.New("policy_verification_restart_duration_invalid")

	// --- Integration ---
	ErrInvalidIntegrationComponent      = errors.New("integration_component_not_found")
	ErrIntegrationComponentAlreadyReady = errors.New("integration_component_already_completed")

	// --- Registration ---
	ErrPasswordHashing = errors.New("password_hashing_failed")

	// --- Onboarding ---
	ErrIntegrationNotCompleted  = errors.New("integration_not_completed")
	ErrVerificationNotCompleted = errors.New("verificaion_nor_completed")
	ErrOnboardSessionExpired    = errors.New("onboard_session_expired")
)
