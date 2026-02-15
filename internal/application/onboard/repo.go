package app

import (
	"context"

	identity "github.com/vocbl/users-svc/internal/domain"
	onboard "github.com/vocbl/users-svc/internal/domain/onboard"

	domain "github.com/vocbl/users-svc/internal/domain"
)

//go:generate mockgen -destination=./mock/mock_verification_repo.go -package=mock github.com/vocbl/users-svc/internal/application/onboard VerificationRepo
type OnboardRepo interface {
	//Pre onboarding
	CheckUsernameExistance(ctx context.Context, username domain.Username) (bool, error)

	//Onboard
	CreateOnboardSession(ctx context.Context, session *onboard.OnboardSession) error
	DeleteOnboardSession(ctx context.Context, onboardID onboard.OnboardID) error
	GetOnboardSession(ctx context.Context, onboardID onboard.OnboardID) (*onboard.OnboardSession, error)

	GetVerification(ctx context.Context, onboardID onboard.OnboardID) (onboard.Verification, identity.Email, error)
	UpdateVerification(ctx context.Context, verification onboard.Verification) error

	CreateUser(ctx context.Context, user *domain.User) error

	//Events
	EmitExistingEmailRegistrationEttemptEvent(ctx context.Context, email domain.Email) error
	EmitUserRegisteredEvent(ctx context.Context, onboardID onboard.OnboardID) error
	EmitVerificationRequestEvent(ctx context.Context, onboardID onboard.OnboardID, email domain.Email, token onboard.VerificationToken) error
	EmitUserVerifiedEvent(ctx context.Context, onboardID onboard.OnboardID) error
	EmitIntegrationCancelationEvent(ctx context.Context, onboardID onboard.OnboardID) error

	WithinTransaction(ctx context.Context, fn func(ctx context.Context, tx OnboardRepo) error) error
}

type IntegrationRepo interface {
	WithinTransaction(ctx context.Context, fn func(tx IntegrationRepo) error) error
	GetOnboardSession(ctx context.Context, onboardID onboard.OnboardID) (*onboard.OnboardSession, error)
	DeleteOnboardSession(ctx context.Context, onboardID onboard.OnboardID) error
	UpdateIntegration(ctx context.Context, integration onboard.Integration) error

	CreateUser(ctx context.Context, user *domain.User) error
}
