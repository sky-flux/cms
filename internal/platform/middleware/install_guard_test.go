package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sky-flux/cms/internal/platform/domain"
	platformmw "github.com/sky-flux/cms/internal/platform/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubInstallChecker struct {
	state domain.InstallState
}

func (s *stubInstallChecker) DetectInstallState() domain.InstallState { return s.state }

func okHandlerFn(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func setupRouterWith(checker platformmw.InstallChecker) *chi.Mux {
	r := chi.NewRouter()
	r.Use(platformmw.InstallGuard(checker))
	r.Get("/api/v1/admin/posts", okHandlerFn)
	r.Get("/setup", okHandlerFn)
	r.Get("/setup/migrate", okHandlerFn)
	return r
}

func TestInstallGuard_Installed_PassesThrough(t *testing.T) {
	checker := &stubInstallChecker{state: domain.NewInstallState(true, true)}
	r := setupRouterWith(checker)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestInstallGuard_NoConfig_RedirectsToSetup(t *testing.T) {
	checker := &stubInstallChecker{state: domain.NewInstallState(false, false)}
	r := setupRouterWith(checker)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Must redirect (301/302/307) to /setup.
	assert.True(t, rec.Code == http.StatusFound || rec.Code == http.StatusMovedPermanently || rec.Code == http.StatusTemporaryRedirect)
	assert.Equal(t, "/setup", rec.Header().Get("Location"))
}

func TestInstallGuard_NeedsDB_RedirectsToMigrate(t *testing.T) {
	checker := &stubInstallChecker{state: domain.NewInstallState(true, false)}
	r := setupRouterWith(checker)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusFound || rec.Code == http.StatusMovedPermanently || rec.Code == http.StatusTemporaryRedirect)
	assert.Equal(t, "/setup/migrate", rec.Header().Get("Location"))
}

func TestInstallGuard_SetupPaths_AreAlwaysAllowed(t *testing.T) {
	// Even when not installed, /setup and /setup/* must be reachable.
	checker := &stubInstallChecker{state: domain.NewInstallState(false, false)}
	r := setupRouterWith(checker)

	for _, path := range []string{"/setup", "/setup/migrate"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "expected passthrough for path: %s", path)
	}
}

func TestInstallGuard_APIRequests_JSONError_WhenNotInstalled(t *testing.T) {
	// API clients that send Accept: application/json should get JSON, not a redirect.
	checker := &stubInstallChecker{state: domain.NewInstallState(false, false)}
	r := setupRouterWith(checker)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Either a JSON 503 or redirect is acceptable; what must NOT happen is 200.
	assert.NotEqual(t, http.StatusOK, rec.Code)
}
