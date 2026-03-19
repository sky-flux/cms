package public

import (
	"context"
	"encoding/json"
	"net/http"
)

// APIKeyValidator is the port used by OptionalAPIKey middleware.
// Implemented by the site BC's infra layer (lookup by SHA-256 hash in sfc_api_keys).
type APIKeyValidator interface {
	// ValidateAPIKey checks whether rawKey is a valid, active API key.
	// Returns (false, nil) for an invalid key; (false, err) for internal errors.
	ValidateAPIKey(ctx context.Context, rawKey string) (bool, error)
}

// OptionalAPIKey returns a Chi middleware that:
//   - Allows requests without an X-API-Key header (public access).
//   - Rejects requests that supply an X-API-Key header but fail validation (401).
//   - Passes requests with a valid key, setting "api_key_valid" in context.
func OptionalAPIKey(v APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				// No key supplied — allow anonymous access.
				next.ServeHTTP(w, r)
				return
			}

			ok, err := v.ValidateAPIKey(r.Context(), key)
			if err != nil {
				// Internal error — fail-open (log and continue).
				next.ServeHTTP(w, r)
				return
			}
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"title":  "Unauthorized",
					"detail": "invalid or inactive API key",
				})
				return
			}
			ctx := context.WithValue(r.Context(), ctxKeyAPIKeyValid{}, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type ctxKeyAPIKeyValid struct{}

// IsAPIKeyAuthenticated reports whether the current request has a validated API key.
func IsAPIKeyAuthenticated(ctx context.Context) bool {
	v, _ := ctx.Value(ctxKeyAPIKeyValid{}).(bool)
	return v
}

// writeJSON writes a JSON response with the given status code and body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
