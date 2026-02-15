package app

import (
	"context"
	"errors"
	"time"

	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
	"github.com/vocbl/users-svc/internal/shared/security"
)

type PasswordService struct {
	repo            PasswordRepo
	sessionDuration time.Duration
}

func NewUserPasswordService(repo PasswordRepo, sessionDuration time.Duration) *PasswordService {
	return &PasswordService{repo: repo, sessionDuration: sessionDuration}
}

func (s *PasswordService) NewChangeSession(ctx context.Context, email string) error {
	session, err := domain.NewPasswordChangeSession(email, s.sessionDuration)
	if err != nil {
		return ErrChangeSessionOperation.Wrap(err)
	}

	token, tokenHash := security.GenerateToken()
	err = session.Enrich(tokenHash)
	if err != nil {
		return ErrChangeSessionOperation.Wrap(err)
	}

	err = s.repo.WithinTransaction(ctx, func(txRepo PasswordRepo) error {
		err := txRepo.SaveChangePasswordSession(ctx, session)
		if err != nil {
			if errors.Is(err, db.ErrNonExistingData) {
				return ErrChangeSessionOperation.Wrap(ErrNonExistingEmail)
			}
			return ErrChangeSessionOperation.Wrap(err)
		}

		return txRepo.EmitPasswordChangeEvent(ctx, session.Email, token)
	})

	return ErrChangeSessionOperation.WrapCheck(err)
}

func (s *PasswordService) KillChangeSession(ctx context.Context, token string) error {
	if !security.IsValidToken(token) {
		return ErrKillChangeSessionOperation.Wrap(ErrInvalidToken)
	}

	_, _, err := s.repo.DeleteChangePasswordSession(ctx, security.HashToken(token))
	return ErrKillChangeSessionOperation.WrapCheck(err)
}

func (s *PasswordService) CompleteChangeSession(ctx context.Context, newPassword, token string) error {
	if !domain.ValidatePassword(newPassword) {
		return ErrCompleteChangeSessionOperation.Wrap(ErrInvalidPassowrd)
	}

	err := s.repo.WithinTransaction(ctx, func(txRepo PasswordRepo) error {
		email, expiresAt, err := txRepo.DeleteChangePasswordSession(ctx, security.HashToken(token))
		if err != nil {
			if errors.Is(err, db.ErrNonExistingData) {
				return ErrNonExistingSession
			}
			return err
		}

		if expiresAt.Before(time.Now().UTC()) {
			return ErrSessionExpired
		}

		passwordHash, err := security.HashPassword(newPassword)
		if err != nil {
			return err
		}

		err = txRepo.ChangePassword(ctx, email, passwordHash)
		if err != nil {
			if errors.Is(err, db.ErrNonExistingData) {
				return ErrNonExistingEmail
			}
		}

		return err
	})

	return ErrCompleteChangeSessionOperation.WrapCheck(err)
}
