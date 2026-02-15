package domain

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid"
	ulidutil "github.com/vocbl/shared/utils/ulid"
)

type UserCreds struct {
	FirstName string
	LastName  string
	Username  string
}

func (c *UserCreds) Validate() []error {
	var errs []error
	if c.FirstName == "" {
		errs = append(errs, ErrInvalidFirstName)
	}

	if c.LastName == "" {
		errs = append(errs, ErrInvalidLastName)
	}

	if c.Username == "" {
		errs = append(errs, ErrInvalidUsername)
	}

	return errs
}

func (c *UserCreds) GenerateUsername(suffixLength int) {
	first := strings.ToLower(strings.ReplaceAll(c.FirstName, " ", ""))

	suffix := make([]byte, suffixLength)
	rand.Read(suffix)

	c.Username = fmt.Sprintf("%s_%s", first, hex.EncodeToString(suffix))
}

type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "Google"
)

type ExternalIdentity struct {
	Provider AuthProvider
	ID       string
}
type User struct {
	ID                 ulid.ULID
	Email              string
	PasswordHash       sql.NullString
	Creds              UserCreds
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ExternalIdentities []ExternalIdentity
}

func (u *User) OauthValidateAndEnrich() []error {
	var errs []error

	if !ValidateEmail(u.Email) {
		errs = append(errs, ErrInvalidEmail)
	}

	if u.Creds.FirstName == "" {
		errs = append(errs, ErrInvalidFirstName)
	}

	if u.Creds.LastName == "" {
		errs = append(errs, ErrInvalidLastName)
	}

	if errs == nil {
		u.ID = ulidutil.NewULID()
	}

	return errs
}

type UserVerificationSession struct {
	SessionID        ulid.ULID
	TokenHash        string
	Email            string
	PasswordHash     sql.NullString
	Creds            UserCreds
	Duration         time.Duration
	RestartableSince time.Time
	CreatedAt        time.Time
	RestartedAt      time.Time
	CompletedAt      sql.NullTime
}

func (uvs *UserVerificationSession) Validate(password string) []error {
	var errs []error

	if !ValidateEmail(uvs.Email) {
		errs = append(errs, ErrInvalidEmail)
	}

	if !ValidatePassword(password) {
		errs = append(errs, ErrInvalidPassword)
	}

	return append(errs, uvs.Creds.Validate()...)
}

func (uvs *UserVerificationSession) Enrich(tokenHash string, duration time.Duration, restartableSince time.Time) {
	uvs.TokenHash = tokenHash
	uvs.SessionID = ulidutil.NewULID()
	uvs.Duration = duration
	uvs.RestartableSince = restartableSince
}

func (uvs *UserVerificationSession) IsRestartable() bool {
	return time.Now().UTC().After(uvs.RestartableSince)
}

func (uvs *UserVerificationSession) ToUser(userID ulid.ULID) *User {
	return &User{
		ID:           userID,
		Email:        uvs.Email,
		Creds:        uvs.Creds,
		PasswordHash: uvs.PasswordHash,
	}
}

type VerificationPolicy struct {
	MaxAttempts     int
	Duration        time.Duration
	RestartDuration time.Duration
	CleanUpDuration time.Duration
}

func (c VerificationPolicy) Validate() error {
	err := ValidateUserVerificationSessionDuration(c.Duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	err = ValidateUserVerificationSessionRestartDuration(c.RestartDuration)
	if err != nil {
		return fmt.Errorf("invalid restart duration: %w", err)
	}

	err = ValidateUserVerificationSessionCleanUpDuration(c.CleanUpDuration)
	if err != nil {
		return fmt.Errorf("invalid clean up duration: %w", err)
	}

	err = ValidateUserVerificationSessionMaxAttempts(c.MaxAttempts)
	if err != nil {
		return fmt.Errorf("invalid max attempts: %w", err)
	}

	return nil
}
