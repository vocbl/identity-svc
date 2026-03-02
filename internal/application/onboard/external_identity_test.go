package app_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	app "github.com/vocbl/users-svc/internal/application/onboard"
	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
	"go.uber.org/mock/gomock"
)

func TestExternalIdentityService_CreateUserFromExternalIdentity(t *testing.T) {
	t.Parallel()

	t.Run("succeeds when passed data is valid", testCreateExternalUserSuccess)
	t.Run("returns validation error when input data is invalid", testCreateExternalUserInvalidInput)
	t.Run("returns error when AddExternalIdentity fails", testCreateExternalUserAddExternalIdentityFails)
	t.Run("returns error when duplicate email", testCreateExternalUserDuplicateEmail)
	t.Run("regenerates username on duplicate username and succeeds", testCreateExternalUserDuplicateUsernameThenSuccess)
	t.Run("returns error when username exhaustion reached", testCreateExternalUserUsernameExhaustion)
	t.Run("returns error when repo.Create fails unexpectedly", testCreateExternalUserUnexpectedError)
}

// ============ CreateUser Sub-Functions ============

func testCreateExternalUserSuccess(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, false)
	expectUserRepoCreateCall(m, nil)

	err := svc.CreateUserFromExternalIdentity(t.Context(), domain.AuthProviderGoogle, validOAuthUser())
	require.NoError(t, err)
}

func testCreateExternalUserInvalidInput(t *testing.T) {
	t.Parallel()

	svc, _ := newExternalIdentityService(t, false)
	ou := app.OAuthUser{
		Email:      domain.TestInvalidEmail,
		FirstName:  "",
		LastName:   "",
		ExternalID: "",
	}

	err := svc.CreateUserFromExternalIdentity(t.Context(), domain.AuthProviderGoogle, ou)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalUserCreateOp)
}

func testCreateExternalUserAddExternalIdentityFails(t *testing.T) {
	t.Parallel()

	svc, _ := newExternalIdentityService(t, false)

	ou := validOAuthUser()
	ou.ExternalID = ""

	err := svc.CreateUserFromExternalIdentity(t.Context(), domain.AuthProviderGoogle, ou)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalUserCreateOp)
	assert.ErrorIs(t, err, domain.ErrInvalidIdentifier)
}

func testCreateExternalUserDuplicateEmail(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, false)
	expectUserRepoCreateCall(m, db.ErrDublicateEmail)

	err := svc.CreateUserFromExternalIdentity(t.Context(), domain.AuthProviderGoogle, validOAuthUser())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalUserCreateOp)
	assert.ErrorIs(t, err, app.ErrConflictEmail)
}

func testCreateExternalUserDuplicateUsernameThenSuccess(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, false)
	gomock.InOrder(
		expectUserRepoCreateCall(m, db.ErrDublicateUsername),
		expectUserRepoCreateCall(m, nil),
	)

	err := svc.CreateUserFromExternalIdentity(t.Context(), domain.AuthProviderGoogle, validOAuthUser())
	require.NoError(t, err)
}

func testCreateExternalUserUsernameExhaustion(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, false)
	for i := 0; i < 5; i++ {
		expectUserRepoCreateCall(m, db.ErrDublicateUsername)
	}

	err := svc.CreateUserFromExternalIdentity(t.Context(), domain.AuthProviderGoogle, validOAuthUser())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalUserCreateOp)
	assert.ErrorIs(t, err, app.ErrUsernameExhaustion)
}

func testCreateExternalUserUnexpectedError(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, false)
	unexpectedErr := errors.New("db timeout")
	expectUserRepoCreateCall(m, unexpectedErr)

	err := svc.CreateUserFromExternalIdentity(t.Context(), domain.AuthProviderGoogle, validOAuthUser())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalUserCreateOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func TestExternalIdentityService_AtachUserExternalIdentity(t *testing.T) {
	t.Parallel()

	t.Run("succeeds when data is valid", testAttachExternalIdentitySuccess)
	t.Run("returns error when email is invalid", testAttachExternalIdentityInvalidEmail)
	t.Run("returns error when user not found", testAttachExternalIdentityNotFound)
	t.Run("returns error when duplicate external identity", testAttachExternalIdentityConflict)
	t.Run("returns error when AddExternalIdentity fails inside transaction", testAttachExternalIdentityAddFails)
	t.Run("returns error when unexpected repo error occurs", testAttachExternalIdentityUnexpectedError)
}

// ============ Attach Sub-Functions ============

func testAttachExternalIdentitySuccess(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, true)
	expectUserRepoGetByEmailCall(m, domain.TestValidEmail, domain.NewValidTestUser(), nil)
	expectUserRepoUpdateCall(m, nil)

	err := svc.AtachUserExternalIdentity(
		t.Context(),
		string(domain.TestValidEmail),
		domain.AuthProviderGoogle,
		"external-id",
	)

	require.NoError(t, err)
}

func testAttachExternalIdentityInvalidEmail(t *testing.T) {
	t.Parallel()

	svc, _ := newExternalIdentityService(t, false)

	err := svc.AtachUserExternalIdentity(
		t.Context(),
		"invalid-email",
		domain.AuthProviderGoogle,
		"external-id",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalIdentityAttachOp)
}

func testAttachExternalIdentityAddFails(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, true)
	expectUserRepoGetByEmailCall(m, domain.TestValidEmail, domain.NewValidTestUser(), nil)

	err := svc.AtachUserExternalIdentity(
		t.Context(),
		string(domain.TestValidEmail),
		domain.AuthProviderGoogle,
		"",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalIdentityAttachOp)
	assert.ErrorIs(t, err, domain.ErrInvalidIdentifier)
}

func testAttachExternalIdentityNotFound(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, false)
	m.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		Return(db.ErrNonExistingData)

	err := svc.AtachUserExternalIdentity(
		t.Context(),
		string(domain.TestValidEmail),
		domain.AuthProviderGoogle,
		"external-id",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalIdentityAttachOp)
	assert.ErrorIs(t, err, app.ErrNotFoundUser)
}

func testAttachExternalIdentityConflict(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, false)
	m.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		Return(db.ErrDublicateData)

	err := svc.AtachUserExternalIdentity(
		t.Context(),
		string(domain.TestValidEmail),
		domain.AuthProviderGoogle,
		"external-id",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalIdentityAttachOp)
	assert.ErrorIs(t, err, app.ErrConflictExternalIdentity)
}

func testAttachExternalIdentityUnexpectedError(t *testing.T) {
	t.Parallel()

	svc, m := newExternalIdentityService(t, false)
	unexpectedErr := errors.New("tx failure")

	m.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		Return(unexpectedErr)

	err := svc.AtachUserExternalIdentity(
		t.Context(),
		string(domain.TestValidEmail),
		domain.AuthProviderGoogle,
		"external-id",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrExternalIdentityAttachOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

// ============ Helpers ============

func validOAuthUser() app.OAuthUser {
	return app.OAuthUser{
		Email:      string(domain.TestValidEmail),
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "external-id",
	}
}
