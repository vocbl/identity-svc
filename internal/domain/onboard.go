package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid"
)

type TokenHash string

func (th TokenHash) String() string {
	return string(th)
}

type Token string

func (th Token) String() string {
	return string(th)
}

func ParseToken(token string) (Token, error) {
	token = strings.TrimSpace(token)
	if err := validateToken(token); err != nil {
		return "", fmt.Errorf("%w: %w", ErrVerificationInvalidToken, err)
	}
	return Token(token), nil
}

type UserVerificationSession struct {
	id               VerificationSessionID
	tokenHash        *TokenHash
	email            Email
	passwordHash     *PasswordHash
	attemptCount     int
	Creds            UserCreds
	expiresAt        *time.Time
	restartableSince *time.Time
	createdAt        time.Time
}

func (uvs *UserVerificationSession) RestartableSince() time.Time {
	return *uvs.restartableSince
}

func (uvs *UserVerificationSession) ID() VerificationSessionID {
	return uvs.id
}

func (uvs *UserVerificationSession) Email() Email {
	return uvs.email
}

func NewUserVerificationSession(emailStr, passwordStr, firstName, lastName, username string) (*UserVerificationSession, Password, error) {
	var errs error

	email, err := ParseEmail(emailStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	creds, err := NewUserCreds(firstName, lastName, username)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	password, err := ParsePassword(passwordStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, "", errs
	}

	return &UserVerificationSession{
		id:    newVerificationSessionID(),
		email: email,
		Creds: creds,
	}, password, nil
}

func RebuildUserVerificationSession(
	ulid ulid.ULID,
	tokenHash *TokenHash,
	email Email,
	passwordHash *PasswordHash,
	attemptCount int,
	creds UserCreds,
	expiresAt *time.Time,
	restartableSince *time.Time,
	createdAt time.Time,
) *UserVerificationSession {
	return &UserVerificationSession{
		id:               newVerificationSessionID(),
		tokenHash:        tokenHash,
		email:            email,
		passwordHash:     passwordHash,
		attemptCount:     attemptCount,
		Creds:            creds,
		expiresAt:        expiresAt,
		restartableSince: restartableSince,
		createdAt:        createdAt,
	}
}

func (uvs *UserVerificationSession) Start(policy VerificationPolicy, passwordHash PasswordHash) (Token, error) {
	if uvs.IsActive() {
		return "", ErrVerificationSessionAlreadyStarted
	}

	uvs.passwordHash = &passwordHash
	return uvs.prepareStart(policy), nil
}

func (uvs *UserVerificationSession) IsActive() bool {
	return uvs.passwordHash != nil && uvs.expiresAt != nil && uvs.restartableSince != nil && uvs.attemptCount != 0
}

func (uvs *UserVerificationSession) prepareStart(policy VerificationPolicy) Token {
	token, tokenHash := policy.generateToken()
	uvs.tokenHash = &tokenHash

	now := time.Now().UTC()
	expiresAt := now.Add(policy.expirationDuration)
	restartableSince := now.Add(policy.restartDuration)

	uvs.restartableSince = &restartableSince
	uvs.expiresAt = &expiresAt

	uvs.attemptCount++

	return token
}

func (uvs *UserVerificationSession) ResetToken(policy VerificationPolicy) (Token, error) {
	if !uvs.IsActive() {
		return "", ErrVerificationSessionNotStarted
	}

	if !uvs.restartableSince.Before(time.Now().UTC()) {
		return "", ErrVerificationSessionNotRestartable
	}

	if uvs.attemptCount == policy.maxAttempts {
		return "", ErrVerificationSessionAttemptLimit
	}

	return uvs.prepareStart(policy), nil
}

func (uvs *UserVerificationSession) Complete(policy VerificationPolicy, token Token) (*User, error) {
	if uvs.passwordHash == nil || uvs.expiresAt == nil {
		return nil, ErrVerificationSessionNotStarted
	}

	if !uvs.expiresAt.After(time.Now().UTC()) {
		return nil, ErrVerificationSessionExpired
	}

	if tokenHash := policy.hashToken(token); tokenHash != *uvs.tokenHash {
		return nil, ErrVerificationTokenMismatch
	}

	return &User{
		id:           newUserID(),
		email:        uvs.email,
		Creds:        uvs.Creds,
		passwordHash: uvs.passwordHash,
	}, nil
}

type VerificationPolicy struct {
	maxAttempts                   int
	expirationDuration            time.Duration
	restartDuration               time.Duration
	cleanUpDuration               time.Duration
	usernameGenerationMaxAttempts int
	hashPassword                  func(password Password) (PasswordHash, error)
	generateToken                 func() (Token, TokenHash)
	hashToken                     func(token Token) TokenHash
}

func (vp *VerificationPolicy) CleanUpDuration() time.Duration {
	return vp.cleanUpDuration
}

func (vp *VerificationPolicy) UsernameGenerationMaxAttempts() int {
	return vp.usernameGenerationMaxAttempts
}

type VerificationOption func(*VerificationPolicy) error

func VerificationPolicyMaxAttempts(n int) VerificationOption {
	return func(p *VerificationPolicy) error {
		min, max := 2, 10
		if min >= 2 && max <= 10 {
			p.maxAttempts = n
			return nil
		}
		return fmt.Errorf("invalid max attempts: expected between %d and %d, got %d", min, max, n)
	}
}

func VerificationPolicyExpirationDuration(d time.Duration) VerificationOption {
	return func(p *VerificationPolicy) error {
		min, max := 10*time.Minute, 24*time.Hour
		if d >= min && d <= max {
			p.expirationDuration = d
			return nil
		}
		return fmt.Errorf("invalid session expiration duration: expected between %v and %v, got %v", min, max, d)
	}
}

func VerificationPolicyRestartDuration(d time.Duration) VerificationOption {
	return func(p *VerificationPolicy) error {
		min, max := time.Minute, 10*time.Minute
		if d >= min && d <= max {
			p.restartDuration = d
			return nil
		}
		return fmt.Errorf("invalid restart duration: expected between %v and %v, got %v", min, max, d)
	}
}

func VerificationPolicyCleanUpDuration(d time.Duration) VerificationOption {
	return func(p *VerificationPolicy) error {
		min, max := 30*time.Minute, 24*time.Hour
		if d >= min && d <= max {
			p.cleanUpDuration = d
			return nil
		}
		return fmt.Errorf("invalid cleanup duration: expected between %v and %v, got %v", min, max, d)
	}
}

func VerificationPolicyUsernameGenerationMaxAttempts(n int) VerificationOption {
	return func(p *VerificationPolicy) error {
		min, max := 3, 10
		if n >= min && n <= max {
			p.usernameGenerationMaxAttempts = n
			return nil
		}
		return fmt.Errorf("invalid username generation max ettempts value: expected between %v and %v, got %v", min, max, n)
	}
}

func NewVerificationPolicy(
	passwordHasher func(password string) (string, error),
	tokenGenerator func() (string, string),
	tokenHasher func(token string) string,
	opts ...VerificationOption,
) (VerificationPolicy, error) {
	policy := VerificationPolicy{
		maxAttempts:                   5,
		expirationDuration:            15 * time.Minute,
		restartDuration:               2 * time.Minute,
		cleanUpDuration:               3 * time.Hour,
		usernameGenerationMaxAttempts: 3,
		hashPassword: func(password Password) (PasswordHash, error) {
			passwordHash, err := passwordHasher(string(password))
			return PasswordHash(passwordHash), err
		},
		generateToken: func() (Token, TokenHash) {
			token, tokenHash := tokenGenerator()
			return Token(token), TokenHash(tokenHash)
		},

		hashToken: func(token Token) TokenHash {
			return TokenHash(tokenHasher(string(token)))
		},
	}

	var errs error
	token, tokenHash := tokenGenerator()
	err := validateTokenHash(tokenHash)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid tokenGenerator: %w", err))
	}

	err = validateToken(token)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid tokenGenerator: %w", err))
	}

	tokenHasherResult := tokenHasher(token)
	if err = validateTokenHash(tokenHash); err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid tokenHasher: %w", err))
	} else if tokenHasherResult != tokenHash {
		errs = errors.Join(errs, fmt.Errorf("invalid tokenHasher: does not match the tokenGenerator hash"))
	}

	passwordHashStr, err := passwordHasher((string(TestValidPassword)))
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid passwordHasher: failed to hash password: %w", err))
	} else if _, err = newPasswordHash(passwordHashStr); err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid passwordHasher: %w", err))
	}

	for _, opt := range opts {
		err = opt(&policy)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if errs != nil {
		return VerificationPolicy{}, errs
	}

	return policy, nil
}
