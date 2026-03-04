package domain

import "errors"

var (
	// --- Identity & Access Invariants ---
	// These occur when Value Objects (Email, Password, Name) fail validation.
	ErrInvalidEmail      = errors.New("malformed_email_address")
	ErrInvalidExternalID = errors.New("external_identity_id_empty")
	ErrInvalidPassword   = errors.New("password_complexity_not_met")
	ErrInvalidFirstName  = errors.New("first_name_too_short_or_long")
	ErrInvalidLastName   = errors.New("last_name_too_short_or_long")
	ErrInvalidUsername   = errors.New("username_format_invalid")
	ErrInvalidDeviceType = errors.New("unsupported_device_type")

	// --- Verification Session Lifecycle ---
	// These represent violations of the Verification Aggregate state machine.
	ErrVerificationSessionExpired        = errors.New("verification_session_expired")
	ErrVerificationSessionNotStarted     = errors.New("verification_session_not_active")
	ErrVerificationSessionAlreadyStarted = errors.New("verification_session_already_in_progress")
	ErrVerificationSessionAttemptLimit   = errors.New("verification_attempts_exhausted")
	ErrVerificationSessionNotRestartable = errors.New("verification_restart_cooloff_active")

	// --- Auth Session Lifecycle ---
	// These represent violations of the Auth Entity state machine.
	ErrAuthSessionRevoked = errors.New("auth_session_revoked")
	ErrAuthSessionExpired = errors.New("auth_session_expired")

	// --- Cryptographic & Token Invariants ---
	ErrVerificationTokenMismatch    = errors.New("verification_token_mismatch")
	ErrVerificationInvalidToken     = errors.New("verification_token_malformed")
	ErrVerificationInvalidTokenHash = errors.New("verification_token_hash_invalid")

	ErrInvalidPasswordHash = errors.New("password_hash_malformed")
	ErrPasswordMismatch    = errors.New("password_mismatch")

	// --- Configuration & Policy Invariants ---
	ErrInvalidVerificationSessionDuration        = errors.New("policy_invalid_session_duration")
	ErrInvalidVerificationRestartSessionDuration = errors.New("policy_invalid_restart_window")

	// --- Identity Conflicts ---
	ErrDuplicateExternalIdentity = errors.New("external_identity_already_linked")

	// --- Technical/Format Invariants ---
	ErrInvalidIdentifier = errors.New("identifier_format_invalid")
)
