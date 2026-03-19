// Package delivery contains Huma handlers for install wizard and audit log.
package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sky-flux/cms/internal/platform/app"
)

// InstallExecutor is the port the install handler needs from the app layer.
type InstallExecutor interface {
	TestDBConnection(ctx context.Context, dsn string) error
	RunMigrations(ctx context.Context) error
	CreateSuperAdmin(ctx context.Context, in app.CreateAdminInput) error
	WriteEnvFile(path string, vals map[string]string) error
}

// InstallHandler serves the web installation wizard endpoints.
type InstallHandler struct {
	installer InstallExecutor
}

// NewInstallHandler creates an InstallHandler.
func NewInstallHandler(installer InstallExecutor) *InstallHandler {
	return &InstallHandler{installer: installer}
}

// RegisterInstallRoutes wires setup wizard endpoints onto a Chi router.
func RegisterInstallRoutes(r chi.Router, h *InstallHandler) {
	r.Get("/setup", h.SetupPage)
	r.Post("/setup/test-db", h.TestDB)
	r.Post("/setup/migrate", h.Migrate)
	r.Post("/setup/create-admin", h.CreateAdmin)
}

// testDBRequest is the request body for POST /setup/test-db.
type testDBRequest struct {
	DatabaseURL string `json:"database_url"`
}

// createAdminRequest is the request body for POST /setup/create-admin.
type createAdminRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	JWTSecret string `json:"jwt_secret"`
}

// SetupPage handles GET /setup — returns current installation status as JSON.
func (h *InstallHandler) SetupPage(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "not_installed",
		"message": "Complete the setup wizard to install Sky Flux CMS.",
	})
}

// TestDB handles POST /setup/test-db — tests a database connection.
func (h *InstallHandler) TestDB(w http.ResponseWriter, r *http.Request) {
	var body testDBRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DatabaseURL == "" {
		writeJSON(w, http.StatusBadRequest, errResp("database_url is required"))
		return
	}
	if err := h.installer.TestDBConnection(r.Context(), body.DatabaseURL); err != nil {
		if errors.Is(err, app.ErrDBConnectionFailed) {
			writeJSON(w, http.StatusUnprocessableEntity, errResp(err.Error()))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errResp("internal error"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Migrate handles POST /setup/migrate — runs all pending migrations.
func (h *InstallHandler) Migrate(w http.ResponseWriter, r *http.Request) {
	if err := h.installer.RunMigrations(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateAdmin handles POST /setup/create-admin — creates super-admin and writes .env.
func (h *InstallHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var body createAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
		return
	}
	if body.Email == "" || body.Password == "" {
		writeJSON(w, http.StatusBadRequest, errResp("email and password are required"))
		return
	}
	if err := h.installer.CreateSuperAdmin(r.Context(), app.CreateAdminInput{
		Email:    body.Email,
		Password: body.Password,
		Name:     body.Name,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp(err.Error()))
		return
	}
	envVals := map[string]string{}
	if body.JWTSecret != "" {
		envVals["JWT_SECRET"] = body.JWTSecret
	}
	// WriteEnvFile is best-effort; errors are logged but do not fail the response.
	_ = h.installer.WriteEnvFile("./.env", envVals)

	writeJSON(w, http.StatusCreated, map[string]string{
		"status": "installed",
		"action": "restart_required",
	})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func errResp(msg string) map[string]string {
	return map[string]string{"error": msg}
}
