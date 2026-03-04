package app

import (
	"context"
	"time"

	"github.com/vocbl/users-svc/internal/domain"
)

//go:generate mockgen -destination=./mock/mock_verification_repo.go -package=mock github.com/vocbl/users-svc/internal/application/onboard VerificationRepo
type VerificationRepo interface {
	CheckUsernameExistance(ctx context.Context, username domain.Username) (bool, error)

	Create(ctx context.Context, session *domain.UserVerificationSession) error
	Delete(ctx context.Context, sessionID domain.VerificationSessionID) error
	Clean(ctx context.Context, cleanUpVerificationDuration time.Duration) error
	Get(ctx context.Context, sessionID domain.VerificationSessionID) (*domain.UserVerificationSession, error)
	Update(ctx context.Context, session *domain.UserVerificationSession) error

	CreateUser(ctx context.Context, user *domain.User) error

	EmitVerificationSessionCreatedEvent(ctx context.Context, sessionID domain.VerificationSessionID, email domain.Email, password domain.Password) error
	EmitVerificationSessionStartedEvent(ctx context.Context, sessionID domain.VerificationSessionID, email domain.Email, token domain.Token) error
	EmitUserVerifiedEvent(ctx context.Context, sessionID domain.VerificationSessionID) error

	WithinTransaction(ctx context.Context, fn func(txRepo VerificationRepo) error) error
}

//go:generate mockgen -destination=./mock/mock_user_repo.go -package=mock github.com/vocbl/users-svc/internal/application/onboard UserRepo
type UserRepo interface {
	GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error

	WithinTransaction(ctx context.Context, fn func(txRepo UserRepo) error) error
}

//go:generate mockgen -destination=./mock/mock_verification_starter_repo.go -package=mock github.com/vocbl/users-svc/internal/application/onboard VerificationStarterRepo
type VerificationStarterRepo interface {
	Get(ctx context.Context, sessionID domain.VerificationSessionID) (*domain.UserVerificationSession, error)
	Update(ctx context.Context, session *domain.UserVerificationSession) error

	EmitVerificationSessionStartedEvent(ctx context.Context, sessionID domain.VerificationSessionID, email domain.Email, token domain.Token, restartableSince time.Time) error

	WithinTransaction(ctx context.Context, fn func(txRepo VerificationStarterRepo) error) error
}
