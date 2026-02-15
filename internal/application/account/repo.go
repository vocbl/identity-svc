package app

import (
	"context"
	"time"

	"github.com/oklog/ulid"
	"github.com/vocbl/users-svc/internal/domain"
)

type AccountRepo interface {
	GetCreds(ctx context.Context, userID ulid.ULID) (*domain.UserCreds, error)
	ChangeCreds(ctx context.Context, userID ulid.ULID, creds *domain.UserCreds) error
}

type PasswordRepo interface {
	SaveChangePasswordSession(ctx context.Context, session *domain.PasswordChangeSession) error
	DeleteChangePasswordSession(ctx context.Context, tokenHash string) (string, time.Time, error)

	ChangePassword(ctx context.Context, email, passwordHash string) error

	EmitPasswordChangeEvent(ctx context.Context, email, token string) error
	WithinTransaction(ctx context.Context, fn func(txRepo PasswordRepo) error) error
}
