package identity

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

// ========== Identity Types ============
type UserID struct {
	ID
}

func NewUserID() UserID {
	return UserID{NewID()}
}

//========== Value Objects ============

type Username string

func ParseUsername(s string) (Username, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ErrInvalidUsername
	}
	return Username(s), nil
}

type Email string

func ParseEmail(s string) (Email, error) {
	s = strings.TrimSpace(s)
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return "", ErrInvalidEmail
	}
	return Email(addr.Address), nil
}

type Password string

func (p Password) String() string {
	return string(p)
}

func ParsePassword(s string) (Password, error) {
	if len(s) < 6 {
		return "", ErrInvalidPassword
	}

	var hasUpper, hasDigit bool
	for _, r := range s {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	if !hasUpper || !hasDigit {
		return "", ErrInvalidPassword
	}

	return Password(s), nil
}

type PasswordHash string

type UserCreds struct {
	firstName string
	lastName  string
	username  Username
}

func NewUserCreds(firstName, lastName, usernameStr string) (UserCreds, error) {
	var errs error

	f := strings.TrimSpace(firstName)
	if f == "" {
		errs = errors.Join(errs, ErrInvalidFirstName)
	}

	l := strings.TrimSpace(lastName)
	if l == "" {
		errs = errors.Join(errs, ErrInvalidLastName)
	}

	u, err := ParseUsername(usernameStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return UserCreds{}, errs
	}

	return UserCreds{
		firstName: f,
		lastName:  l,
		username:  u,
	}, nil
}

func RebuildUserCreds(f, l string, u Username) UserCreds {
	return UserCreds{f, l, u}
}

func (c *UserCreds) GenerateUsername(suffixLength int) {
	prefix := strings.ToLower(strings.ReplaceAll(c.firstName, " ", ""))

	buf := make([]byte, suffixLength)
	rand.Read(buf)

	c.username = Username(fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(buf)))
}

func (c UserCreds) FirstName() string  { return c.firstName }
func (c UserCreds) LastName() string   { return c.lastName }
func (c UserCreds) Username() Username { return c.username }

type ExternalProvider string

const (
	ExternalProviderGoogle ExternalProvider = "Google"
)

type ExternalID string

type ExternalIdentity struct {
	provider   ExternalProvider
	externalID ExternalID
}

func RebuildExternalIdentity(provider, externalID string) (ExternalIdentity, error) {
	externalID = strings.ToLower(strings.TrimSpace(externalID))
	if externalID == "" {
		return ExternalIdentity{}, ErrInvalidExternalIdentifier
	}

	switch providerParsed := ExternalProvider(provider); providerParsed {
	case ExternalProviderGoogle:
		return ExternalIdentity{
			provider:   providerParsed,
			externalID: ExternalID(externalID),
		}, nil
	default:
		return ExternalIdentity{}, errors.New("undefiend external provider")
	}

}

func NewExternalIdentity(provider ExternalProvider, externalID string) (ExternalIdentity, error) {
	externalID = strings.ToLower(strings.TrimSpace(externalID))
	if externalID == "" {
		return ExternalIdentity{}, ErrInvalidExternalIdentifier
	}
	return ExternalIdentity{provider: provider, externalID: ExternalID(externalID)}, nil
}

func (e ExternalIdentity) Provider() ExternalProvider { return e.provider }
func (e ExternalIdentity) ExternalID() ExternalID     { return e.externalID }

//========== User Aggregate ============

type User struct {
	id                 UserID
	email              Email
	passwordHash       *PasswordHash
	creds              UserCreds
	externalIdentities []ExternalIdentity
	createdAt          time.Time
	updatedAt          time.Time
}

func NewUser(email Email, hash PasswordHash, creds UserCreds) *User {
	now := time.Now()
	u := &User{
		id:           NewUserID(),
		email:        email,
		passwordHash: &hash,
		creds:        creds,
		createdAt:    now,
		updatedAt:    now,
	}

	return u
}

func (u *User) AddExternalIdentity(ext ExternalIdentity) error {
	exists := slices.ContainsFunc(u.externalIdentities, func(e ExternalIdentity) bool {
		return e.externalID == ext.externalID && e.provider == ext.provider
	})

	if exists {
		return ErrDuplicateExternalIdentity
	}

	u.externalIdentities = append(u.externalIdentities, ext)
	u.updatedAt = time.Now()
	return nil
}

// Getters

func (u *User) ID() UserID                             { return u.id }
func (u *User) Email() Email                           { return u.email }
func (u *User) Creds() UserCreds                       { return u.creds }
func (u *User) ExternalIdentities() []ExternalIdentity { return slices.Clone(u.externalIdentities) }
func (u *User) CreatedAt() time.Time                   { return u.createdAt }
func (u *User) UpdatedAt() time.Time                   { return u.updatedAt }
