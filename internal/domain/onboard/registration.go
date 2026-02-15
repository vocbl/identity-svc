package onboard

import (
	"errors"
	"fmt"
	"strings"

	identity "github.com/vocbl/users-svc/internal/domain"
)

type Registration struct {
	onboardID        OnboardID
	email            identity.Email
	passwordHash     *identity.PasswordHash
	Creds            identity.UserCreds
	externalIdentity *identity.ExternalIdentity
}

func (r *Registration) Email() identity.Email {
	return r.email
}

func (r *Registration) OnboardID() OnboardID {
	return r.onboardID
}

func newRegistration(policy *RegistrationPolicy, emailStr, passwordStr, firstName, lastName, username string) (Registration, error) {
	var errs error

	email, err := identity.ParseEmail(emailStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	creds, err := identity.NewUserCreds(firstName, lastName, username)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	password, err := identity.ParsePassword(passwordStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return Registration{}, errs
	}

	passwordHash, err := policy.hashPassword(password)
	if err != nil {
		return Registration{}, fmt.Errorf("%w: %w", ErrPasswordHashing, err)
	}

	return Registration{
		onboardID:    OnboardID{identity.NewID()},
		email:        email,
		passwordHash: &passwordHash,
		Creds:        creds,
	}, nil
}

func newRegistrationFromExternalIdentity(provider identity.ExternalProvider, externalID, emailStr, firstName, lastName string) (Registration, error) {
	var errs error

	email, err := identity.ParseEmail(emailStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	creds, err := identity.NewUserCreds(firstName, lastName, strings.ToLower(firstName+"_"+lastName))
	if err != nil {
		errs = errors.Join(errs, err)
	}

	externalIdentity, err := identity.NewExternalIdentity(provider, externalID)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return Registration{}, errs
	}

	return Registration{
		onboardID:        OnboardID{identity.NewID()},
		email:            email,
		Creds:            creds,
		externalIdentity: &externalIdentity,
	}, nil
}

func RebuildRegistration(
	onboardID OnboardID,
	emailStr string,
	passwordHashStr *string,
	firstName string,
	lastName string,
	username string,
	externalIdentityProvider *string,
	externalIdentityID *string,

) (Registration, error) {

	var errs error

	email, err := identity.ParseEmail(emailStr)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	creds, err := identity.NewUserCreds(firstName, lastName, username)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	var passwordHashPointer *identity.PasswordHash
	if passwordHashStr != nil {
		if externalIdentityID == nil || externalIdentityProvider == nil {
			errs = errors.Join(errs, errors.New("invalid registration state: has external identity and password"))
		} else {
			hash := identity.PasswordHash(*passwordHashStr)
			passwordHashPointer = &hash
		}
	}

	var externalIdentityPointer *identity.ExternalIdentity
	if externalIdentityID != nil && externalIdentityProvider != nil {
		if passwordHashStr != nil {
			errs = errors.Join(errs, errors.New("invalid registration state: has external identity and password"))
		} else {
			externalIdentity, err := identity.RebuildExternalIdentity(*externalIdentityProvider, *externalIdentityID)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			externalIdentityPointer = &externalIdentity
		}
	}

	if errs != nil {
		return Registration{}, errs
	}

	return Registration{
		onboardID:        onboardID,
		email:            email,
		passwordHash:     passwordHashPointer,
		Creds:            creds,
		externalIdentity: externalIdentityPointer,
	}, nil
}
