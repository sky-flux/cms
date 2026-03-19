// Package domain contains pure value objects and entities for the Platform BC.
package domain

// InstallPhase represents the current installation state of the CMS.
type InstallPhase int

const (
	// InstallPhaseNeedsConfig means DATABASE_URL is not set; user must complete setup step 1.
	InstallPhaseNeedsConfig InstallPhase = iota
	// InstallPhaseNeedsMigration means DATABASE_URL is set but migrations have not run.
	InstallPhaseNeedsMigration
	// InstallPhaseComplete means the CMS is fully installed and operational.
	InstallPhaseComplete
)

// InstallState is an immutable value object describing whether the CMS is installed.
// It is the source of truth for InstallGuard middleware routing decisions.
type InstallState struct {
	hasConfig bool // DATABASE_URL is present and non-empty
	hasDB     bool // sfc_migrations table exists in the database
}

// NewInstallState constructs an InstallState from two boolean checks.
//   - hasConfig: true when DATABASE_URL (or equivalent) is configured
//   - hasDB:     true when the migrations metadata table exists in PostgreSQL
func NewInstallState(hasConfig, hasDB bool) InstallState {
	return InstallState{hasConfig: hasConfig, hasDB: hasDB}
}

// IsInstalled reports whether both config and DB are present.
func (s InstallState) IsInstalled() bool {
	return s.hasConfig && s.hasDB
}

// Phase returns the granular installation phase.
func (s InstallState) Phase() InstallPhase {
	switch {
	case !s.hasConfig:
		return InstallPhaseNeedsConfig
	case !s.hasDB:
		return InstallPhaseNeedsMigration
	default:
		return InstallPhaseComplete
	}
}

// RedirectPath returns the setup URL the InstallGuard should redirect to,
// or an empty string when the CMS is fully installed.
func (s InstallState) RedirectPath() string {
	switch s.Phase() {
	case InstallPhaseNeedsConfig:
		return "/setup"
	case InstallPhaseNeedsMigration:
		return "/setup/migrate"
	default:
		return ""
	}
}
