package domain

import "errors"

var (
	//Client input errors
	ErrInvalidEmail      = errors.New("email is not valid")
	ErrInvalidPassword   = errors.New("password is not valid")
	ErrInvalidFirstName  = errors.New("first name is not valid")
	ErrInvalidLastName   = errors.New("last name is not valid")
	ErrInvalidUsername   = errors.New("username is not valid")
	ErrInvalidDeviceType = errors.New("device type is not valid")

	//Domain input errors
	ErrInvalidSessionDuration        = errors.New("invalid session duration")
	ErrInvalidRestartSessionDuration = errors.New("invalid session restart duration")
	ErrInvalidTokenHash              = errors.New("token hash is not valid")
	ErrInvalidToken                  = errors.New("token is not valid")
	ErrInvalidPasswordHash           = errors.New("password hash is not valid")
	ErrNilULID                       = errors.New("ulid is not valid")
)
