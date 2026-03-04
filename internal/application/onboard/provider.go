package app

import (
	"time"

	"github.com/vocbl/users-svc/internal/domain"
)

type VerificationCfg struct {
	MaxAttempts                   int
	ExpirationDuration            time.Duration
	RestartDuration               time.Duration
	CleanUpDuration               time.Duration
	PasswordHasher                func(string) (string, error)
	TokenGenerator                func() (string, string)
	TokenHasher                   func(string) string
	UsernameGenerationMaxAttempts int
}

func NewOnboardService(
	repo VerificationRepo,
	cfg VerificationCfg,
) (*OnboardService, error) {
	verificationPolicy, err := domain.NewVerificationPolicy(
		cfg.PasswordHasher,
		cfg.TokenGenerator,
		cfg.TokenHasher,
		domain.VerificationPolicyMaxAttempts(cfg.MaxAttempts),
		domain.VerificationPolicyExpirationDuration(cfg.ExpirationDuration),
		domain.VerificationPolicyCleanUpDuration(cfg.CleanUpDuration),
		domain.VerificationPolicyRestartDuration(cfg.RestartDuration),
		domain.VerificationPolicyUsernameGenerationMaxAttempts(cfg.UsernameGenerationMaxAttempts),
	)

	if err != nil {
		return nil, err
	}

	return &OnboardService{
		repo:               repo,
		verificationPolicy: verificationPolicy,
	}, nil
}

func NewVerificationStarterService(
	repo VerificationStarterRepo,
	cfg VerificationCfg,
) (*VerificationStarterService, error) {
	verificationPolicy, err := domain.NewVerificationPolicy(
		cfg.PasswordHasher,
		cfg.TokenGenerator,
		cfg.TokenHasher,
		domain.VerificationPolicyMaxAttempts(cfg.MaxAttempts),
		domain.VerificationPolicyExpirationDuration(cfg.ExpirationDuration),
		domain.VerificationPolicyCleanUpDuration(cfg.CleanUpDuration),
		domain.VerificationPolicyRestartDuration(cfg.RestartDuration),
	)

	if err != nil {
		return nil, err
	}

	return &VerificationStarterService{
		repo:               repo,
		verificationPolicy: verificationPolicy,
	}, nil
}

func NewExternalIdentityService(
	repo UserRepo,
	cfg VerificationCfg,
) (*ExternalIdentityService, error) {
	verificationPolicy, err := domain.NewVerificationPolicy(
		cfg.PasswordHasher,
		cfg.TokenGenerator,
		cfg.TokenHasher,
		domain.VerificationPolicyMaxAttempts(cfg.MaxAttempts),
		domain.VerificationPolicyExpirationDuration(cfg.ExpirationDuration),
		domain.VerificationPolicyCleanUpDuration(cfg.CleanUpDuration),
		domain.VerificationPolicyRestartDuration(cfg.RestartDuration),
	)

	if err != nil {
		return nil, err
	}

	return &ExternalIdentityService{
		repo:               repo,
		verificationPolicy: verificationPolicy,
	}, nil
}
