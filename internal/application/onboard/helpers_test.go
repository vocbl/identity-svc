package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid"
	"github.com/stretchr/testify/require"
	app "github.com/vocbl/users-svc/internal/application/onboard"

	"github.com/vocbl/users-svc/internal/application/onboard/mock"
	"github.com/vocbl/users-svc/internal/domain"
	"github.com/vocbl/users-svc/internal/shared/security"
	"go.uber.org/mock/gomock"
)

// ============ Service Set Up ==============
func newRepoMock(t *testing.T) *mock.MockOnboardRepo {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	return mock.NewMockOnboardRepo(ctrl)
}

func newOnboardService(t *testing.T, expectTransaction bool) (*app.OnboardService, *mock.MockOnboardRepo) {
	mockRepo := newRepoMock(t)

	if expectTransaction {
		mockRepo.EXPECT().
			WithinTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(app.OnboardRepo) error) error {
				return fn(mockRepo)
			})
	}

	service, err := app.NewOnboardService(
		mockRepo,
		domain.TestValidVerificationPolicy(),
		security.HashPassword,
		security.NewTokenGenerator(domain.TokenLenght),
	)

	require.NoError(t, err)
	return service, mockRepo
}

//
//
//
//============ Mock Repo Calls=============

func expectCheckUsernameExistanceCall(mock *mock.MockOnboardRepo, username string, repoUsernameExists bool, repoErr error) {
	mock.EXPECT().
		CheckUsernameExistance(gomock.Any(), username).
		Return(repoUsernameExists, repoErr)
}

func expectSaveUserVerificationSessionCall(mock *mock.MockOnboardRepo, err error) {
	mock.EXPECT().
		SaveUserVerificationSession(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectEmitVerifyUserEventCall(mock *mock.MockOnboardRepo, email string, err error) {
	mock.EXPECT().
		EmitVerifyUserEvent(gomock.Any(), email, gomock.Any()).
		Return(err)
}

func expectEmitUserCreatedEventCall(mock *mock.MockOnboardRepo, email, password string, err error) {
	mock.EXPECT().
		EmitUserCreatedEvent(gomock.Any(), gomock.Any(), email, password, gomock.Any()).
		Return(err)
}

func expectRestartUserVerificationSessionCall(mock *mock.MockOnboardRepo, sessionID ulid.ULID, email string, attempts int, err error) {
	mock.EXPECT().
		RestartUserVerificationSession(gomock.Any(), sessionID, gomock.Any(), gomock.Any()).
		Return(email, attempts, err)
}

func expectCompleteUserVerificationSessionCall(mock *mock.MockOnboardRepo, session *domain.UserVerificationSession, err error) {
	mock.EXPECT().
		CompleteUserVerificationSession(gomock.Any(), gomock.Any()).
		Return(session, err)
}

func expectDeleteVerificationSessionCall(mock *mock.MockOnboardRepo, sessionID ulid.ULID, err error) {
	mock.EXPECT().
		DeleteUserVerificationSession(gomock.Any(), sessionID).
		Return(err)
}

func expectSaveUserCall(mock *mock.MockOnboardRepo, err error) {
	mock.EXPECT().
		SaveUser(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectEmitUserVerifiedEventCall(mock *mock.MockOnboardRepo, sessionID ulid.ULID, err error) {
	mock.EXPECT().
		EmitUserVerifiedEvent(gomock.Any(), sessionID).
		Return(err)
}

func expectCleanUpVerificationSessionsCall(mock *mock.MockOnboardRepo, duration time.Duration, err error) {
	mock.EXPECT().
		CleanUpVerificationSessions(gomock.Any(), duration).
		Return(err)
}

func expectWithinTransactionCall(mock *mock.MockOnboardRepo, err error) {
	mock.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectSetUserVerificationSessionPasswordHashCall(mockRepo *mock.MockOnboardRepo, sessionID ulid.ULID, updateErr error,
) {
	mockRepo.EXPECT().
		SetUserVerificationSessionPasswordHash(gomock.Any(), sessionID, gomock.Any()).
		Return(updateErr)
}

func expectSaveUserExternalIdentityCall(mockRepo *mock.MockOnboardRepo, email string, userID ulid.ULID, err error) {
	mockRepo.EXPECT().
		SaveUserExternalIdentity(gomock.Any(), email, gomock.Any()).
		Return(userID, err)
}

// ================== Test Data =========================

func validNewUser() app.NewUser {
	return app.NewUser{
		Email:     domain.TestValidEmail,
		Password:  domain.TestValidPassword,
		FirstName: domain.TestValidFirstName,
		LastName:  domain.TestValidLastName,
		Username:  domain.TestValidUsername,
	}

}

func validOAuthUser() app.OAuthUser {
	return app.OAuthUser{
		Email:      "nazar@example.com",
		FirstName:  "Nazar",
		LastName:   "Volynets",
		ExternalID: "google-sub-12345",
	}
}
