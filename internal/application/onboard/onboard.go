package app

import (
	"context"
	"errors"
	"time"

	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
)

type OnboardService struct {
	repo               VerificationRepo
	verificationPolicy domain.VerificationPolicy
}

type NewUser struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Username  string
}

func (s *OnboardService) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, ErrUsernameAvailabilityCheckOp.Wrap(domain.ErrInvalidUsername)
	}

	exists, err := s.repo.CheckUsernameExistance(ctx, username)
	if err != nil {
		return false, ErrUsernameAvailabilityCheckOp.Wrap(err)
	}

	return !exists, nil
}

func (s *OnboardService) CreateVerificationSession(ctx context.Context, user NewUser) (domain.VerificationSessionID, time.Time, error) {
	var errs error

	session, err := domain.NewUserVerificationSession(user.Email, user.FirstName, user.LastName, user.Username)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	password, err := domain.NewPassword(user.Password)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		ErrSessionCreateOp.ValidationWrap(errs)
	}

	err = s.repo.WithinTransaction(ctx, func(txRepo VerificationRepo) error {
		err := txRepo.Create(ctx, session)
		if err != nil {
			return err
		}

		return txRepo.EmitVerificationSessionCreatedEvent(ctx, session.ID(), session.Email(), password)
	})

	if err != nil {
		var sessionID domain.VerificationSessionID
		var restartableSince time.Time

		switch {
		case errors.Is(err, db.ErrDublicateEmail):
			err = ErrConflictEmail
		case errors.Is(err, db.ErrDublicateUsername):
			err = ErrConflictUsername
		default:
			var upvErr db.UserPendingVerificationError
			if errors.As(err, &upvErr) {
				err = ErrUserUnverified
				sessionID = upvErr.SessionID
				restartableSince = upvErr.RestartableSince
			}

		}
		return sessionID, restartableSince, ErrSessionCreateOp.Wrap(err)
	}

	return session.ID(), session.RestartableSince(), nil
}

func (s *OnboardService) RestartVerificationSession(ctx context.Context, id string) error {
	sessionID, err := domain.ParseVerificationSessionID(id)
	if err != nil {
		return ErrSessionRestartOp.Wrap(err)
	}

	err = s.repo.WithinTransaction(ctx, func(txRepo VerificationRepo) error {
		session, err := s.repo.Get(ctx, sessionID)
		if err != nil {
			return err
		}

		token, err := session.ResetToken(s.verificationPolicy)
		if err != nil {
			return err
		}

		err = txRepo.Update(ctx, session)
		if err != nil {
			return err
		}

		return txRepo.EmitVerificationSessionStartedEvent(ctx, sessionID, session.Email(), token)
	})

	if err != nil {
		if errors.Is(err, db.ErrNonExistingData) {
			err = ErrNotFoundVerificationSession
		}
		return ErrSessionRestartOp.Wrap(err)
	}

	return nil
}

func (s *OnboardService) DeleteVerificationSession(ctx context.Context, id string) error {
	sessionID, err := domain.ParseVerificationSessionID(id)
	if err != nil {
		return ErrSessionDeleteOp.Wrap(err)
	}

	err = s.repo.Delete(ctx, sessionID)
	if err != nil {
		if errors.Is(err, db.ErrNonExistingData) {
			err = ErrNotFoundVerificationSession
		}
		return ErrSessionDeleteOp.Wrap(err)
	}

	return nil
}

func (s *OnboardService) CompleteVerification(ctx context.Context, id, tokenStr string) error {
	sessionID, err := domain.ParseVerificationSessionID(id)
	if err != nil {
		return ErrSessionCompleteOp.Wrap(err)
	}

	token, err := domain.NewToken(tokenStr)
	if err != nil {
		return ErrSessionCompleteOp.Wrap(err)
	}

	err = s.repo.WithinTransaction(ctx, func(tx VerificationRepo) error {
		session, err := tx.Get(ctx, sessionID)
		if err != nil {
			return err
		}

		user, err := session.Complete(s.verificationPolicy, token)
		if err != nil {
			return err
		}

		err = tx.CreateUser(ctx, user)
		if err != nil {
			return err
		}

		return tx.EmitUserVerifiedEvent(ctx, sessionID)
	})

	if err != nil {
		if errors.Is(err, db.ErrNonExistingData) {
			err = ErrNotFoundVerificationSession
		}

		return ErrSessionCompleteOp.Wrap(err)
	}

	return nil
}

func (s *OnboardService) CleanUpVerificationSessions(ctx context.Context) error {
	return ErrSessionCleanupOp.WrapCheck(s.repo.Clean(ctx, s.verificationPolicy.CleanUpDuration()))
}
