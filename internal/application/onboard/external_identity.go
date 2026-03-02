package app

import (
	"context"
	"errors"

	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
)

type ExternalIdentityService struct {
	repo               UserRepo
	verificationPolicy domain.VerificationPolicy
}

type OAuthUser struct {
	Email      string
	FirstName  string
	LastName   string
	ExternalID string
}

func (s *ExternalIdentityService) CreateUserFromExternalIdentity(ctx context.Context, provider domain.AuthProvider, ou OAuthUser) error {
	user, err := domain.NewUser(ou.Email, ou.FirstName, ou.LastName, domain.MakeUsername(ou.FirstName, ou.LastName))
	if err != nil {
		return ErrExternalUserCreateOp.ValidationWrap(err)
	}

	err = user.AddExternalIdentity(ou.ExternalID, provider)
	if err != nil {
		return ErrExternalUserCreateOp.ValidationWrap(err)
	}

	for i := 0; i < 5; i++ {
		err = s.repo.Create(ctx, user)
		if err == nil {
			return nil
		}

		switch {
		case errors.Is(err, db.ErrDublicateEmail):
			return ErrExternalUserCreateOp.Wrap(ErrConflictEmail)
		case errors.Is(err, db.ErrDublicateUsername):
			user.Creds.GenerateSetUsername(i)
			continue
		default:
			return ErrExternalUserCreateOp.Wrap(err)
		}

	}

	return ErrExternalUserCreateOp.Wrap(ErrUsernameExhaustion)
}

func (s *ExternalIdentityService) AtachUserExternalIdentity(ctx context.Context, emailStr string, provider domain.AuthProvider, externalID string) error {
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return ErrExternalIdentityAttachOp.ValidationWrap(err)
	}

	err = s.repo.WithinTransaction(ctx, func(txRepo UserRepo) error {
		user, err := txRepo.GetByEmail(ctx, email)
		if err != nil {
			return err
		}

		err = user.AddExternalIdentity(externalID, provider)
		if err != nil {
			return err
		}

		return txRepo.Update(ctx, user)
	})

	if err != nil {
		switch {
		case errors.Is(err, db.ErrNonExistingData):
			err = ErrNotFoundUser
		case errors.Is(err, db.ErrDublicateData):
			err = ErrConflictExternalIdentity
		}
		return ErrExternalIdentityAttachOp.Wrap(err)
	}

	return nil
}
