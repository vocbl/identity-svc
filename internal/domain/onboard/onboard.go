package onboard

import (
	"time"

	identity "github.com/vocbl/users-svc/internal/domain"
)

type OnboardID struct {
	identity.ID
}

func ParseOnboardID(s string) (OnboardID, error) {
	val, err := identity.ParseID(s)
	return OnboardID{val}, err
}

type OnboardSession struct {
	id           OnboardID
	Registration Registration
	Verification Verification
	Integration  Integration
	createdAt    time.Time
	expiresAt    time.Time
}

func (o *OnboardSession) ID() OnboardID {
	return o.id
}

func (o *OnboardSession) Complete(policy *IntegrationPolicy) (*identity.User, error) {
	if o.expiresAt.Before(time.Now().UTC()) {
		return nil, ErrOnboardSessionExpired
	} else if !o.Integration.IsCompleted(policy) {
		return nil, ErrIntegrationNotCompleted
	} else if !o.Verification.IsCompleted() {
		return nil, ErrVerificationNotCompleted
	}

	user := identity.NewUser(o.Registration.email, *o.Registration.passwordHash, o.Registration.Creds)
	if o.Registration.externalIdentity != nil {
		user.AddExternalIdentity(*o.Registration.externalIdentity)
	}

	return user, nil
}

type Factory struct {
	registrationPolicy *RegistrationPolicy
	verificationPolicy *VerificationPolicy
	integrationPolicy  *IntegrationPolicy
	onboardPolicy      *OnboardPolicy
}

func NewOnboardFactory(rp *RegistrationPolicy, vp *VerificationPolicy, ip *IntegrationPolicy, op *OnboardPolicy) *Factory {
	return &Factory{rp, vp, ip, op}
}

func (f *Factory) NewOnboarding(email, password, firstName, lastName, username string) (*OnboardSession, VerificationToken, error) {
	onboard := OnboardSession{
		id:        OnboardID{identity.NewID()},
		expiresAt: time.Now().UTC().Add(f.onboardPolicy.expirationDuration),
	}

	var err error
	onboard.Registration, err = newRegistration(f.registrationPolicy, email, password, firstName, lastName, username)
	if err != nil {
		return nil, "", err
	}

	var token VerificationToken
	onboard.Verification, token = newEmailVerification(f.verificationPolicy, onboard.id)

	onboard.Integration = newIntegration(f.integrationPolicy, onboard.id)

	return &onboard, token, nil
}

func (f *Factory) NewOnboardingFromExternalIdentity(email, firstName, lastName string, externalProvider identity.ExternalProvider, externalId string) (*OnboardSession, error) {
	onboard := OnboardSession{
		id: OnboardID{identity.NewID()},
	}

	var err error
	onboard.Registration, err = newRegistrationFromExternalIdentity(externalProvider, externalId, email, firstName, lastName)
	if err != nil {
		return nil, err
	}

	onboard.Verification = newCompletedVerificationFromExternalIdentity(onboard.id)
	onboard.Integration = newIntegration(f.integrationPolicy, onboard.id)

	return &onboard, nil
}

func RebuildOnboardSession(id OnboardID, expiresAt, createdAt time.Time, registration Registration, verification Verification, integration Integration) *OnboardSession {
	return &OnboardSession{
		id:           id,
		createdAt:    createdAt,
		expiresAt:    expiresAt,
		Registration: registration,
		Verification: verification,
		Integration:  integration,
	}
}
