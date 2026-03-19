package public_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sky-flux/cms/internal/delivery/public"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAPIKeyValidator struct {
	valid bool
}

func (s *stubAPIKeyValidator) ValidateAPIKey(ctx context.Context, rawKey string) (bool, error) {
	return s.valid, nil
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) //nolint:errcheck
}

func TestOptionalAPIKey_NoKey_Passthrough(t *testing.T) {
	// Without a key header the request must still reach the handler (optional auth).
	mw := public.OptionalAPIKey(&stubAPIKeyValidator{valid: false})
	r := chi.NewRouter()
	r.With(mw).Get("/test", okHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestOptionalAPIKey_InvalidKey_Returns401(t *testing.T) {
	// An explicit but invalid key must be rejected.
	mw := public.OptionalAPIKey(&stubAPIKeyValidator{valid: false})
	r := chi.NewRouter()
	r.With(mw).Get("/test", okHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "bad-key")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestOptionalAPIKey_ValidKey_Passthrough(t *testing.T) {
	mw := public.OptionalAPIKey(&stubAPIKeyValidator{valid: true})
	r := chi.NewRouter()
	r.With(mw).Get("/test", okHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "valid-key-123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
