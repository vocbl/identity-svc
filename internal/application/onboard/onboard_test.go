package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oklog/ulid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	errorutil "github.com/vocbl/shared/errors"
	timetest "github.com/vocbl/shared/test/time"
	ulidtest "github.com/vocbl/shared/test/ulid"
	ulidutil "github.com/vocbl/shared/utils/ulid"
	app "github.com/vocbl/users-svc/internal/application/onboard"
	"github.com/vocbl/users-svc/internal/domain"
	db "github.com/vocbl/users-svc/internal/infrastructure/persistance"
	"github.com/vocbl/users-svc/internal/shared/security"
)

//
//
//
//
// ============ NewOnboardService =============

func TestNewOnboardService(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates service when all parameters are valid", testNewOnboardService_Success)
	t.Run("returns error when config validation fails", testNewOnboardService_InvalidConfig)
	t.Run("returns error when token generator produces invalid token", testNewOnboardService_InvalidToken)
	t.Run("returns error when token generator produces invalid token hash", testNewOnboardService_InvalidTokenHash)
	t.Run("returns error when password hasher fails", testNewOnboardService_PasswordHasherError)
	t.Run("returns error when password hasher produces invalid hash", testNewOnboardService_InvalidPasswordHash)
}

func testNewOnboardService(
	t *testing.T,
	verificationPolicy domain.VerificationPolicy,
	hasher security.PasswordHasher,
	gen security.TokenGenerator,
) (*app.OnboardService, error) {
	t.Helper()
	repo := newRepoMock(t)
	return app.NewOnboardService(repo, verificationPolicy, hasher, gen)
}

func testNewOnboardService_Success(t *testing.T) {
	t.Parallel()

	service, err := testNewOnboardService(t,
		domain.TestValidVerificationPolicy(),
		security.HashPassword,
		security.NewTokenGenerator(domain.TokenLenght),
	)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func testNewOnboardService_InvalidConfig(t *testing.T) {
	t.Parallel()

	cfg := domain.TestValidVerificationPolicy()
	cfg.MaxAttempts = domain.TestInvalidUserVerificationMaxAttempts

	_, err := testNewOnboardService(t, cfg, security.HashPassword, security.NewTokenGenerator(domain.TokenLenght))

	assert.Error(t, err)
}

func testNewOnboardService_InvalidToken(t *testing.T) {
	t.Parallel()

	_, err := testNewOnboardService(t,
		domain.TestValidVerificationPolicy(),
		security.HashPassword,
		func() (string, string) { return domain.TestInvalidToken, domain.TestValidTokenHash },
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tokenGenerator")
}

func testNewOnboardService_InvalidTokenHash(t *testing.T) {
	t.Parallel()

	_, err := testNewOnboardService(t,
		domain.TestValidVerificationPolicy(),
		security.HashPassword,
		func() (string, string) { return domain.TestValidToken, domain.TestInvalidTokenHash },
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tokenGenerator")
}

func testNewOnboardService_PasswordHasherError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("internal hasher failure")
	_, err := testNewOnboardService(t,
		domain.TestValidVerificationPolicy(),
		func(p string) (string, error) { return "", expectedErr },
		security.NewTokenGenerator(domain.TokenLenght),
	)

	assert.ErrorIs(t, err, expectedErr)
	assert.Contains(t, err.Error(), "invalid passwordHasher")
}

func testNewOnboardService_InvalidPasswordHash(t *testing.T) {
	t.Parallel()

	_, err := testNewOnboardService(t,
		domain.TestValidVerificationPolicy(),
		func(p string) (string, error) { return domain.TestInvalidPasswordHash, nil },
		security.NewTokenGenerator(domain.TokenLenght),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid passwordHasher")
}

//
//
//
//
// ============ CheckUsernameAvailability =============

func TestOnboardService_CheckUsernameAvailability(t *testing.T) {
	t.Parallel()

	t.Run("returns true when username is available", testOnboardService_CheckUsernameAvailability_Available)
	t.Run("returns false when username is taken", testOnboardService_CheckUsernameAvailability_Taken)
	t.Run("returns error when repository fails", testOnboardService_CheckUsernameAvailability_RepositoryError)
	t.Run("returns error when username is blank", testOnboardService_CheckUsernameAvailability_BlankUsername)
}

func testCheckUsernameAvailability(t *testing.T, username string, existsInRepo bool, repoError error) (bool, error) {
	t.Helper()

	service, mock := newOnboardService(t, false)
	expectCheckUsernameExistanceCall(mock, username, existsInRepo, repoError)
	return service.CheckUsernameAvailability(context.Background(), username)
}

func testOnboardService_CheckUsernameAvailability_Available(t *testing.T) {
	t.Parallel()

	available, err := testCheckUsernameAvailability(t, "JohnDoe", false, nil)
	assert.NoError(t, err)
	assert.True(t, available)
}

func testOnboardService_CheckUsernameAvailability_Taken(t *testing.T) {
	t.Parallel()

	available, err := testCheckUsernameAvailability(t, "JohnDoe", true, nil)
	assert.NoError(t, err)
	assert.False(t, available)
}

func testOnboardService_CheckUsernameAvailability_RepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("unexpected repo error")

	available, err := testCheckUsernameAvailability(t, "JohnDoe", false, repoErr)
	assert.False(t, available)
	assert.ErrorIs(t, err, app.ErrCheckUsernameAvailabilityOperation)
}

func testOnboardService_CheckUsernameAvailability_BlankUsername(t *testing.T) {
	t.Parallel()

	service, _ := newOnboardService(t, false)

	available, err := service.CheckUsernameAvailability(context.Background(), "")
	assert.False(t, available)
	assert.ErrorIs(t, err, domain.ErrInvalidUsername)
}

//
//
//
//
// ===================== CreateVerificationSession =======================

func TestNewUser_ToVerificationSession(t *testing.T) {
	t.Parallel()

	validNewUser := validNewUser()
	session, password := validNewUser.ToVerificationSession()

	assert.NotNil(t, session)
	assert.Equal(t, validNewUser.Email, session.Email)
	assert.Equal(t, validNewUser.Password, password)
	assert.Equal(t, validNewUser.FirstName, session.Creds.FirstName)
	assert.Equal(t, validNewUser.LastName, session.Creds.LastName)
	assert.Equal(t, validNewUser.Username, session.Creds.Username)

	ulidtest.Nil(t, session.SessionID)
	assert.Empty(t, session.TokenHash)
}
func TestOnboardService_CreateVerificationSession(t *testing.T) {
	t.Parallel()

	t.Run("returns error when user data is not valid", testOnboardService_CreateVerificationSession_InvalidUserData)
	t.Run("returns error when email already exists", testOnboardService_CreateVerificationSession_EmailAlreadyExists)
	t.Run("returns error when username already exists", testOnboardService_CreateVerificationSession_UsernameAlreadyExists)
	t.Run("returns sessionId, restartableSince and error when user has pending verification", testOnboardService_CreateVerificationSession_PendingVerification)
	t.Run("returns error when  SaveUserVerificationSession unexpectedly fails", testOnboardService_CreateVerificationSession_UnexpectedUserSavingError)
	t.Run("returns error when EmitUserCreatedEvent unexpectedly fails", testOnboardService_CreateVerificationSession_CreationEventEmitingError)
	t.Run("returns nil when operation succeeds", testOnboardService_CreateVerificationSession_Success)
}

func testOnboardService_CreateVerificationSession_InvalidUserData(t *testing.T) {
	t.Parallel()

	s, _ := newOnboardService(t, false)

	sessionID, restartableAt, err := s.CreateVerificationSession(context.Background(), app.NewUser{
		Email:     domain.TestInvalidEmail,
		Password:  domain.TestInvalidPassword,
		FirstName: domain.TestInvalidFirstName,
		LastName:  domain.TestInvalidLastName,
		Username:  domain.TestInvalidUsername,
	})
	ulidtest.Nil(t, sessionID)
	timetest.Nil(t, restartableAt)

	var errValidation *errorutil.ValidationErr
	require.ErrorAs(t, err, &errValidation)

	assert.Contains(t, errValidation.Errs, domain.ErrInvalidEmail)
	assert.Contains(t, errValidation.Errs, domain.ErrInvalidPassword)
	assert.Contains(t, errValidation.Errs, domain.ErrInvalidFirstName)
	assert.Contains(t, errValidation.Errs, domain.ErrInvalidLastName)
	assert.Contains(t, errValidation.Errs, domain.ErrInvalidUsername)
}

func testSaveUserError(t *testing.T, repoErr, expectedServiceErr error) {
	t.Helper()

	service, mock := newOnboardService(t, true)
	expectSaveUserVerificationSessionCall(mock, repoErr)

	sessionID, restartableAt, err := service.CreateVerificationSession(t.Context(), validNewUser())
	assert.ErrorIs(t, err, expectedServiceErr)
	ulidtest.Nil(t, sessionID)
	timetest.Nil(t, restartableAt)
}

func testOnboardService_CreateVerificationSession_EmailAlreadyExists(t *testing.T) {
	t.Parallel()
	testSaveUserError(t, db.ErrDublicateEmail, app.ErrEmailAlreadyExists)
}

func testOnboardService_CreateVerificationSession_UsernameAlreadyExists(t *testing.T) {
	t.Parallel()
	testSaveUserError(t, db.ErrDublicateUsername, app.ErrUsernameTaken)
}

func testOnboardService_CreateVerificationSession_PendingVerification(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)

	expectedSessionID := ulidutil.NewULID()

	repoErr := db.UserPendingVerificationError{
		SessionID:        expectedSessionID,
		RestartableSince: time.Now().Add(time.Hour * 10),
	}

	expectSaveUserVerificationSessionCall(mock, repoErr)

	sessionID, restartableAt, err := service.CreateVerificationSession(t.Context(), validNewUser())

	assert.ErrorIs(t, err, app.ErrUserPendingVerification)
	assert.Equal(t, expectedSessionID, sessionID)
	assert.True(t, restartableAt.Equal(repoErr.RestartableSince))
}

func testOnboardService_CreateVerificationSession_UnexpectedUserSavingError(t *testing.T) {
	t.Parallel()
	unexpectedErr := errors.New("unexpected repo error")
	testSaveUserError(t, unexpectedErr, unexpectedErr)
}

func testOnboardService_CreateVerificationSession_CreationEventEmitingError(t *testing.T) {
	service, mock := newOnboardService(t, true)
	expectSaveUserVerificationSessionCall(mock, nil)

	validNewUser := validNewUser()
	expectedErr := errors.New("unexpected repo error")
	expectEmitUserCreatedEventCall(mock, validNewUser.Email, validNewUser.Password, expectedErr)

	sessionID, restartableAt, err := service.CreateVerificationSession(t.Context(), validNewUser)
	assert.ErrorIs(t, err, expectedErr)
	ulidtest.Nil(t, sessionID)
	timetest.Nil(t, restartableAt)
}

func testOnboardService_CreateVerificationSession_Success(t *testing.T) {
	validNewUser := validNewUser()
	service, mock := newOnboardService(t, true)
	expectSaveUserVerificationSessionCall(mock, nil)
	expectEmitUserCreatedEventCall(mock, validNewUser.Email, validNewUser.Password, nil)

	sessionID, restartableAt, err := service.CreateVerificationSession(t.Context(), validNewUser)
	assert.NoError(t, err)
	ulidtest.NotNil(t, sessionID)
	timetest.NotNil(t, restartableAt)
}

//
//
//
//
//
// ===================== StartVerificationSession =======================

func TestOnboardService_StartVerificationSession(t *testing.T) {
	t.Parallel()

	t.Run("returns error when passed sessionID is not valid", testOnboardService_StartVerificationSession_InvalidULID)
	t.Run("returns error when user data is not valid", testOnboardService_StartVerificationSession_InvalidUserData)
	t.Run("returns error when hashing fails", testOnboardService_StartVerificationSession_HashFailure)
	t.Run("returns error when session does not exist in DB", testOnboardService_StartVerificationSession_NonExistingSession)
	t.Run("returns error when DB update fails", testOnboardService_StartVerificationSession_UpdateFailure)
	t.Run("returns error when event emission fails", testOnboardService_StartVerificationSession_EmissionFailure)
	t.Run("returns nil when operation succeeds", testOnboardService_StartVerificationSession_Success)
}

func testOnboardService_StartVerificationSession_InvalidULID(t *testing.T) {
	t.Parallel()
	s, _ := newOnboardService(t, false)

	err := s.StartVerificationSession(t.Context(), "invalid-ulid", domain.TestValidEmail, domain.TestValidPassword, domain.TestValidToken)
	assert.ErrorIs(t, err, app.ErrStartVerificationSessionOperation)
	assert.ErrorIs(t, err, app.ErrInvalidULID)
}

func testOnboardService_StartVerificationSession_InvalidUserData(t *testing.T) {
	t.Parallel()
	s, _ := newOnboardService(t, false)

	err := s.StartVerificationSession(t.Context(), ulidutil.NewULID().String(), domain.TestInvalidEmail, domain.TestInvalidPassword, domain.TestInvalidToken)
	var validationErr *errorutil.ValidationErr
	require.ErrorAs(t, err, &validationErr)
	assert.Contains(t, validationErr.Errs, domain.ErrInvalidEmail)
	assert.Contains(t, validationErr.Errs, domain.ErrInvalidPassword)
	assert.Len(t, validationErr.Errs, 3)
}

func testOnboardService_StartVerificationSession_HashFailure(t *testing.T) {
	t.Parallel()
	expectedErr := errors.New("cpu failure")
	returnError := false
	s, err := app.NewOnboardService(
		newRepoMock(t),
		domain.TestValidVerificationPolicy(),
		func(p string) (string, error) {
			if !returnError {
				return security.HashPassword(p)
			}
			return "", expectedErr
		},
		security.NewTokenGenerator(domain.TokenLenght),
	)
	require.NoError(t, err)

	returnError = true
	err = s.StartVerificationSession(t.Context(), ulidutil.NewULID().String(), domain.TestValidEmail, domain.TestValidPassword, domain.TestValidToken)
	assert.ErrorIs(t, err, expectedErr)
}

func testOnboardService_StartVerificationSession_NonExistingSession(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, true)
	id := ulidutil.NewULID()

	expectSetUserVerificationSessionPasswordHashCall(m, id, db.ErrNonExistingData)

	err := s.StartVerificationSession(t.Context(), id.String(), domain.TestValidEmail, domain.TestValidPassword, domain.TestValidToken)
	assert.ErrorIs(t, err, app.ErrNonExistingVerificationSession)
}

func testOnboardService_StartVerificationSession_UpdateFailure(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, true)
	id := ulidutil.NewULID()
	dbErr := errors.New("deadlock")

	expectSetUserVerificationSessionPasswordHashCall(m, id, dbErr)

	err := s.StartVerificationSession(t.Context(), id.String(), domain.TestValidEmail, domain.TestValidPassword, domain.TestValidToken)
	assert.ErrorIs(t, err, dbErr)
}

func testOnboardService_StartVerificationSession_EmissionFailure(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, true)
	id := ulidutil.NewULID()
	natsErr := errors.New("nats unavailable")

	expectSetUserVerificationSessionPasswordHashCall(m, id, nil)
	expectEmitVerifyUserEventCall(m, domain.TestValidEmail, natsErr)

	err := s.StartVerificationSession(t.Context(), id.String(), domain.TestValidEmail, domain.TestValidPassword, domain.TestValidToken)
	assert.ErrorIs(t, err, natsErr)
}

func testOnboardService_StartVerificationSession_Success(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, true)
	id := ulidutil.NewULID()

	expectSetUserVerificationSessionPasswordHashCall(m, id, nil)
	expectEmitVerifyUserEventCall(m, domain.TestValidEmail, nil)

	err := s.StartVerificationSession(t.Context(), id.String(), domain.TestValidEmail, domain.TestValidPassword, domain.TestValidToken)
	assert.NoError(t, err)
}

//
//
//
//
// ===================== RestartVerification =======================

func TestOnboardService_RestartVerification(t *testing.T) {
	t.Parallel()

	t.Run("returns error when passed sessionID is not valid", testOnboardService_RestartVerification_InvalidULID)
	t.Run("returns error when session does not exist", testOnboardService_RestartVerification_NonExistingSession)
	t.Run("returns error when session has already been completed", testOnboardService_RestartVerification_AlreadyCompletedSession)
	t.Run("returns error when session cannot be restarted yet", testOnboardService_RestartVerification_RestartSinceError)
	t.Run("returns error when session DeleteUserVerificationSeiion unexpected error", testOnboardService_RestartVerification_MaxAttemptsError)
	t.Run("returns error when session reaches max attempts", testOnboardService_RestartVerification_MaxAttempts)
	t.Run("returns error when RestartUserVerificationSession unexpectedly fails", testOnboardService_RestartVerification_UnexpectedRestartFailure)
	t.Run("returns error when EmitVerifyUserEvent unexpectedly fails", testOnboardService_RestartVerification_EmissionFailure)
	t.Run("returns nex avalable restart time when succeeds", testOnboardService_RestartVerification_Success)
}

func testOnboardService_RestartVerification_InvalidULID(t *testing.T) {
	t.Parallel()
	service, _ := newOnboardService(t, false)

	restartableSince, err := service.RestartVerification(t.Context(), "invalid-ulid")

	assert.ErrorIs(t, err, app.ErrRestartVerificationSessionOperation)
	assert.ErrorIs(t, err, app.ErrInvalidULID)
	timetest.Nil(t, restartableSince)
}

func testOnboardService_RestartVerification_NonExistingSession(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)
	oldID := ulid.MustNew(ulid.Now(), nil)

	expectRestartUserVerificationSessionCall(mock, oldID, "", 0, db.ErrNonExistingData)

	restartableSince, err := service.RestartVerification(t.Context(), oldID.String())
	assert.ErrorIs(t, err, app.ErrRestartVerificationSessionOperation)
	assert.ErrorIs(t, err, app.ErrNonExistingVerificationSession)
	timetest.Nil(t, restartableSince)
}

func testOnboardService_RestartVerification_AlreadyCompletedSession(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)
	oldID := ulid.MustNew(ulid.Now(), nil)

	expectRestartUserVerificationSessionCall(mock, oldID, "", 0, db.ErrUserVerificationCompleted)

	restartableSince, err := service.RestartVerification(t.Context(), oldID.String())
	assert.ErrorIs(t, err, app.ErrRestartVerificationSessionOperation)
	assert.ErrorIs(t, err, app.ErrVerificationSessionAlreadyCompleted)
	timetest.Nil(t, restartableSince)
}

func testOnboardService_RestartVerification_UnexpectedRestartFailure(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)
	oldID := ulid.MustNew(ulid.Now(), nil)
	unexpectedErr := errors.New("db connection lost")

	expectRestartUserVerificationSessionCall(mock, oldID, "", 0, unexpectedErr)

	restartableSince, err := service.RestartVerification(t.Context(), oldID.String())
	assert.ErrorIs(t, err, app.ErrRestartVerificationSessionOperation)
	assert.ErrorIs(t, err, unexpectedErr)
	timetest.Nil(t, restartableSince)
}

func testOnboardService_RestartVerification_RestartSinceError(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)
	oldID := ulid.MustNew(ulid.Now(), nil)

	expectRestartUserVerificationSessionCall(mock, oldID, "", 0, db.ErrUserVerificationRestartSince)

	restartableSince, err := service.RestartVerification(t.Context(), oldID.String())
	assert.ErrorIs(t, err, app.ErrRestartVerificationSessionOperation)
	assert.ErrorIs(t, err, app.ErrUserVerificationRestartSince)
	timetest.Nil(t, restartableSince)
}

func testOnboardService_RestartVerification_MaxAttemptsError(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)
	sessionID := ulid.MustNew(ulid.Now(), nil)

	expectedErr := errors.New("failed to delete the session")

	expectRestartUserVerificationSessionCall(mock, sessionID, "", domain.TestValidUserVerificationMaxAttempts+1, nil)
	expectDeleteVerificationSessionCall(mock, sessionID, expectedErr)

	restartableSince, err := service.RestartVerification(t.Context(), sessionID.String())
	assert.ErrorIs(t, err, app.ErrRestartVerificationSessionOperation)
	assert.ErrorIs(t, err, expectedErr)
	timetest.Nil(t, restartableSince)
}

func testOnboardService_RestartVerification_MaxAttempts(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)
	sessionID := ulid.MustNew(ulid.Now(), nil)

	expectRestartUserVerificationSessionCall(mock, sessionID, "email@test.com", domain.TestValidUserVerificationMaxAttempts+1, nil)
	expectDeleteVerificationSessionCall(mock, sessionID, nil)

	restartableSince, err := service.RestartVerification(t.Context(), sessionID.String())
	assert.ErrorIs(t, err, app.ErrRestartVerificationSessionOperation)
	assert.ErrorIs(t, err, app.ErrUserVerificationRestartAttempts)
	timetest.Nil(t, restartableSince)
}

func testOnboardService_RestartVerification_EmissionFailure(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)
	oldID := ulid.MustNew(ulid.Now(), nil)
	expectedEmail := "test@example.com"
	emissionErr := errors.New("nats timeout")

	expectRestartUserVerificationSessionCall(mock, oldID, expectedEmail, domain.TestInvalidUserVerificationMaxAttempts-1, nil)
	expectEmitVerifyUserEventCall(mock, expectedEmail, emissionErr)

	restartableSince, err := service.RestartVerification(t.Context(), oldID.String())
	assert.ErrorIs(t, err, app.ErrRestartVerificationSessionOperation)
	assert.ErrorIs(t, err, emissionErr)
	timetest.Nil(t, restartableSince)
}

func testOnboardService_RestartVerification_Success(t *testing.T) {
	t.Parallel()

	service, mock := newOnboardService(t, true)
	oldID := ulid.MustNew(ulid.Now(), nil)
	expectedEmail := "test@example.com"

	expectRestartUserVerificationSessionCall(mock, oldID, expectedEmail, domain.TestInvalidUserVerificationMaxAttempts-1, nil)
	expectEmitVerifyUserEventCall(mock, expectedEmail, nil)

	restartableSince, err := service.RestartVerification(t.Context(), oldID.String())

	assert.NoError(t, err)
	assert.Greater(t, restartableSince, time.Now().UTC())
}

//
//
//
//
// ===================== CompleteVerification =======================

func TestOnboardService_CompleteVerification(t *testing.T) {
	t.Parallel()

	t.Run("returns error when passed invalid token", testOnboardService_CompleteVerification_InvalidToken)
	t.Run("returns error when session does not exist", testOnboardService_CompleteVerification_NonExistingSession)
	t.Run("returns error when session has already been completed", testOnboardService_CompleteVerification_AlreadyCompletedSession)
	t.Run("returns error when session is expired", testOnboardService_CompleteVerification_Expired)
	t.Run("returns error when CompleteUserVerificationSession unexpectedly fails", testOnboardService_CompleteVerification_UnexpectedRepoFailure)
	t.Run("returns error when SaveUser unexpectedly fails", testOnboardService_CompleteVerification_SaveFailure)
	t.Run("returns error when EmitUserVerifiedEvent unexpectedly fails", testOnboardService_CompleteVerification_EmissionFailure)
	t.Run("returns nil when succeeds", testOnboardService_CompleteVerification_Success)
}

func testOnboardService_CompleteVerification_InvalidToken(t *testing.T) {
	t.Parallel()

	s, _ := newOnboardService(t, false)

	err := s.CompleteVerification(t.Context(), domain.TestInvalidToken)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func testOnboardService_CompleteVerification_NonExistingSession(t *testing.T) {
	t.Parallel()

	s, m := newOnboardService(t, true)
	expectCompleteUserVerificationSessionCall(m, nil, db.ErrNonExistingData)

	err := s.CompleteVerification(t.Context(), domain.TestValidToken)
	assert.ErrorIs(t, err, app.ErrNonExistingVerificationSession)
}

func testOnboardService_CompleteVerification_AlreadyCompletedSession(t *testing.T) {
	t.Parallel()

	s, m := newOnboardService(t, true)
	expectCompleteUserVerificationSessionCall(m, nil, db.ErrUserVerificationCompleted)

	err := s.CompleteVerification(t.Context(), domain.TestValidToken)
	assert.ErrorIs(t, err, app.ErrVerificationSessionAlreadyCompleted)
}

func testOnboardService_CompleteVerification_UnexpectedRepoFailure(t *testing.T) {
	t.Parallel()

	s, m := newOnboardService(t, true)
	unexpectedErr := errors.New("db connection failure")
	expectCompleteUserVerificationSessionCall(m, nil, unexpectedErr)

	err := s.CompleteVerification(t.Context(), domain.TestValidToken)
	assert.ErrorIs(t, err, unexpectedErr)
}

func testOnboardService_CompleteVerification_Expired(t *testing.T) {
	t.Parallel()

	s, m := newOnboardService(t, true)
	session := domain.NewValidTestUserVerificationSession(t)
	session.Duration = time.Minute
	session.CreatedAt = time.Now().Add(-time.Hour)
	expectCompleteUserVerificationSessionCall(m, &session, db.ErrSessionExpired)

	err := s.CompleteVerification(t.Context(), domain.TestValidToken)
	assert.ErrorIs(t, err, app.ErrUserVerificationSessionExpired)
}

func testOnboardService_CompleteVerification_SaveFailure(t *testing.T) {
	t.Parallel()

	s, m := newOnboardService(t, true)
	session := domain.NewValidTestUserVerificationSession(t)
	saveErr := errors.New("unique constraint violation on username")
	expectCompleteUserVerificationSessionCall(m, &session, nil)
	expectSaveUserCall(m, saveErr)

	err := s.CompleteVerification(t.Context(), domain.TestValidToken)
	assert.ErrorIs(t, err, app.ErrCompleteVerificationOperation)
	assert.ErrorIs(t, err, saveErr)
}

func testOnboardService_CompleteVerification_EmissionFailure(t *testing.T) {
	t.Parallel()

	s, m := newOnboardService(t, true)
	session := domain.NewValidTestUserVerificationSession(t)
	emitErr := errors.New("nats publisher error")
	expectCompleteUserVerificationSessionCall(m, &session, nil)
	expectSaveUserCall(m, nil)
	expectEmitUserVerifiedEventCall(m, session.SessionID, emitErr)

	err := s.CompleteVerification(t.Context(), domain.TestValidToken)
	assert.ErrorIs(t, err, app.ErrCompleteVerificationOperation)
	assert.ErrorIs(t, err, emitErr)
}

func testOnboardService_CompleteVerification_Success(t *testing.T) {
	t.Parallel()

	s, m := newOnboardService(t, true)
	session := domain.NewValidTestUserVerificationSession(t)
	expectCompleteUserVerificationSessionCall(m, &session, nil)
	expectSaveUserCall(m, nil)
	expectEmitUserVerifiedEventCall(m, session.SessionID, nil)

	err := s.CompleteVerification(t.Context(), domain.TestValidToken)
	assert.NoError(t, err)
}

//
//
//
//
// ===================== DeleteVerificationSession =======================

func TestOnboardService_DeleteVerificationSession(t *testing.T) {
	t.Parallel()

	t.Run("returns error when passed sessionID is not valid", testOnboardService_DeleteVerificationSession_InvalidULID)
	t.Run("returns error when repository fails", testOnboardService_DeleteVerificationSession_RepoFailure)
	t.Run("returns nil when operation succeeds", testOnboardService_DeleteVerificationSession_Success)
}

func testOnboardService_DeleteVerificationSession_InvalidULID(t *testing.T) {
	t.Parallel()
	s, _ := newOnboardService(t, false)

	err := s.DeleteVerificationSession(t.Context(), "invalid-ulid")
	assert.ErrorIs(t, err, app.ErrInvalidULID)
}

func testOnboardService_DeleteVerificationSession_RepoFailure(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)
	id := ulidutil.NewULID()
	repoErr := errors.New("delete restricted")

	expectDeleteVerificationSessionCall(m, id, repoErr)

	err := s.DeleteVerificationSession(t.Context(), id.String())
	assert.ErrorIs(t, err, repoErr)
}

func testOnboardService_DeleteVerificationSession_Success(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)
	id := ulidutil.NewULID()

	expectDeleteVerificationSessionCall(m, id, nil)

	err := s.DeleteVerificationSession(t.Context(), id.String())
	assert.NoError(t, err)
}

//
//
//
//
// ===================== CleanUpVerificationSessions =======================

func TestOnboardService_CleanUpVerificationSessions(t *testing.T) {
	t.Parallel()

	t.Run("returns error when repository fails", func(t *testing.T) {
		t.Parallel()
		service, mock := newOnboardService(t, false)
		expectedErr := errors.New("db cleanup error")

		expectCleanUpVerificationSessionsCall(mock, domain.TestValidUserVerificationCleanUpDuration, expectedErr)

		err := service.CleanUpVerificationSessions(t.Context())
		assert.ErrorIs(t, err, app.ErrCleanUpVerificationSessionsOperation)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns nil when operation succeeds", func(t *testing.T) {
		t.Parallel()
		service, mock := newOnboardService(t, false)

		expectCleanUpVerificationSessionsCall(mock, domain.TestValidUserVerificationCleanUpDuration, nil)

		err := service.CleanUpVerificationSessions(t.Context())
		assert.NoError(t, err)
	})
}

//
//
//
//
// ============ OAuthUser Mapping =============

func TestOAuthUser_ToDomainUser(t *testing.T) {
	t.Parallel()

	ou := validOAuthUser()
	provider := domain.AuthProviderGoogle
	user := ou.ToDomainUser(provider)

	assert.NotNil(t, user)
	assert.Equal(t, ou.Email, user.Email)
	assert.Equal(t, ou.FirstName, user.Creds.FirstName)
	assert.Equal(t, ou.LastName, user.Creds.LastName)
	require.Len(t, user.ExternalIdentities, 1)
	assert.Equal(t, ou.ExternalID, user.ExternalIdentities[0].ID)
	assert.Equal(t, provider, user.ExternalIdentities[0].Provider)
}

//
//
//
//
// ============ CreateOauthUser =============

func TestOnboardService_CreateOauthUser(t *testing.T) {
	t.Parallel()

	t.Run("returns error when validation fails", testOnboardService_CreateOauthUser_ValidationFailure)
	t.Run("successfully creates new user on first attempt", testOnboardService_CreateOauthUser_SuccessFirstTry)
	t.Run("retries and succeeds when username is taken", testOnboardService_CreateOauthUser_UsernameRetrySuccess)
	t.Run("returns error when username is taken 5 times", testOnboardService_CreateOauthUser_UsernameRetryExhausted)
	t.Run("links to existing account when email duplicate occurs", testOnboardService_CreateOauthUser_LinkToExistingEmail)
	t.Run("returns error when linking fails with unexpected error", testOnboardService_CreateOauthUser_LinkingFailure)
	t.Run("returns error when linking fails with dublicata data error", testOnboardService_CreateOauthUser_LinkingFailureAlreasyExists)
	t.Run("returns error when SaveUser fails unexpectedly", testOnboardService_CreateOauthUser_UnexpectedSaveFailure)
}

func testOnboardService_CreateOauthUser_ValidationFailure(t *testing.T) {
	t.Parallel()
	s, _ := newOnboardService(t, false)

	user := validOAuthUser()
	user.Email = "not-an-email"

	id, err := s.CreateOauthUser(t.Context(), domain.AuthProviderGoogle, user)

	ulidtest.Nil(t, id)
	var expectedErr *errorutil.ValidationErr
	assert.ErrorAs(t, err, &expectedErr)
}

func testOnboardService_CreateOauthUser_SuccessFirstTry(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)

	expectSaveUserCall(m, nil)

	id, err := s.CreateOauthUser(t.Context(), domain.AuthProviderGoogle, validOAuthUser())
	assert.NoError(t, err)
	ulidtest.NotNil(t, id)
}

func testOnboardService_CreateOauthUser_UsernameRetrySuccess(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)

	expectSaveUserCall(m, db.ErrDublicateUsername)
	expectSaveUserCall(m, nil)

	id, err := s.CreateOauthUser(t.Context(), domain.AuthProviderGoogle, validOAuthUser())

	assert.NoError(t, err)
	ulidtest.NotNil(t, id)
}

func testOnboardService_CreateOauthUser_UsernameRetryExhausted(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)
	ou := validOAuthUser()

	for i := 1; i < 6; i++ {
		expectSaveUserCall(m, db.ErrDublicateUsername)
	}

	id, err := s.CreateOauthUser(t.Context(), domain.AuthProviderGoogle, ou)

	ulidtest.Nil(t, id)
	assert.ErrorIs(t, err, app.ErrUniqueUsernameGeneration)
}

func testOnboardService_CreateOauthUser_LinkToExistingEmail(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)
	ou := validOAuthUser()
	existingUserID := ulidutil.NewULID()

	expectSaveUserCall(m, db.ErrDublicateEmail)
	expectSaveUserExternalIdentityCall(m, ou.Email, existingUserID, nil)

	id, err := s.CreateOauthUser(t.Context(), domain.AuthProviderGoogle, ou)

	assert.NoError(t, err)
	assert.Equal(t, existingUserID, id)
}

func testOnboardService_CreateOauthUser_LinkingFailure(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)
	ou := validOAuthUser()
	linkErr := errors.New("db connection lost")

	expectSaveUserCall(m, db.ErrDublicateEmail)
	expectSaveUserExternalIdentityCall(m, ou.Email, ulid.ULID{}, linkErr)

	id, err := s.CreateOauthUser(t.Context(), domain.AuthProviderGoogle, ou)

	ulidtest.Nil(t, id)
	assert.ErrorIs(t, err, linkErr)
}

func testOnboardService_CreateOauthUser_LinkingFailureAlreasyExists(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)
	ou := validOAuthUser()

	expectSaveUserCall(m, db.ErrDublicateEmail)
	expectSaveUserExternalIdentityCall(m, ou.Email, ulid.ULID{}, db.ErrDublicateData)

	id, err := s.CreateOauthUser(t.Context(), domain.AuthProviderGoogle, ou)
	ulidtest.Nil(t, id)
	assert.ErrorIs(t, err, app.ErrExternalIdentityAlreadyExists)
}

func testOnboardService_CreateOauthUser_UnexpectedSaveFailure(t *testing.T) {
	t.Parallel()
	s, m := newOnboardService(t, false)
	fatalErr := errors.New("critical storage failure")

	expectSaveUserCall(m, fatalErr)

	id, err := s.CreateOauthUser(t.Context(), domain.AuthProviderGoogle, validOAuthUser())

	ulidtest.Nil(t, id)
	assert.ErrorIs(t, err, app.ErrCreateOauthUserOperation)
}
