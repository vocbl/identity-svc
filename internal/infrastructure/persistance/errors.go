package db

import (
	"errors"
	"time"

	"github.com/vocbl/users-svc/internal/domain"
)

var (
	ErrDublicateData                   = errors.New("data already exists (unique constraint)")
	ErrDublicateEmail                  = errors.New("email already exists (unique constraint)")
	ErrDublicateExternalIdentity       = errors.New("email already exists (unique constraint)")
	ErrDublicateUsername               = errors.New("username already exists (unique constraint)")
	ErrNonExistingData                 = errors.New("non-existing data")
	ErrUserVerificationCompleted       = errors.New("user verification has already been comleted")
	ErrUserVerificationRestartSince    = errors.New("user verification cannot be restarted yet")
	ErrUserVerificationRestartAttempts = errors.New("maximum user verification attempts have been reached")
	ErrSessionExpired                  = errors.New("session has expired")
	ErrUserPendingVerification         = errors.New("user verificaiton has not been completed yet")
)

type UserPendingVerificationError struct {
	VerificaitonID   domain.VerificationSessionID
	PasswordHash     domain.PasswordHash
	RestartableSince time.Time
}

func (UserPendingVerificationError) Error() string {
	return "user has not been verified yet"
}
