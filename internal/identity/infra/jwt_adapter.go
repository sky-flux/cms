package infra

import (
	"github.com/sky-flux/cms/internal/identity/domain"
	pkgjwt "github.com/sky-flux/cms/internal/pkg/jwt"
)

// JWTAdapter adapts pkg/jwt.Manager to domain.TokenService.
type JWTAdapter struct {
	mgr *pkgjwt.Manager
}

func NewJWTAdapter(mgr *pkgjwt.Manager) *JWTAdapter {
	return &JWTAdapter{mgr: mgr}
}

func (a *JWTAdapter) IssueAccessToken(userID string) (string, error) {
	return a.mgr.SignAccessToken(userID)
}

func (a *JWTAdapter) IssueTempToken(userID, purpose string) (string, error) {
	return a.mgr.SignTempToken(userID, purpose)
}

func (a *JWTAdapter) Verify(token string) (*domain.TokenClaims, error) {
	c, err := a.mgr.Verify(token)
	if err != nil {
		return nil, err
	}
	return &domain.TokenClaims{UserID: c.Subject, JTI: c.JTI, Purpose: c.Purpose}, nil
}
