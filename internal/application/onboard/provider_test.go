package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	app "github.com/vocbl/users-svc/internal/application/onboard"
)

func TestNewOnboardService(t *testing.T) {
	t.Parallel()
	m := newTestMocks(t)

	t.Run("Happy Path: Successful Initialization", func(t *testing.T) {
		t.Parallel()

		svc, err := app.NewOnboardService(m.verification, newValidVerificationCfg())
		assert.NoError(t, err)
		assert.NotNil(t, svc)
	})

	t.Run("Total Fail: Domain Policy Violations", func(t *testing.T) {
		svc, err := app.NewOnboardService(m.verification, newInvalidVerificationCfg())
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}

func TestNewVerificationStarterService(t *testing.T) {
	t.Parallel()
	m := newTestMocks(t)

	t.Run("Happy Path: Successful Initialization", func(t *testing.T) {
		t.Parallel()

		svc, err := app.NewVerificationStarterService(m.starter, newValidVerificationCfg())
		assert.NoError(t, err)
		assert.NotNil(t, svc)
	})

	t.Run("Total Fail: Domain Policy Violations", func(t *testing.T) {
		t.Parallel()

		svc, err := app.NewVerificationStarterService(m.starter, newInvalidVerificationCfg())
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}

func TestNewExternalIdentityService(t *testing.T) {
	t.Parallel()
	m := newTestMocks(t)

	t.Run("Happy Path: Successful Initialization", func(t *testing.T) {
		t.Parallel()

		svc, err := app.NewExternalIdentityService(m.user, newValidVerificationCfg())
		assert.NoError(t, err)
		assert.NotNil(t, svc)
	})

	t.Run("Total Fail: Domain Policy Violations", func(t *testing.T) {
		t.Parallel()

		svc, err := app.NewExternalIdentityService(m.user, newInvalidVerificationCfg())
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}
