package onboard

import (
	"errors"
	"fmt"
	"strings"
	"time"

	identity "github.com/vocbl/users-svc/internal/domain"
)

const (
	// Verification (defaults are minimums)
	DefaultVerificationMaxAttempts        = 5
	DefaultVerificationExpirationDuration = 15 * time.Minute
	DefaultVerificationRestartDuration    = 2 * time.Minute

	// Registration
	DefaultRegistrationUsernameGenerationMaxAttempts = 5

	// Onboard
	DefaultOnboardExpirationDuration = 30 * time.Second
)

// ========== Verification Policy ============

type VerificationToken string

func (t VerificationToken) String() string {
	return string(t)
}

func ParseVerificationToken(policy *VerificationPolicy, tokenStr string) (VerificationToken, error) {
	if policy.securityProvider.ValidateToken(tokenStr) {
		return VerificationToken(tokenStr), nil
	}

	return "", ErrInvalidVerificationToken
}

type VerificationTokenHash string
type VerificationSecurityProvider interface {
	GenerateToken() (string, string)
	HashToken(token string) string

	ValidateToken(token string) bool
}

type VerificationPolicy struct {
	maxAttempts        int
	restartDuration    time.Duration
	expirationDuration time.Duration
	securityProvider   VerificationSecurityProvider
}

func (p *VerificationPolicy) generateToken() (VerificationToken, VerificationTokenHash) {
	token, tokenHash := p.securityProvider.GenerateToken()
	return VerificationToken(token), VerificationTokenHash(tokenHash)
}

func (p *VerificationPolicy) hashToken(token VerificationToken) VerificationTokenHash {
	tokenHash := p.securityProvider.HashToken(token.String())

	return VerificationTokenHash(tokenHash)
}

type VerificationOption func(*VerificationPolicy) error

func VerificationMaxAttempts(n int) VerificationOption {
	return func(p *VerificationPolicy) error {
		if n < DefaultVerificationMaxAttempts {
			return fmt.Errorf(
				"invalid verification max attempts: expected >= %d, got %d",
				DefaultVerificationMaxAttempts,
				n,
			)
		}
		p.maxAttempts = n
		return nil
	}
}

func VerificationRestartDuration(d time.Duration) VerificationOption {
	return func(p *VerificationPolicy) error {
		if d < DefaultVerificationRestartDuration {
			return fmt.Errorf(
				"invalid verification restart duration: expected >= %v, got %v",
				DefaultVerificationRestartDuration,
				d,
			)
		}
		p.restartDuration = d
		return nil
	}
}

func VerificationExpirationDuration(d time.Duration) VerificationOption {
	return func(p *VerificationPolicy) error {
		if d < DefaultVerificationExpirationDuration {
			return fmt.Errorf(
				"invalid verification expiration duration: expected >= %v, got %v",
				DefaultVerificationExpirationDuration,
				d,
			)
		}
		p.expirationDuration = d
		return nil
	}
}

// ========== Registration Policy ============

type RegistrationSecurityProvider interface {
	HashPassword(password string) (string, error)
}

type RegistrationPolicy struct {
	usernameGenerationMaxAttempts int
	securityProvider              RegistrationSecurityProvider
}

func (p *RegistrationPolicy) hashPassword(password identity.Password) (identity.PasswordHash, error) {
	hash, err := p.securityProvider.HashPassword(password.String())
	if err != nil {
		return "", err
	}
	return identity.PasswordHash(hash), nil
}

type RegistrationOption func(*RegistrationPolicy) error

func RegistrationUsernameGenerationMaxAttempts(n int) RegistrationOption {
	return func(p *RegistrationPolicy) error {
		if n < DefaultRegistrationUsernameGenerationMaxAttempts {
			return fmt.Errorf(
				"invalid username generation attempts: expected >= %d, got %d",
				DefaultRegistrationUsernameGenerationMaxAttempts,
				n,
			)
		}
		p.usernameGenerationMaxAttempts = n
		return nil
	}
}

// ========== Integration Policy ============

type IntegrationPolicy struct {
	components []IntegrationComponent
}

// ========= Onboard Policy ========

type OnboardPolicy struct {
	expirationDuration time.Duration
}

// ========== Constructors ============

func NewOnboardPolicy(duration time.Duration) (*OnboardPolicy, error) {
	if duration < DefaultOnboardExpirationDuration {
		return nil, fmt.Errorf("invalid onboard expiration duration: expected >= %v, got %v", DefaultOnboardExpirationDuration, duration)
	}

	return &OnboardPolicy{
		expirationDuration: duration,
	}, nil
}

func NewVerificationPolicy(
	securityProvider VerificationSecurityProvider,
	opts ...VerificationOption,
) (*VerificationPolicy, error) {

	p := &VerificationPolicy{
		maxAttempts:        DefaultVerificationMaxAttempts,
		restartDuration:    DefaultVerificationRestartDuration,
		expirationDuration: DefaultVerificationExpirationDuration,
		securityProvider:   securityProvider,
	}

	var errs error

	token, tokenHash := securityProvider.GenerateToken()

	if !securityProvider.ValidateToken(token) {
		errs = errors.Join(errs, errors.New("generated token failed provided validation"))
	}

	if securityProvider.HashToken(token) != tokenHash {
		errs = errors.Join(errs, errors.New("hashed token does not match generated one"))
	}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if errs != nil {
		return nil, errs
	}

	return p, nil
}

func NewRegistrationPolicy(securityProvider RegistrationSecurityProvider, opts ...RegistrationOption) (*RegistrationPolicy, error) {
	p := &RegistrationPolicy{
		securityProvider:              securityProvider,
		usernameGenerationMaxAttempts: DefaultRegistrationUsernameGenerationMaxAttempts,
	}

	var errs error

	for _, opt := range opts {
		if err := opt(p); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if errs != nil {
		return nil, errs
	}

	return p, nil
}

func NewIntegrationPolicy(components []string) (*IntegrationPolicy, error) {
	p := &IntegrationPolicy{
		components: make([]IntegrationComponent, len(components)),
	}

	var errs error

	if len(components) == 0 {
		errs = errors.Join(errs, errors.New("integration policy requires at least one component"))
	}

	for i, component := range components {
		if strings.Contains(component, " ") || component == "" {
			errs = errors.Join(errs, fmt.Errorf("integration policy component has to contain no spaces and be not blank, got '%v'", component))
		} else {
			p.components[i] = IntegrationComponent(component)
		}
	}

	if errs != nil {
		return nil, errs
	}

	return p, nil
}

// ========== Getters ============

func (p *VerificationPolicy) getMaxAttempts() int {
	return p.maxAttempts
}

func (p *VerificationPolicy) getExpirationDuration() time.Duration {
	return p.expirationDuration
}

func (p *VerificationPolicy) getRestartDuration() time.Duration {
	return p.restartDuration
}

func (p *RegistrationPolicy) UsernameGenerationMaxAttempts() int {
	return p.usernameGenerationMaxAttempts
}
