package delivery

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/sky-flux/cms/internal/identity/app"
)

// LoginExecutor is the minimal port the handler needs from the app layer.
type LoginExecutor interface {
	Execute(ctx context.Context, in app.LoginInput) (*app.LoginOutput, error)
}

// Handler holds all identity delivery dependencies.
type Handler struct {
	login LoginExecutor
}

func NewHandler(login LoginExecutor) *Handler {
	return &Handler{login: login}
}

// RegisterRoutes wires all identity endpoints onto the Huma API.
func RegisterRoutes(api huma.API, login LoginExecutor) {
	h := NewHandler(login)
	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/admin/auth/login",
		Summary:     "Login with email and password",
		Tags:        []string{"Auth"},
	}, h.Login)
}

// Login handles POST /api/v1/admin/auth/login.
func (h *Handler) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	out, err := h.login.Execute(ctx, app.LoginInput{
		Email:    req.Body.Email,
		Password: req.Body.Password,
	})
	if err != nil {
		return nil, mapError(err)
	}

	resp := &LoginResponse{}

	if out.Requires2FA {
		// 2FA challenge: return temp token and requires field.
		resp.Body.Requires = "totp"
		resp.Body.TempToken = out.TempToken
		return resp, nil
	}

	resp.Body.UserID = out.UserID
	resp.Body.AccessToken = out.AccessToken
	resp.Body.TokenType = "Bearer"
	resp.Body.ExpiresIn = 900 // 15 min
	return resp, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, app.ErrInvalidCredentials):
		return huma.NewError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, app.ErrAccountDisabled):
		return huma.NewError(http.StatusForbidden, err.Error())
	case errors.Is(err, app.ErrAccountLocked):
		return huma.NewError(http.StatusTooManyRequests, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, "internal error")
	}
}
