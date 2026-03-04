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

type Username string

func ParseUsername(username string) (Username, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", ErrInvalidUsername
	}
	return Username(username), nil
}

type UserCreds struct {
	firstName string
	lastName  string
	username  Username
}

func NewUserCreds(firstName, lastName, usernameStr string) (UserCreds, error) {
	var errs error

	firstName = strings.TrimSpace(firstName)
	if firstName == "" {
		errs = errors.Join(errs, ErrInvalidFirstName)
	}

	firstName = strings.TrimSpace(lastName)
	if lastName == "" {
		errs = errors.Join(errs, ErrInvalidLastName)
	}

	username, err := ParseUsername(usernameStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return UserCreds{}, errs
	}

	return UserCreds{
		firstName: firstName,
		lastName:  lastName,
		username:  Username(username),
	}, nil
}

func (c *UserCreds) GenerateSetUsername(suffixLength int) {
	first := strings.ToLower(strings.ReplaceAll(c.firstName, " ", ""))

	suffix := make([]byte, 3+suffixLength)
	rand.Read(suffix)

	c.username = Username(fmt.Sprintf("%s_%s", first, hex.EncodeToString(suffix)))
}

func MakeUsername(firstName, lastName string) string {
	return strings.ToLower(strings.TrimSpace(firstName) + "_" + strings.TrimSpace(lastName))
}

type Password string

func (p Password) Hash(policy VerificationPolicy) (PasswordHash, error) {
	hash, err := policy.hashPassword(p)
	if err != nil {
		return "", err
	}

	return PasswordHash(hash), nil
}

func ParsePassword(passwordStr string) (Password, error) {
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

func newPasswordHash(passwordHashStr string) (PasswordHash, error) {
	const expected = 60
	actual := len(passwordHashStr)
	if actual == expected {
		return PasswordHash(passwordHashStr), nil
	}
	return "", fmt.Errorf("%w: invalid password hash length: expected %d, got %d", ErrInvalidPasswordHash, expected, actual)

}

func (ph PasswordHash) Verify(policy *AuthPolicy, password Password) error {
	passwordHash, err := policy.hashPassword(password)
	if err != nil {
		return err
	}

	if passwordHash != ph {
		return ErrPasswordMismatch
	}

	return nil
}

type Email string

func ParseEmail(email string) (Email, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil || addr.Address != email {
		return "", ErrInvalidEmail
	}

	return Email(addr.Address), nil
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

	email, err := ParseEmail(emailStr)
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
