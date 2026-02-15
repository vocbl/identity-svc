package onboard

import (
	"errors"
	"time"
)

type verificationMethod string

const (
	VerificationMethodEmailToken       verificationMethod = "email_token"
	VerificationMethodExternalIdentity verificationMethod = "external_identity"
)

func parseVerificationMethod(method string) (verificationMethod, error) {
	switch parseMethod := verificationMethod(method); parseMethod {
	case VerificationMethodEmailToken, VerificationMethodExternalIdentity:
		return parseMethod, nil
	default:
		return "", errors.New("invalid verification method")
	}
}

type Verification struct {
	onboradID        OnboardID
	method           verificationMethod
	tokenHash        *VerificationTokenHash
	attemptCount     int
	completedAt      *time.Time
	expiresAt        *time.Time
	restartableSince *time.Time
}

func (v *Verification) IsCompleted() bool {
	return v.completedAt != nil
}

func newCompletedVerificationFromExternalIdentity(onboardID OnboardID) Verification {
	now := time.Now().UTC()

	return Verification{
		onboradID:    onboardID,
		method:       VerificationMethodExternalIdentity,
		attemptCount: 1,
		completedAt:  &now,
	}
}

func newEmailVerification(policy *VerificationPolicy, onboardID OnboardID) (Verification, VerificationToken) {
	verification := Verification{onboradID: onboardID, method: VerificationMethodEmailToken}
	token := verification.prepareStart(policy)
	return verification, token
}

func (v *Verification) ResetToken(policy *VerificationPolicy) (VerificationToken, error) {
	if !v.restartableSince.Before(time.Now().UTC()) {
		return "", ErrVerificationNotRestartable
	}

	if v.attemptCount == policy.maxAttempts {
		return "", ErrVerificationAttemptLimit
	}

	return v.prepareStart(policy), nil
}

func (v *Verification) prepareStart(policy *VerificationPolicy) VerificationToken {
	token, tokenHash := policy.generateToken()
	v.tokenHash = &tokenHash

	now := time.Now().UTC()
	expiresAt := now.Add(policy.expirationDuration)
	restartableSince := now.Add(policy.restartDuration)

	v.restartableSince = &restartableSince
	v.expiresAt = &expiresAt
	v.attemptCount++

	return token
}

func (v *Verification) Complete(policy *VerificationPolicy, token VerificationToken) error {
	if !v.expiresAt.After(time.Now().UTC()) {
		return ErrVerificationExpired
	} else if v.completedAt != nil {
		return ErrVerificationAlreadyCompleted
	}

	if tokenHash := policy.hashToken(token); tokenHash != *v.tokenHash {
		return ErrVerificationTokenMismatch
	}

	now := time.Now().UTC()
	v.completedAt = &now
	v.expiresAt = nil
	v.restartableSince = nil

	return nil
}

func RebuildVerification(
	onboardID OnboardID,
	methodStr string,
	tokenHashStr *string,
	attemptCount int,
	completedAt *time.Time,
	expiresAt *time.Time,
	restartableSince *time.Time,
) (Verification, error) {
	var errs error

	method, err := parseVerificationMethod(methodStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	var tokenHashPointer *VerificationTokenHash
	if tokenHashStr != nil {
		tokenHash := VerificationTokenHash(*tokenHashStr)
		tokenHashPointer = &tokenHash
	}

	if attemptCount < 1 {
		errs = errors.Join(errs, errors.New("invalid attemptCount value"))
	}

	if completedAt != nil {
		if expiresAt != nil || restartableSince != nil {
			errs = errors.Join(errs, errors.New("invalid verification state: completed verification cannot expire or restart"))
		}
	} else {
		if expiresAt == nil || restartableSince == nil {
			errs = errors.Join(errs, errors.New("invalid verification state: not completed verification has to be restartable"))
		}
	}

	if method == VerificationMethodExternalIdentity {
		if tokenHashPointer != nil {
			errs = errors.Join(errs, errors.New("invalid verification state: external identity does not require token hash"))
		}

		if expiresAt != nil || restartableSince != nil || completedAt == nil {
			errs = errors.Join(errs, errors.New("invalid verification state: external identity verification always has to be completed"))
		}
	} else {
		if tokenHashPointer == nil {
			errs = errors.Join(errs, errors.New("invalid verification state: has to contain verification token hash"))
		}
	}

	if errs != nil {
		return Verification{}, errs
	}

	return Verification{
		onboradID:        onboardID,
		method:           method,
		tokenHash:        tokenHashPointer,
		attemptCount:     attemptCount,
		completedAt:      completedAt,
		expiresAt:        expiresAt,
		restartableSince: restartableSince,
	}, nil
}
