// Package middleware contains Chi middleware for the Platform BC.
package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sky-flux/cms/internal/platform/domain"
)

// InstallChecker provides the current install state for the InstallGuard middleware.
// The infra layer implements this by calling domain.NewInstallState with live checks.
type InstallChecker interface {
	// DetectInstallState performs the two-step detection:
	//  1. Is DATABASE_URL configured?
	//  2. Does sfc_migrations exist in the database?
	DetectInstallState() domain.InstallState
}

// isSetupPath returns true for paths that must always be reachable regardless of
// install state: the setup wizard endpoints and the console SPA.
func isSetupPath(path string) bool {
	return path == "/setup" ||
		strings.HasPrefix(path, "/setup/") ||
		strings.HasPrefix(path, "/console")
}

// InstallGuard is a Chi middleware that intercepts all requests when the CMS
// is not fully installed, redirecting browsers to the setup wizard and returning
// JSON errors to API clients.
//
// Passthrough rules (always allowed regardless of install state):
//   - /setup and anything under /setup/
//   - /console/* (so the setup wizard SPA loads)
//   - /health
func InstallGuard(checker InstallChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Always allow setup paths, the console SPA, and the health endpoint.
			if isSetupPath(path) || path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			state := checker.DetectInstallState()
			if state.IsInstalled() {
				next.ServeHTTP(w, r)
				return
			}

			redirectTo := state.RedirectPath()

			// API clients that explicitly accept JSON get a structured error response.
			// Browser requests (no Accept: application/json) get a redirect.
			if strings.Contains(r.Header.Get("Accept"), "application/json") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
					"title":     "Service Unavailable",
					"detail":    "CMS is not installed. Complete the setup wizard.",
					"setup_url": redirectTo,
				})
				return
			}

			http.Redirect(w, r, redirectTo, http.StatusFound)
		})
	}
}
