package delivery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-flux/cms/internal/identity/app"
	"github.com/sky-flux/cms/internal/identity/delivery"
)

// stubLoginUseCase satisfies the delivery layer's LoginExecutor interface.
type stubLoginUseCase struct {
	out *app.LoginOutput
	err error
}

func (s *stubLoginUseCase) Execute(ctx context.Context, in app.LoginInput) (*app.LoginOutput, error) {
	return s.out, s.err
}

func newTestAPI(t *testing.T, login delivery.LoginExecutor) huma.API {
	t.Helper()
	_, api := humatest.New(t, huma.DefaultConfig("CMS API", "1.0.0"))
	delivery.RegisterRoutes(api, login)
	return api
}

func TestLoginHandler_Success(t *testing.T) {
	api := newTestAPI(t, &stubLoginUseCase{
		out: &app.LoginOutput{UserID: "u1", AccessToken: "jwt-token"},
	})

	resp := httptest.NewRecorder()
	body := `{"email":"alice@example.com","password":"secret123"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	assert.Equal(t, "jwt-token", out["access_token"])
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	api := newTestAPI(t, &stubLoginUseCase{err: app.ErrInvalidCredentials})

	resp := httptest.NewRecorder()
	body := `{"email":"alice@example.com","password":"wrongpas"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestLoginHandler_AccountLocked(t *testing.T) {
	api := newTestAPI(t, &stubLoginUseCase{err: app.ErrAccountLocked})

	resp := httptest.NewRecorder()
	body := `{"email":"alice@example.com","password":"password1"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
}

func TestLoginHandler_Requires2FA(t *testing.T) {
	api := newTestAPI(t, &stubLoginUseCase{
		out: &app.LoginOutput{UserID: "u1", Requires2FA: true, TempToken: "temp-tok"},
	})

	resp := httptest.NewRecorder()
	body := `{"email":"alice@example.com","password":"correct1"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	api.Adapter().ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	var out map[string]any
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
	assert.Equal(t, "totp", out["requires"])
	assert.Equal(t, "temp-tok", out["temp_token"])
}
