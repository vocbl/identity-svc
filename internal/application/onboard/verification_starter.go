package app

import (
	"context"
	"errors"

	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
)

type VerificationStarterService struct {
	repo               VerificationStarterRepo
	verificationPolicy domain.VerificationPolicy
}

func (s *VerificationStarterService) StartVerificationSession(ctx context.Context, id, emailStr, passwordStr string) error {
	var errs error

	sessionID, err := domain.ParseVerificationSessionID(id)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	password, err := domain.NewPassword(passwordStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return ErrSessionStartOp.ValidationWrap(errs)
	}

	passwordHash, err := password.Hash(s.verificationPolicy)
	if err != nil {
		return ErrSessionStartOp.Wrap(err)
	}

	err = s.repo.WithinTransaction(ctx, func(txRepo VerificationStarterRepo) error {
		session, err := txRepo.Get(ctx, sessionID)
		if err != nil {
			return err
		}

		token, err := session.Start(s.verificationPolicy, passwordHash)
		if err != nil {
			return err
		}

		err = txRepo.Update(ctx, session)
		if err != nil {
			return err
		}

		return txRepo.EmitVerificationSessionStartedEvent(ctx, sessionID, email, token)
	})

	if err != nil {
		if errors.Is(err, db.ErrNonExistingData) {
			err = ErrNotFoundVerificationSession
		}
		return ErrSessionStartOp.Wrap(err)
	}

	return nil
}
