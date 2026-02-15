package app

import (
	"errors"

	errorutil "github.com/vocbl/shared/errors"
)

var (
	//Operation errors
	ErrLoginOperation            = errorutil.NewOperationSt("failed to authenticate user")
	ErrRevokeSessionOperation    = errorutil.NewOperationSt("failed to revoke user's session")
	ErrGetSessionsOperation      = errorutil.NewOperationSt("failed to get user's sessions")
	ErrIssueAccessTokenOperation = errorutil.NewOperationSt("failed to issue access token")

	//Client errors
	ErrInvalidEmail            = errors.New("provided email is not valid")
	ErrNonExistingUser         = errors.New("there is no user assosiated with provided email")
	ErrInvalidPassword         = errors.New("provided password is invalid")
	ErrInvalidUserID           = errors.New("provided id is not of a valid format")
	ErrInvalidRefreshTokenID   = errors.New("provided id is not of a valid format")
	ErrInvalidRefreshToken     = errors.New("provided token is not valid")
	ErrRevokedRefreshToken     = errors.New("provided token has been revoked")
	ErrUserPendingVerification = errors.New("user")
	ErrNonExistingSession      = errors.New("there is no session assosiated with provided id")
)
