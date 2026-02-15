package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"
	errorutil "github.com/vocbl/shared/errors"
)

var (
	//Operation errors
	ErrLoginOperation            = errorutil.Operation("failed to authenticate user")
	ErrRevokeSessionOperation    = errorutil.Operation("failed to revoke user's session")
	ErrGetSessionsOperation      = errorutil.Operation("failed to get user's sessions")
	ErrIssueAccessTokenOperation = errorutil.Operation("failed to issue access token")

	//Client errors
	ErrInvalidEmail          = errors.New("provided email is not valid")
	ErrNonExistingUser       = errors.New("there is no user assosiated with provided email")
	ErrInvalidPassword       = errors.New("provided password is invalid")
	ErrInvalidUserID         = errors.New("provided id is not of a valid format")
	ErrInvalidRefreshTokenID = errors.New("provided id is not of a valid format")
	ErrInvalidRefreshToken   = errors.New("provided token is not valid")
	ErrRevokedRefreshToken   = errors.New("provided token has been revoked")

	ErrNonExistingSession = errors.New("there is no session assosiated with provided id")
)

type ErrUserPendingVerification struct {
	VerificaitonID   ulid.ULID
	RestartableSince time.Time
}

func (e *ErrUserPendingVerification) Error() string {
	return fmt.Sprintf("%s :user has not been verified yet", ErrLoginOperation)
}
