package domain_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/platform/domain"
	"github.com/stretchr/testify/assert"
)

func TestInstallState_NotInstalled_WhenDBURLEmpty(t *testing.T) {
	state := domain.NewInstallState(false, false)
	assert.False(t, state.IsInstalled())
	assert.Equal(t, domain.InstallPhaseNeedsConfig, state.Phase())
}

func TestInstallState_NeedsMigration_WhenConfigPresentButNoDB(t *testing.T) {
	state := domain.NewInstallState(true, false)
	assert.False(t, state.IsInstalled())
	assert.Equal(t, domain.InstallPhaseNeedsMigration, state.Phase())
}

func TestInstallState_Installed_WhenBothPresent(t *testing.T) {
	state := domain.NewInstallState(true, true)
	assert.True(t, state.IsInstalled())
	assert.Equal(t, domain.InstallPhaseComplete, state.Phase())
}

func TestInstallState_RedirectPath_NeedsConfig(t *testing.T) {
	state := domain.NewInstallState(false, false)
	assert.Equal(t, "/setup", state.RedirectPath())
}

func TestInstallState_RedirectPath_NeedsMigration(t *testing.T) {
	state := domain.NewInstallState(true, false)
	assert.Equal(t, "/setup/migrate", state.RedirectPath())
}

func TestInstallState_RedirectPath_Complete(t *testing.T) {
	state := domain.NewInstallState(true, true)
	assert.Equal(t, "", state.RedirectPath())
}
