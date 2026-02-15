package app

import (
	"errors"

	errorutil "github.com/vocbl/users-svc/internal/shared/error"
)

var (
	//Operation errors
	ErrGetCredsOperation              = errorutil.Operation("failed to get user's creds")
	ErrUpdateCredsOperation           = errorutil.Operation("failed to update user's creds")
	ErrUpdateCredsOperationValidation = errorutil.Validation("failed to update user's creds")

	//Client errors
	ErrInvalidID       = errors.New("invalid id format")
	ErrNonExistingUser = errors.New("there is no user assosiated with provided id")
)
