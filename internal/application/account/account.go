package app

import (
	"context"
	"errors"

	"github.com/oklog/ulid"
	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
)

type AccountService struct {
	repo AccountRepo
}

func NewAccountService(repo AccountRepo) *AccountService {
	return &AccountService{
		repo: repo,
	}
}

func (s *AccountService) GetCreds(ctx context.Context, userID string) (*domain.UserCreds, error) {
	UserULID, err := ulid.Parse(userID)
	if err != nil {
		return nil, ErrGetCredsOperation.Wrap(ErrInvalidID)
	}

	creds, err := s.repo.GetCreds(ctx, UserULID)
	if err != nil {
		if errors.Is(err, db.ErrNonExistingData) {
			return nil, ErrGetCredsOperation.Wrap(ErrNonExistingUser)
		}
		return nil, ErrGetCredsOperation.Wrap(err)
	}

	return creds, nil
}

type CredsUpdateRequest struct {
	UserID    string
	FirstName string
	LastName  string
	Username  string
}

func (r CredsUpdateRequest) ParseCreds() *domain.UserCreds {
	return &domain.UserCreds{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Username:  r.Username,
	}
}

func (s *AccountService) UpdateCreds(ctx context.Context, req CredsUpdateRequest) error {
	UserULID, err := ulid.Parse(req.UserID)
	if err != nil {
		return ErrUpdateCredsOperation.Wrap(ErrInvalidID)
	}

	creds := req.ParseCreds()
	errs := creds.Validate()
	if errs != nil {
		return ErrUpdateCredsOperationValidation.Wrap(errs...)
	}

	err = s.repo.ChangeCreds(ctx, UserULID, creds)
	if err != nil {
		if errors.Is(err, db.ErrNonExistingData) {
			return ErrUpdateCredsOperation.Wrap(ErrNonExistingUser)
		}
		return ErrUpdateCredsOperation.Wrap(err)
	}

	return nil
}
