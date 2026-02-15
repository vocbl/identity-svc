package app

import (
	"context"
	"time"

	"github.com/oklog/ulid"
	"github.com/vocbl/users-svc/internal/domain"
)

type AuthRepo interface {
	SaveSession(ctx context.Context, session *domain.UserJwtSession) error
	UpdateRefreshToken(ctx context.Context, refreshTokenID ulid.ULID, token string) error
	RevokeSession(ctx context.Context, userID, refreshTokenID ulid.ULID) error
	GetSessions(ctx context.Context, userID ulid.ULID) ([]domain.UserJwtSession, error)
	WithinTransaction(ctx context.Context, fn func(txRepo AuthRepo) error) error
	EmitNewSessionEvent(ctx context.Context, refreshTokenID ulid.ULID, duration time.Duration) error

	//returns password, userID and error
	GetAuthData(ctx context.Context, email string) (string, ulid.ULID, error)
	GetUserPasswordHash(ctx context.Context, userID ulid.ULID) (string, error)
	GetVerificationSessionID(ctx context.Context, email string) (ulid.ULID, error)
	EmitSessionRevokationEvent(ctx context.Context, refreshTokenID ulid.ULID) error

	IsActive(ctx context.Context, refreshTokenID ulid.ULID) (bool, error)
}

type authCache interface {
	IsActive(ctx context.Context, refreshTokenID ulid.ULID) (bool, error)
}
