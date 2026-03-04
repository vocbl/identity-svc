package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	app "github.com/vocbl/users-svc/internal/application/onboard"
	"github.com/vocbl/users-svc/internal/application/onboard/mock"
	"github.com/vocbl/users-svc/internal/domain"
	"go.uber.org/mock/gomock"
)

// ============ Controller & Mock Helpers ==============

type testMocks struct {
	ctrl         *gomock.Controller
	verification *mock.MockVerificationRepo
	user         *mock.MockUserRepo
	starter      *mock.MockVerificationStarterRepo
}

func newTestMocks(t *testing.T) *testMocks {
	ctrl := gomock.NewController(t)
	return &testMocks{
		ctrl:         ctrl,
		verification: mock.NewMockVerificationRepo(ctrl),
		user:         mock.NewMockUserRepo(ctrl),
		starter:      mock.NewMockVerificationStarterRepo(ctrl),
	}
}

func newValidVerificationCfg() app.VerificationCfg {
	return app.VerificationCfg{
		MaxAttempts:                   domain.TestValidUserVerificationMaxAttempts,
		ExpirationDuration:            domain.TestValidUserVerificationExpireDuration,
		RestartDuration:               domain.TestValidUserVerificationRestartDuration,
		CleanUpDuration:               domain.TestValidUserVerificationCleanUpDuration,
		UsernameGenerationMaxAttempts: domain.TestValidUsernameGenerationMaxAttempts,
		PasswordHasher: func(s string) (string, error) {
			return string(domain.TestValidPasswordHash), nil
		},
		TokenGenerator: func() (string, string) {
			return string(domain.TestValidToken), string(domain.TestValidTokenHash)
		},

		TokenHasher: func(string) string {
			return string(domain.TestValidTokenHash)
		},
	}
}

func newInvalidVerificationCfg() app.VerificationCfg {
	return app.VerificationCfg{
		MaxAttempts:                   domain.TestInvalidUserVerificationMaxAttempts,
		ExpirationDuration:            domain.TestInvalidUserVerificationDuration,
		RestartDuration:               domain.TestInvalidUserVerificationRestartDuration,
		CleanUpDuration:               domain.TestInvalidUserVerificationCleanUpDuration,
		UsernameGenerationMaxAttempts: domain.TestInvalidUsernameGenerationMaxAttempts,

		PasswordHasher: func(s string) (string, error) {
			return domain.TestInvalidPasswordHash, nil
		},

		TokenGenerator: func() (string, string) {
			return domain.TestInvalidToken, domain.TestInvalidTokenHash
		},

		TokenHasher: func(s string) string {
			return "mismatch_hash_format_that_does_not_match_generator"
		},
	}
}

// ============ Service Factories ==============

func newOnboardService(t *testing.T, expectTransaction bool) (*app.OnboardService, *mock.MockVerificationRepo) {
	ctrl := gomock.NewController(t)
	m := mock.NewMockVerificationRepo(ctrl)
	s, err := app.NewOnboardService(m, newValidVerificationCfg())
	require.NoError(t, err)

	if expectTransaction {
		m.EXPECT().
			WithinTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(app.VerificationRepo) error) error {
				return fn(m)
			}).AnyTimes()
	}

	return s, m
}

func newVerificationStarterService(t *testing.T, expectTransaction bool) (*app.VerificationStarterService, *mock.MockVerificationStarterRepo) {
	ctrl := gomock.NewController(t)
	m := mock.NewMockVerificationStarterRepo(ctrl)
	s, err := app.NewVerificationStarterService(m, newValidVerificationCfg())
	require.NoError(t, err)

	if expectTransaction {
		m.EXPECT().
			WithinTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(app.VerificationStarterRepo) error) error {
				return fn(m)
			}).AnyTimes()
	}

	return s, m
}

func newExternalIdentityService(t *testing.T, expectTransaction bool) (*app.ExternalIdentityService, *mock.MockUserRepo) {
	ctrl := gomock.NewController(t)
	m := mock.NewMockUserRepo(ctrl)
	s, err := app.NewExternalIdentityService(m, newValidVerificationCfg())
	require.NoError(t, err)

	if expectTransaction {
		m.EXPECT().
			WithinTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(app.UserRepo) error) error {
				return fn(m)
			}).AnyTimes()
	}

	return s, m
}

// ============ VerificationRepo (OnboardService) ============

func expectVerificationRepoCheckUsernameExistanceCall(
	m *mock.MockVerificationRepo,
	exists bool,
	err error,
) *gomock.Call {
	return m.EXPECT().
		CheckUsernameExistance(gomock.Any(), gomock.Any()).
		Return(exists, err)
}

func expectVerificationRepoCreateCall(
	m *mock.MockVerificationRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectVerificationRepoDeleteCall(
	m *mock.MockVerificationRepo,
	id domain.VerificationSessionID,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Delete(gomock.Any(), id).
		Return(err)
}

func expectVerificationRepoCleanCall(
	m *mock.MockVerificationRepo,
	d time.Duration,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Clean(gomock.Any(), d).
		Return(err)
}

func expectVerificationRepoGetCall(
	m *mock.MockVerificationRepo,
	sessionID domain.VerificationSessionID,
	session *domain.UserVerificationSession,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Get(gomock.Any(), sessionID).
		Return(session, err)
}

func expectVerificationRepoUpdateCall(
	m *mock.MockVerificationRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectVerificationRepoCreateUserCall(
	m *mock.MockVerificationRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectVerificationRepoEmitVerificationSessionCreatedEventCall(
	m *mock.MockVerificationRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		EmitVerificationSessionCreatedEvent(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Return(err)
}

func expectVerificationRepoEmitVerificationSessionStartedEventCall(
	m *mock.MockVerificationRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		EmitVerificationSessionStartedEvent(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Return(err)
}

func expectVerificationRepoEmitUserVerifiedEventCall(
	m *mock.MockVerificationRepo,
	sessionID domain.VerificationSessionID,
	err error,
) *gomock.Call {
	return m.EXPECT().
		EmitUserVerifiedEvent(gomock.Any(), sessionID).
		Return(err)
}

// ============ UserRepo (ExternalIdentityService) ============

func expectUserRepoGetByEmailCall(
	m *mock.MockUserRepo,
	email domain.Email,
	user *domain.User,
	err error,
) *gomock.Call {
	return m.EXPECT().
		GetByEmail(gomock.Any(), email).
		Return(user, err)
}

func expectUserRepoCreateCall(
	m *mock.MockUserRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectUserRepoUpdateCall(
	m *mock.MockUserRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(err)
}

// ============ VerificationStarterRepo (VerificationStarterService) ============

func expectVerificationStarterRepoGetCall(
	m *mock.MockVerificationStarterRepo,
	sessionID domain.VerificationSessionID,
	session *domain.UserVerificationSession,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Get(gomock.Any(), sessionID).
		Return(session, err)
}

func expectVerificationStarterRepoUpdateCall(
	m *mock.MockVerificationStarterRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(err)
}

func expectVerificationStarterRepoEmitVerificationSessionStartedEventCall(
	m *mock.MockVerificationStarterRepo,
	err error,
) *gomock.Call {
	return m.EXPECT().
		EmitVerificationSessionStartedEvent(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).
		Return(err)
}

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
		Email:      string(domain.TestValidEmail),
		FirstName:  "John",
		LastName:   "Doe",
		ExternalID: "external-id",
	}
}
