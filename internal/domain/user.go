package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"
	"unicode"
)

type UserCreds struct {
	firstName string
	lastName  string
	username  string
}

func NewUserCreds(firstName, lastName, username string) (UserCreds, error) {
	var errs error
	if firstName == "" {
		errs = errors.Join(errs, ErrInvalidFirstName)
	}

	if lastName == "" {
		errs = errors.Join(errs, ErrInvalidLastName)
	}

	if username == "" {
		errs = errors.Join(errs, ErrInvalidUsername)
	}

	if errs != nil {
		return UserCreds{}, errs
	}

	return UserCreds{
		firstName: firstName,
		lastName:  lastName,
		username:  username,
	}, nil
}

func (c *UserCreds) GenerateSetUsername(suffixLength int) {
	first := strings.ToLower(strings.ReplaceAll(c.firstName, " ", ""))

	suffix := make([]byte, 2+suffixLength)
	rand.Read(suffix)

	c.username = fmt.Sprintf("%s_%s", first, hex.EncodeToString(suffix))
}

func MakeUsername(firstName, lastName string) string {
	return strings.Join([]string{strings.TrimSpace(firstName), strings.TrimSpace(lastName)}, "_")
}

type Password string

func (p Password) Hash(policy VerificationPolicy) (PasswordHash, error) {
	hash, err := policy.hashPassword(p)
	if err != nil {
		return "", err
	}

	return PasswordHash(hash), nil
}

func NewPassword(passwordStr string) (Password, error) {
	if len(passwordStr) < 6 {
		return "", ErrInvalidPassword
	}

	var hasUpper, hasDigit bool
	for _, r := range passwordStr {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if hasUpper && hasDigit {
			return Password(passwordStr), nil
		}
	}

	return "", ErrInvalidPassword
}

type PasswordHash string

func NewPasswordHash(passwordHashStr string) (PasswordHash, error) {
	const expected = 60
	actual := len(passwordHashStr)
	if actual == expected {
		return PasswordHash(passwordHashStr), nil
	}
	return "", fmt.Errorf("%w: invalid password hash length: expected %d, got %d", ErrInvalidPasswordHash, expected, actual)

}

type Email string

func NewEmail(emailStr string) (Email, error) {
	addr, err := mail.ParseAddress(emailStr)
	if err != nil || addr.Address != emailStr {
		return "", ErrInvalidEmail
	}

	return Email(emailStr), nil
}

type User struct {
	id                 UserID
	email              Email
	passwordHash       *PasswordHash
	Creds              UserCreds
	createdAt          time.Time
	updatedAt          time.Time
	externalIdentities []ExternalIdentity
}

type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "Google"
)

type ExternalIdentity struct {
	provider AuthProvider
	id       string
}

func NewUser(emailStr, firstNameStr, lastNameStr, usernameStr string) (*User, error) {
	var errs error

	email, err := NewEmail(emailStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	creds, err := NewUserCreds(firstNameStr, lastNameStr, usernameStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, errs
	}

	return &User{
		id:    newUserID(),
		email: email,
		Creds: creds,
	}, nil
}

func (u *User) AddExternalIdentity(externnalID string, provider AuthProvider) error {
	if slices.ContainsFunc(u.externalIdentities, func(identity ExternalIdentity) bool {
		return identity.id == externnalID
	}) {
		return ErrDuplicateExternalIdentity
	}

	if strings.TrimSpace(externnalID) == "" {
		return ErrInvalidIdentifier
	}

	u.externalIdentities = append(u.externalIdentities, ExternalIdentity{
		provider: provider,
		id:       externnalID,
	})

	return nil
}
