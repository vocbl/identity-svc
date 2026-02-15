package app

import (
	"fmt"
	"time"

	"github.com/vocbl/users-svc/internal/domain/onboard"
)

type OnboardCfg struct {
	//Verification
	VerificationMaxAttempts        int
	VerificationRestartDuration    time.Duration
	VerificationExpirationDuration time.Duration
	VerificationSecurityProvider   onboard.VerificationSecurityProvider

	//Registration
	UsernameGenerationMaxAttempts int
	RegistrationSecurityProvider  onboard.RegistrationSecurityProvider

	//Integration
	IntegrationComponents []string

	//Onboard
	OnboardExpirationDuration time.Duration
}

func NewOnboardService(
	repo OnboardRepo,
	cfg OnboardCfg,
) (*OnboardService, error) {
	registrationPolicy, err := onboard.NewRegistrationPolicy(
		cfg.RegistrationSecurityProvider,
		onboard.RegistrationUsernameGenerationMaxAttempts(cfg.UsernameGenerationMaxAttempts),
	)
	if err != nil {
		return nil, fmt.Errorf("faild to configure registration policy: %w", err)
	}

	verificationPolicy, err := onboard.NewVerificationPolicy(
		cfg.VerificationSecurityProvider,
		onboard.VerificationExpirationDuration(cfg.VerificationExpirationDuration),
		onboard.VerificationMaxAttempts(cfg.VerificationMaxAttempts),
		onboard.VerificationRestartDuration(cfg.VerificationRestartDuration),
	)
	if err != nil {
		return nil, fmt.Errorf("faild to configure verification policy: %w", err)
	}

	integrationPolicy, err := onboard.NewIntegrationPolicy(cfg.IntegrationComponents)
	if err != nil {
		return nil, fmt.Errorf("faild to configure integration policy: %w", err)
	}

	onboardPolicy, err := onboard.NewOnboardPolicy(cfg.OnboardExpirationDuration)
	if err != nil {
		return nil, fmt.Errorf("faild to configure onboard policy: %w", err)
	}

	return &OnboardService{
		repo: repo,
		policy: policy{
			verification: verificationPolicy,
			registration: registrationPolicy,
			integration:  integrationPolicy,
		},
		factory: onboard.NewOnboardFactory(registrationPolicy, verificationPolicy, integrationPolicy, onboardPolicy),
	}, nil
}

type IntegrationCfg struct {
	//Integration
	IntegrationComponents []string
}

func NewIntegrationService(
	repo IntegrationRepo,
	cfg IntegrationCfg,
) (*IntegrationService, error) {
	integrationPolicy, err := onboard.NewIntegrationPolicy(cfg.IntegrationComponents)
	if err != nil {
		return nil, fmt.Errorf("faild to configure integration policy: %w", err)
	}

	return &IntegrationService{
		repo:   repo,
		policy: integrationPolicy,
	}, nil
}
