package identity

import "errors"

var (
	//ID
	ErrInvalidIdentifier = errors.New("invalid identifier")

	//User
	ErrInvalidUsername           = errors.New("invalid username")
	ErrInvalidFirstName          = errors.New("invalid first name")
	ErrInvalidLastName           = errors.New("invalid last name")
	ErrInvalidPassword           = errors.New("password must be at least 6 chars with 1 uppercase and 1 digit")
	ErrInvalidPasswordHash       = errors.New("invalid password hash")
	ErrInvalidEmail              = errors.New("invalid email address")
	ErrInvalidExternalIdentifier = errors.New("invalid external identifier")
	ErrDuplicateExternalIdentity = errors.New("external identity already exists")

	//Token
	ErrInvalidToken = errors.New("invalid or expired token")
)
