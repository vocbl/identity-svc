package app_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	app "github.com/vocbl/users-svc/internal/application/onboard"
	"github.com/vocbl/users-svc/internal/application/onboard/mock"
	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
	"go.uber.org/mock/gomock"
)

func TestVerificationStarterService_StartVerificationSession(t *testing.T) {
	t.Parallel()

	t.Run("succeeds when passed data is valid", testStartVerificationSucess)
	t.Run("returns error when input data is invalid", testStartVerificationInvalidInput)
	t.Run("returns error when password hashing fails", testStartVerificationPasswordHashError)
	t.Run("returns error when session does not exist", testStartVerificationNotFound)
	t.Run("returns error when repo.Get fails", testStartVerificationRepoGetError)
	t.Run("returns error when session.Start fails", testStartVerificationSessionAlreadyStartedError)
	t.Run("returns error when repo.Update fails", testStartVerificationRepoUpdateError)
	t.Run("returns error when repo.Emit fails", testStartVerificationRepoEmitError)
	t.Run("returns error when transaction itself fails", testStartVerificationTransactionError)
}

// ============ Sub-Functions ============

func testStartVerificationSucess(t *testing.T) {
	t.Parallel()

	svc, m := newVerificationStarterService(t, true)
	session := domain.NewValidTestUserVerificationSession()

	expectVerificationStarterRepoGetCall(m, domain.TestValidVerificationSessionID, session, nil)
	expectVerificationStarterRepoUpdateCall(m, nil)
	expectVerificationStarterRepoEmitVerificationSessionStartedEventCall(m, nil)

	err := svc.StartVerificationSession(
		t.Context(),
		domain.TestValidVerificationSessionID.String(),
		string(domain.TestValidEmail),
		string(domain.TestValidPassword),
	)

	require.NoError(t, err)
}

func testStartVerificationInvalidInput(t *testing.T) {
	t.Parallel()

	svc, _ := newVerificationStarterService(t, false)

	err := svc.StartVerificationSession(
		t.Context(),
		domain.TestInvalidID,
		domain.TestInvalidEmail,
		domain.TestInvalidPassword,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionStartOp)
}

func testStartVerificationPasswordHashError(t *testing.T) {
	t.Parallel()

	m := mock.NewMockVerificationStarterRepo(gomock.NewController(t))
	cfg := newValidVerificationCfg()

	fail := false
	expectedErr := errors.New("cpu exoustion")
	cfg.PasswordHasher = func(s string) (string, error) {
		if fail {
			return "", expectedErr
		}
		return domain.TestValidPasswordHash, nil
	}

	svc, err := app.NewVerificationStarterService(m, cfg)
	require.NoError(t, err)

	fail = true
	err = svc.StartVerificationSession(
		t.Context(),
		domain.TestValidVerificationSessionID.String(),
		domain.TestValidEmail,
		domain.TestValidPassword,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionStartOp)
	assert.ErrorIs(t, err, expectedErr)
}

func testStartVerificationNotFound(t *testing.T) {
	svc, m := newVerificationStarterService(t, true)

	expectVerificationStarterRepoGetCall(
		m,
		domain.TestValidVerificationSessionID,
		nil,
		db.ErrNonExistingData,
	)

	err := svc.StartVerificationSession(
		t.Context(),
		domain.TestValidVerificationSessionID.String(),
		string(domain.TestValidEmail),
		string(domain.TestValidPassword),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionStartOp)
	assert.ErrorIs(t, err, app.ErrNotFoundVerificationSession)
}

func testStartVerificationRepoGetError(t *testing.T) {
	svc, m := newVerificationStarterService(t, true)
	unexpectedErr := errors.New("db connection timeout")

	expectVerificationStarterRepoGetCall(
		m,
		domain.TestValidVerificationSessionID,
		nil,
		unexpectedErr,
	)

	err := svc.StartVerificationSession(
		t.Context(),
		domain.TestValidVerificationSessionID.String(),
		string(domain.TestValidEmail),
		string(domain.TestValidPassword),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionStartOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testStartVerificationSessionAlreadyStartedError(t *testing.T) {
	svc, m := newVerificationStarterService(t, true)

	expectVerificationStarterRepoGetCall(
		m,
		domain.TestValidVerificationSessionID,
		domain.NewValidTestActiveUserVerificationSession(),
		nil,
	)

	err := svc.StartVerificationSession(
		t.Context(),
		domain.TestValidVerificationSessionID.String(),
		string(domain.TestValidEmail),
		string(domain.TestValidPassword),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionStartOp)
	assert.ErrorIs(t, err, domain.ErrSessionAlreadyStarted)
}

func testStartVerificationRepoUpdateError(t *testing.T) {
	svc, m := newVerificationStarterService(t, true)
	session := domain.NewValidTestUserVerificationSession()
	unexpectedErr := errors.New("concurrent update conflict")

	expectVerificationStarterRepoGetCall(
		m,
		domain.TestValidVerificationSessionID,
		session,
		nil,
	)
	expectVerificationStarterRepoUpdateCall(m, unexpectedErr)

	err := svc.StartVerificationSession(
		t.Context(),
		domain.TestValidVerificationSessionID.String(),
		string(domain.TestValidEmail),
		string(domain.TestValidPassword),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionStartOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testStartVerificationRepoEmitError(t *testing.T) {
	svc, m := newVerificationStarterService(t, true)
	session := domain.NewValidTestUserVerificationSession()
	unexpectedErr := errors.New("nats publisher error")

	expectVerificationStarterRepoGetCall(
		m,
		domain.TestValidVerificationSessionID,
		session,
		nil,
	)
	expectVerificationStarterRepoUpdateCall(m, nil)
	expectVerificationStarterRepoEmitVerificationSessionStartedEventCall(m, unexpectedErr)

	err := svc.StartVerificationSession(
		t.Context(),
		domain.TestValidVerificationSessionID.String(),
		string(domain.TestValidEmail),
		string(domain.TestValidPassword),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionStartOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testStartVerificationTransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mock.NewMockVerificationStarterRepo(ctrl)
	svc, _ := app.NewVerificationStarterService(m, newValidVerificationCfg())

	unexpectedErr := errors.New("tx rollback failed")

	m.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		Return(unexpectedErr)

	err := svc.StartVerificationSession(
		t.Context(),
		domain.TestValidVerificationSessionID.String(),
		string(domain.TestValidEmail),
		string(domain.TestValidPassword),
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionStartOp)
	assert.ErrorIs(t, err, unexpectedErr)
}
