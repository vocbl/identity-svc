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

// ======================================================
// CheckUsernameAvailability
// ======================================================

func TestOnboardService_CheckUsernameAvailability(t *testing.T) {
	t.Parallel()

	t.Run("returns error when username empty", testCheckUsernameAvailabilityInvalidUsername)
	t.Run("returns error when repo fails", testCheckUsernameAvailabilityRepoError)
	t.Run("returns false when exists", testCheckUsernameAvailabilityExists)
	t.Run("returns true when not exists", testCheckUsernameAvailabilityNotExists)
}

func testCheckUsernameAvailabilityInvalidUsername(t *testing.T) {
	t.Parallel()

	svc, _ := newOnboardService(t, false)

	ok, err := svc.CheckUsernameAvailability(t.Context(), "")
	require.Error(t, err)
	assert.False(t, ok)
	assert.ErrorIs(t, err, app.ErrUsernameAvailabilityCheckOp)
	assert.ErrorIs(t, err, domain.ErrInvalidUsername)
}

func testCheckUsernameAvailabilityRepoError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, false)
	expectedErr := errors.New("db timeout")
	expectVerificationRepoCheckUsernameExistanceCall(m, false, expectedErr)

	ok, err := svc.CheckUsernameAvailability(t.Context(), "john")
	require.Error(t, err)
	assert.False(t, ok)
	assert.ErrorIs(t, err, app.ErrUsernameAvailabilityCheckOp)
	assert.ErrorIs(t, err, expectedErr)
}

func testCheckUsernameAvailabilityExists(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, false)
	expectVerificationRepoCheckUsernameExistanceCall(m, true, nil)

	ok, err := svc.CheckUsernameAvailability(t.Context(), "john")
	require.NoError(t, err)
	assert.False(t, ok)
}

func testCheckUsernameAvailabilityNotExists(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, false)
	expectVerificationRepoCheckUsernameExistanceCall(m, false, nil)

	ok, err := svc.CheckUsernameAvailability(t.Context(), "john")
	require.NoError(t, err)
	assert.True(t, ok)
}

// ======================================================
// CreateVerificationSession
// ======================================================

func TestOnboardService_CreateVerificationSession(t *testing.T) {
	t.Parallel()

	t.Run("succeeds", testCreateVerificationSessionSuccess)
	t.Run("returns validation error when input invalid", testCreateVerificationSessionInvalidInput)
	t.Run("returns error when duplicate email and username", testCreateVerificationSessionDuplicateEmailAndUsername)
	t.Run("returns error when unexpected Create repo error", testCreateVerificationSessionUnexpectedCreateError)
	t.Run("returns error when EmitVerificationSessionCreatedEvent fails", testCreateVerificationSessionEmitEventError)
	t.Run("returns error when transaction fails", testCreateVerificationSessionTransactionError)
}

func testCreateVerificationSessionSuccess(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	expectVerificationRepoCreateCall(m, nil)
	expectVerificationRepoEmitVerificationSessionCreatedEventCall(m, nil)

	sessionID, err := svc.CreateVerificationSession(t.Context(), validNewUser())
	require.NoError(t, err)
	assert.False(t, sessionID.IsNil())
}

func testCreateVerificationSessionInvalidInput(t *testing.T) {
	t.Parallel()

	svc, _ := newOnboardService(t, false)

	sessionID, err := svc.CreateVerificationSession(t.Context(), app.NewUser{})
	require.Error(t, err)
	assert.True(t, sessionID.IsNil())
	assert.ErrorIs(t, err, app.ErrSessionCreateOp)
}

func testCreateVerificationSessionDuplicateEmailAndUsername(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	expectVerificationRepoCreateCall(m, errors.Join(db.ErrDublicateEmail, db.ErrDublicateUsername))

	sessionID, err := svc.CreateVerificationSession(t.Context(), validNewUser())
	require.Error(t, err)
	assert.True(t, sessionID.IsNil())
	assert.ErrorIs(t, err, app.ErrSessionCreateOp)
	assert.ErrorIs(t, err, app.ErrConflictEmail)
	assert.ErrorIs(t, err, app.ErrConflictUsername)

}

func testCreateVerificationSessionUnexpectedCreateError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	unexpectedErr := errors.New("db timeout")
	expectVerificationRepoCreateCall(m, unexpectedErr)

	sessionID, err := svc.CreateVerificationSession(t.Context(), validNewUser())
	require.Error(t, err)
	assert.True(t, sessionID.IsNil())
	assert.ErrorIs(t, err, app.ErrSessionCreateOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testCreateVerificationSessionEmitEventError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	unexpectedErr := errors.New("event bus down")
	expectVerificationRepoCreateCall(m, nil)
	expectVerificationRepoEmitVerificationSessionCreatedEventCall(m, unexpectedErr)

	sessionID, err := svc.CreateVerificationSession(t.Context(), validNewUser())
	require.Error(t, err)
	assert.True(t, sessionID.IsNil())
	assert.ErrorIs(t, err, app.ErrSessionCreateOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testCreateVerificationSessionTransactionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	m := mock.NewMockVerificationRepo(ctrl)
	svc, _ := app.NewOnboardService(m, newValidVerificationCfg())
	expectedErr := errors.New("tx failure")
	m.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).Return(expectedErr)

	sessionID, err := svc.CreateVerificationSession(t.Context(), validNewUser())
	require.Error(t, err)
	assert.True(t, sessionID.IsNil())
	assert.ErrorIs(t, err, app.ErrSessionCreateOp)
	assert.ErrorIs(t, err, expectedErr)
}

// ======================================================
// RestartVerificationSession
// ======================================================

func TestOnboardService_RestartVerificationSession(t *testing.T) {
	t.Parallel()

	t.Run("succeeds", testRestartVerificationSessionSuccess)
	t.Run("returns error when id is invalid", testRestartVerificationSessionInvalidID)
	t.Run("returns error when user is not found", testRestartVerificationSessionNotFound)
	t.Run("returns error when Get fails unexpectedly", testRestartVerificationSessionGetUnexpectedError)
	t.Run("returns error when session.ResetToken fails", testRestartVerificationSessionResetTokenError)
	t.Run("returns error when Update fails unexpectedly", testRestartVerificationSessionUpdateUnexpectedError)
	t.Run("returns error when Emit event fails unexpectedly", testRestartVerificationSessionEmitUnexpectedError)
	t.Run("returns error when transaction fails", testRestartVerificationSessionTransactionError)
}
func testRestartVerificationSessionSuccess(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, domain.NewValidTestInactiveUserVerificationSession(), nil)
	expectVerificationRepoUpdateCall(m, nil)
	expectVerificationRepoEmitVerificationSessionStartedEventCall(m, nil)

	err := svc.RestartVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.NoError(t, err)
}

func testRestartVerificationSessionInvalidID(t *testing.T) {
	t.Parallel()

	svc, _ := newOnboardService(t, false)

	err := svc.RestartVerificationSession(t.Context(), domain.TestInvalidID)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionRestartOp)
	assert.ErrorIs(t, err, domain.ErrInvalidIdentifier)
}

func testRestartVerificationSessionNotFound(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, nil, db.ErrNonExistingData)

	err := svc.RestartVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionRestartOp)
	assert.ErrorIs(t, err, app.ErrNotFoundVerificationSession)
}

func testRestartVerificationSessionGetUnexpectedError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	unexpectedErr := errors.New("db timeout")
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, nil, unexpectedErr)

	err := svc.RestartVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionRestartOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testRestartVerificationSessionResetTokenError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, domain.NewValidTestActiveUserVerificationSession(), nil)

	err := svc.RestartVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionRestartOp)
}

func testRestartVerificationSessionUpdateUnexpectedError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	unexpectedErr := errors.New("update failed")
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, domain.NewValidTestInactiveUserVerificationSession(), nil)
	expectVerificationRepoUpdateCall(m, unexpectedErr)

	err := svc.RestartVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionRestartOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testRestartVerificationSessionEmitUnexpectedError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	unexpectedErr := errors.New("event bus down")
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, domain.NewValidTestInactiveUserVerificationSession(), nil)
	expectVerificationRepoUpdateCall(m, nil)
	expectVerificationRepoEmitVerificationSessionStartedEventCall(m, unexpectedErr)

	err := svc.RestartVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionRestartOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testRestartVerificationSessionTransactionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	m := mock.NewMockVerificationRepo(ctrl)
	svc, _ := app.NewOnboardService(m, newValidVerificationCfg())
	expectedErr := errors.New("tx error")
	m.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).Return(expectedErr)

	err := svc.RestartVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionRestartOp)
	assert.ErrorIs(t, err, app.ErrSessionRestartOp)
	assert.ErrorIs(t, err, expectedErr)
}

// ======================================================
// DeleteVerificationSession
// ======================================================

func TestOnboardService_DeleteVerificationSession(t *testing.T) {
	t.Parallel()

	t.Run("succeeds", testDeleteVerificationSessionSuccess)
	t.Run("returns error when id invalid", testDeleteVerificationSessionInvalidID)
	t.Run("returns error when not found", testDeleteVerificationSessionNotFound)
	t.Run("returns error when repo.Delete fails", testDeleteVerificationRepoFail)
}

func testDeleteVerificationSessionSuccess(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, false)
	expectVerificationRepoDeleteCall(m, domain.TestValidVerificationSessionID, nil)

	err := svc.DeleteVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.NoError(t, err)
}

func testDeleteVerificationSessionInvalidID(t *testing.T) {
	t.Parallel()

	svc, _ := newOnboardService(t, false)

	err := svc.DeleteVerificationSession(t.Context(), "invalid")
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionDeleteOp)
}

func testDeleteVerificationSessionNotFound(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, false)
	expectVerificationRepoDeleteCall(m, domain.TestValidVerificationSessionID, db.ErrNonExistingData)

	err := svc.DeleteVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionDeleteOp)
	assert.ErrorIs(t, err, app.ErrNotFoundVerificationSession)
}

func testDeleteVerificationRepoFail(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, false)
	unexpectedErr := errors.New("connection error")
	expectVerificationRepoDeleteCall(m, domain.TestValidVerificationSessionID, unexpectedErr)

	err := svc.DeleteVerificationSession(t.Context(), domain.TestValidVerificationSessionID.String())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionDeleteOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

// ======================================================
// CompleteVerification
// ======================================================

// ======================================================
// CompleteVerification
// ======================================================

func TestOnboardService_CompleteVerification(t *testing.T) {
	t.Parallel()

	t.Run("succeeds", testCompleteVerificationSuccess)
	t.Run("returns error when id invalid", testCompleteVerificationInvalidID)
	t.Run("returns error when token invalid", testCompleteVerificationInvalidToken)
	t.Run("returns error when not found", testCompleteVerificationNotFound)
	t.Run("returns error when repo.Get fails unexpectedly", testCompleteVerificationGetUnexpectedError)
	t.Run("returns error when session.Complete fails", testCompleteVerificationDomainError)
	t.Run("returns error when repo.CreateUser fails unexpectedly", testCompleteVerificationCreateUserError)
	t.Run("returns error when EmitUserVerifiedEvent fails unexpectedly", testCompleteVerificationEmitError)
	t.Run("returns error when transaction fails", testCompleteVerificationTransactionError)
}

func testCompleteVerificationSuccess(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	session := domain.NewValidTestActiveUserVerificationSession()
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, session, nil)
	expectVerificationRepoCreateUserCall(m, nil)
	expectVerificationRepoEmitUserVerifiedEventCall(m, domain.TestValidVerificationSessionID, nil)

	err := svc.CompleteVerification(t.Context(), domain.TestValidVerificationSessionID.String(), domain.TestValidToken)
	require.NoError(t, err)
}

func testCompleteVerificationInvalidID(t *testing.T) {
	t.Parallel()

	svc, _ := newOnboardService(t, false)

	err := svc.CompleteVerification(t.Context(), domain.TestInvalidID, domain.TestValidToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCompleteOp)
}

func testCompleteVerificationInvalidToken(t *testing.T) {
	t.Parallel()

	svc, _ := newOnboardService(t, false)

	err := svc.CompleteVerification(t.Context(), domain.TestValidID, domain.TestInvalidToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCompleteOp)
}

func testCompleteVerificationNotFound(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, nil, db.ErrNonExistingData)

	err := svc.CompleteVerification(t.Context(), domain.TestValidVerificationSessionID.String(), domain.TestValidToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCompleteOp)
	assert.ErrorIs(t, err, app.ErrNotFoundVerificationSession)
}

func testCompleteVerificationGetUnexpectedError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	unexpectedErr := errors.New("db timeout")
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, nil, unexpectedErr)

	err := svc.CompleteVerification(t.Context(), domain.TestValidVerificationSessionID.String(), domain.TestValidToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCompleteOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testCompleteVerificationDomainError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, domain.NewValidTestInactiveUserVerificationSession(), nil)

	err := svc.CompleteVerification(t.Context(), domain.TestValidVerificationSessionID.String(), domain.TestValidToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCompleteOp)
}

func testCompleteVerificationCreateUserError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	session := domain.NewValidTestActiveUserVerificationSession()
	unexpectedErr := errors.New("insert failed")
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, session, nil)
	expectVerificationRepoCreateUserCall(m, unexpectedErr)

	err := svc.CompleteVerification(t.Context(), domain.TestValidVerificationSessionID.String(), domain.TestValidToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCompleteOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testCompleteVerificationEmitError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, true)
	session := domain.NewValidTestActiveUserVerificationSession()
	unexpectedErr := errors.New("event bus down")
	expectVerificationRepoGetCall(m, domain.TestValidVerificationSessionID, session, nil)
	expectVerificationRepoCreateUserCall(m, nil)
	expectVerificationRepoEmitUserVerifiedEventCall(m, domain.TestValidVerificationSessionID, unexpectedErr)

	err := svc.CompleteVerification(t.Context(), domain.TestValidVerificationSessionID.String(), domain.TestValidToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCompleteOp)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testCompleteVerificationTransactionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	m := mock.NewMockVerificationRepo(ctrl)
	svc, _ := app.NewOnboardService(m, newValidVerificationCfg())
	expectedErr := errors.New("tx failure")
	m.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).Return(expectedErr)

	err := svc.CompleteVerification(t.Context(), domain.TestValidVerificationSessionID.String(), domain.TestValidToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCompleteOp)
	assert.ErrorIs(t, err, expectedErr)
}

// ======================================================
// CleanUpVerificationSessions
// ======================================================

func TestOnboardService_CleanUpVerificationSessions(t *testing.T) {
	t.Parallel()

	t.Run("succeeds", testCleanUpVerificationSessionsSuccess)
	t.Run("returns error when repo fails", testCleanUpVerificationSessionsRepoError)
}

func testCleanUpVerificationSessionsSuccess(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, false)
	expectVerificationRepoCleanCall(m, domain.TestValidUserVerificationCleanUpDuration, nil)

	err := svc.CleanUpVerificationSessions(t.Context())
	require.NoError(t, err)
}

func testCleanUpVerificationSessionsRepoError(t *testing.T) {
	t.Parallel()

	svc, m := newOnboardService(t, false)
	expectedErr := errors.New("cleanup failed")
	expectVerificationRepoCleanCall(m, domain.TestValidUserVerificationCleanUpDuration, expectedErr)

	err := svc.CleanUpVerificationSessions(t.Context())
	require.Error(t, err)
	assert.ErrorIs(t, err, app.ErrSessionCleanupOp)
	assert.ErrorIs(t, err, expectedErr)
}
