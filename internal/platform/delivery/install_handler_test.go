package delivery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sky-flux/cms/internal/platform/app"
	"github.com/sky-flux/cms/internal/platform/delivery"
	"github.com/sky-flux/cms/internal/platform/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- stubs ---

type stubInstaller struct {
	testDBErr      error
	migrateErr     error
	createAdminErr error
	writeEnvErr    error
}

func (s *stubInstaller) TestDBConnection(_ context.Context, _ string) error { return s.testDBErr }
func (s *stubInstaller) RunMigrations(_ context.Context) error               { return s.migrateErr }
func (s *stubInstaller) CreateSuperAdmin(_ context.Context, _ app.CreateAdminInput) error {
	return s.createAdminErr
}
func (s *stubInstaller) WriteEnvFile(_ string, _ map[string]string) error { return s.writeEnvErr }

func newInstallRouter(installer delivery.InstallExecutor) *chi.Mux {
	h := delivery.NewInstallHandler(installer)
	r := chi.NewRouter()
	delivery.RegisterInstallRoutes(r, h)
	return r
}

func TestSetupPage_GET_Returns200(t *testing.T) {
	r := newInstallRouter(&stubInstaller{})
	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}

func TestTestDB_Success_Returns200(t *testing.T) {
	r := newInstallRouter(&stubInstaller{})
	body, _ := json.Marshal(map[string]string{"database_url": "postgres://localhost/cms"})
	req := httptest.NewRequest(http.MethodPost, "/setup/test-db", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestTestDB_Failure_Returns422(t *testing.T) {
	r := newInstallRouter(&stubInstaller{testDBErr: app.ErrDBConnectionFailed})
	body, _ := json.Marshal(map[string]string{"database_url": "postgres://bad"})
	req := httptest.NewRequest(http.MethodPost, "/setup/test-db", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestMigrate_Success_Returns200(t *testing.T) {
	r := newInstallRouter(&stubInstaller{})
	req := httptest.NewRequest(http.MethodPost, "/setup/migrate", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateAdmin_Success_Returns201WithRestartRequired(t *testing.T) {
	r := newInstallRouter(&stubInstaller{})
	body, _ := json.Marshal(map[string]string{
		"email":      "admin@example.com",
		"password":   "secret123",
		"name":       "Admin",
		"jwt_secret": "changeme",
	})
	req := httptest.NewRequest(http.MethodPost, "/setup/create-admin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "installed", resp["status"])
	assert.Equal(t, "restart_required", resp["action"])
}

func TestCreateAdmin_MissingEmail_Returns400(t *testing.T) {
	r := newInstallRouter(&stubInstaller{})
	body, _ := json.Marshal(map[string]string{"password": "secret123"})
	req := httptest.NewRequest(http.MethodPost, "/setup/create-admin", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Verify stubInstaller satisfies the interface.
var _ delivery.InstallExecutor = (*stubInstaller)(nil)

// Keep compiler happy with domain import.
var _ = domain.InstallPhaseComplete
var _ = errors.New
