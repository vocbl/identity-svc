package app

import (
	"context"
	"errors"

	user "github.com/vocbl/users-svc/internal/domain"
	onboard "github.com/vocbl/users-svc/internal/domain/onboard"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
)

type policy struct {
	registration *onboard.RegistrationPolicy
	verification *onboard.VerificationPolicy
	integration  *onboard.IntegrationPolicy
}

type OnboardService struct {
	repo    OnboardRepo
	policy  policy
	factory *onboard.Factory
}

func (s *OnboardService) CheckUsernameAvailability(ctx context.Context, usernameStr string) (bool, error) {
	username, err := user.ParseUsername(usernameStr)
	if err != nil {
		return false, ErrUsernameAvailabilityCheckOp.WrapValidation(err)
	}

	exists, err := s.repo.CheckUsernameExistance(ctx, username)
	if err != nil {
		return false, ErrUsernameAvailabilityCheckOp.Wrap(err)
	}

	return !exists, nil
}

type OAuthUser struct {
	Email      string
	FirstName  string
	LastName   string
	ExternalID string
}

func (s *OnboardService) OnboardFromExternalIdentiry(ctx context.Context, provider user.ExternalProvider, ou OAuthUser) (onboard.OnboardID, error) {
	session, err := s.factory.NewOnboardingFromExternalIdentity(ou.Email, ou.FirstName, ou.LastName, provider, ou.ExternalID)
	if err != nil {
		return onboard.OnboardID{}, ErrOnboardFromExternalIdentityOp.WrapValidation(err)
	}

	err = s.repo.WithinTransaction(ctx, func(ctx context.Context, tx OnboardRepo) error {
		for i := 0; i < s.policy.registration.UsernameGenerationMaxAttempts(); i++ {
			err = tx.CreateOnboardSession(ctx, session)
			if err == nil {
				tx.EmitUserRegisteredEvent(ctx, session.ID())
				tx.EmitUserVerifiedEvent(ctx, session.ID())
				return nil
			}

			switch {
			case errors.Is(err, db.ErrDublicateEmail):
				return ErrOnboardFromExternalIdentityOp.WrapFew(ErrConflictEmail, db.ErrDublicateEmail)
			case errors.Is(err, db.ErrDublicateUsername):
				session.Registration.Creds.GenerateUsername(i)
				continue
			default:
				return ErrOnboardFromExternalIdentityOp.Wrap(err)
			}
		}
		return ErrUsernameGenerationExhaustion
	})

	if err != nil {
		return onboard.OnboardID{}, ErrOnboardFromExternalIdentityOp.WrapIs(err, db.ErrDublicateEmail, ErrConflictEmail)
	}

	return session.ID(), nil
}

type NewUser struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Username  string
}

func (s *OnboardService) Onboard(ctx context.Context, user NewUser) (onboard.OnboardID, error) {
	session, token, err := s.factory.NewOnboarding(user.Email, user.Password, user.FirstName, user.LastName, user.Username)
	if err != nil {
		if errors.Is(err, onboard.ErrPasswordHashing) {
			err = ErrOnboardOp.Wrap(err)
		} else {
			err = ErrOnboardOp.WrapValidation(err)
		}
		return onboard.OnboardID{}, err
	}

	err = s.repo.WithinTransaction(ctx, func(ctx context.Context, tx OnboardRepo) error {
		err := tx.CreateOnboardSession(ctx, session)
		if err != nil {
			if errors.Is(err, db.ErrDublicateEmail) {
				emitErr := tx.EmitExistingEmailRegistrationEttemptEvent(ctx, session.Registration.Email())
				if emitErr != nil {
					err = errors.Join(err, emitErr)
				}
			}
			return err
		}

		err = tx.EmitUserRegisteredEvent(ctx, session.ID())
		if err != nil {
			return err
		}

		return tx.EmitVerificationRequestEvent(ctx, session.ID(), session.Registration.Email(), token)
	})

	if err != nil {
		var errs error

		if errors.Is(err, db.ErrDublicateEmail) {
			errs = errors.Join(errs, ErrConflictEmail)
		}

		if errors.Is(err, db.ErrDublicateUsername) {
			errs = errors.Join(errs, ErrConflictUsername)
		}

		if errs != nil {
			return onboard.OnboardID{}, ErrOnboardOp.WrapFew(errs, err)
		}

		return onboard.OnboardID{}, ErrOnboardOp.Wrap(err)
	}

	return session.ID(), nil
}

func (s *OnboardService) RestartVerification(ctx context.Context, onboardIdStr string) error {
	onboardID, err := onboard.ParseOnboardID(onboardIdStr)
	if err != nil {
		return ErrRestartVerificationOp.WrapValidation(err)
	}

	err = s.repo.WithinTransaction(ctx, func(ctx context.Context, tx OnboardRepo) error {
		verification, email, err := s.repo.GetVerification(ctx, onboardID)
		if err != nil {
			return err
		}

		token, err := verification.ResetToken(s.policy.verification)
		if err != nil {
			return err
		}

		err = tx.UpdateVerification(ctx, verification)
		if err != nil {
			return err
		}

		return tx.EmitVerificationRequestEvent(ctx, onboardID, email, token)
	})

	return ErrRestartVerificationOp.WrapCheckIs(err, db.ErrNonExistingData, ErrNotFoundSession)
}

func (s *OnboardService) CancelOnboarding(ctx context.Context, onboardIdStr string) error {
	onboardID, err := onboard.ParseOnboardID(onboardIdStr)
	if err != nil {
		return ErrCancelOnboardingOp.WrapValidation(err)
	}

	err = s.repo.WithinTransaction(ctx, func(ctx context.Context, tx OnboardRepo) error {
		err := s.repo.DeleteOnboardSession(ctx, onboardID)
		if err != nil {
			return err
		}

		return tx.EmitIntegrationCancelationEvent(ctx, onboardID)
	})

	return ErrCancelOnboardingOp.WrapCheckIs(err, db.ErrNonExistingData, ErrNotFoundSession)
}

func (s *OnboardService) CompleteVerification(ctx context.Context, onboardIdStr, tokenStr string) error {
	onboardID, err := onboard.ParseOnboardID(onboardIdStr)
	if err != nil {
		return ErrCompleteVerificationOp.WrapValidation(err)
	}

	token, err := onboard.ParseVerificationToken(s.policy.verification, tokenStr)
	if err != nil {
		return ErrCompleteVerificationOp.WrapValidation(err)
	}

	err = s.repo.WithinTransaction(ctx, func(ctx context.Context, tx OnboardRepo) error {
		session, err := tx.GetOnboardSession(ctx, onboardID)
		if err != nil {
			return err
		}

		err = session.Verification.Complete(s.policy.verification, token)
		if err != nil {
			return err
		}

		user, err := session.Complete(s.policy.integration)
		if err != nil {
			err = tx.EmitUserVerifiedEvent(ctx, onboardID)
			if err != nil {
				return err
			}

			return tx.UpdateVerification(ctx, session.Verification)
		}

		err = tx.DeleteOnboardSession(ctx, onboardID)
		if err != nil {
			return nil
		}

		return tx.CreateUser(ctx, user)
	})

	return ErrCompleteVerificationOp.WrapCheckIs(err, db.ErrNonExistingData, ErrNotFoundSession)
}

type IntegrationService struct {
	repo   IntegrationRepo
	policy *onboard.IntegrationPolicy
}

func (s IntegrationService) ProcessUpdate(ctx context.Context, onboardIdStr, componentStr string) error {
	onboardID, err := onboard.ParseOnboardID(onboardIdStr)
	if err != nil {
		return ErrProcessIntegrationUpdateOp.WrapValidation(err)
	}

	component, err := onboard.ParseIntegrationComponent(s.policy, componentStr)
	if err != nil {
		return ErrProcessIntegrationUpdateOp.WrapValidation(err)
	}

	err = s.repo.WithinTransaction(ctx, func(tx IntegrationRepo) error {
		session, err := tx.GetOnboardSession(ctx, onboardID)
		if err != nil {
			return err
		}

		err = session.Integration.MarkComponentReady(s.policy, component)
		if err != nil {
			return err
		}

		user, err := session.Complete(s.policy)
		if err != nil {
			return tx.UpdateIntegration(ctx, session.Integration)
		}

		err = tx.DeleteOnboardSession(ctx, onboardID)
		if err != nil {
			return err
		}

		return tx.CreateUser(ctx, user)
	})

	return ErrProcessIntegrationUpdateOp.WrapCheckIs(err, db.ErrNonExistingData, ErrNotFoundSession)
}
