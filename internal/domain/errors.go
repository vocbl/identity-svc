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
	// These represent violations of the Verification Aggregate's state machine.
	ErrSessionExpired        = errors.New("verification_session_expired")
	ErrSessionNotStarted     = errors.New("verification_session_not_active")
	ErrSessionAlreadyStarted = errors.New("verification_session_already_in_progress")
	ErrSessionAttemptLimit   = errors.New("verification_attempts_exhausted")
	ErrSessionNotRestartable = errors.New("verification_restart_cooloff_active")

	// --- Cryptographic & Token Invariants ---
	ErrTokenMismatch       = errors.New("verification_token_mismatch")
	ErrInvalidToken        = errors.New("verification_token_malformed")
	ErrInvalidTokenHash    = errors.New("verification_token_hash_invalid")
	ErrInvalidPasswordHash = errors.New("password_hash_malformed")

	// --- Configuration & Policy Invariants ---
	ErrInvalidSessionDuration        = errors.New("policy_invalid_session_duration")
	ErrInvalidRestartSessionDuration = errors.New("policy_invalid_restart_window")

	// --- Identity Conflicts ---
	ErrDuplicateExternalIdentity = errors.New("external_identity_already_linked")

	// --- Technical/Format Invariants ---
	ErrInvalidIdentifier = errors.New("identifier_format_invalid")
)
