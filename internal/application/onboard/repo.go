package app

import (
	"context"
	"time"

	"github.com/oklog/ulid"
	"github.com/vocbl/users-svc/internal/domain"
)

//go:generate mockgen -destination=./mock/mock_onboard_repo.go -package=mock github.com/vocbl/users-svc/internal/application/onboard OnboardRepo
type OnboardRepo interface {
	UserLookupRepo
	VerificationSessionRepo
	UserIdentityRepo
	OnboardingEventRepo

	WithinTransaction(ctx context.Context, fn func(txRepo OnboardRepo) error) error
}

type UserLookupRepo interface {
	CheckUsernameExistance(ctx context.Context, username string) (bool, error)
}

type VerificationSessionRepo interface {
	SaveUserVerificationSession(ctx context.Context, session *domain.UserVerificationSession) error
	SetUserVerificationSessionPasswordHash(ctx context.Context, sessionID ulid.ULID, passwordHash string) error
	CompleteUserVerificationSession(ctx context.Context, hashedToken string) (*domain.UserVerificationSession, error)
	RestartUserVerificationSession(ctx context.Context, sessionID ulid.ULID, restartableSince time.Time, hashedToken string) (string, int, error)
	GetVerificationSessionExpirationTime(ctx context.Context, sessionID ulid.ULID) (time.Time, error)
	DeleteUserVerificationSession(ctx context.Context, sessionID ulid.ULID) error
	CleanUpVerificationSessions(ctx context.Context, cleanUpVerificationDuration time.Duration) error
}

type UserIdentityRepo interface {
	SaveUser(ctx context.Context, user *domain.User) error
	SaveUserExternalIdentity(ctx context.Context, email string, identity domain.ExternalIdentity) (ulid.ULID, error)
}

type OnboardingEventRepo interface {
	EmitUserCreatedEvent(ctx context.Context, sessionID, email, password, token string) error
	EmitVerifyUserEvent(ctx context.Context, email, token string) error
	EmitUserVerifiedEvent(ctx context.Context, sessionID ulid.ULID) error
}
