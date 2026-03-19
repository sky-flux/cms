package domain_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/identity/domain"
)

// Verify the TokenService interface is well-defined by compile check.
var _ domain.TokenService = (*fakeTokenService)(nil)

type fakeTokenService struct{}

func (f *fakeTokenService) IssueAccessToken(userID string) (string, error) { return "tok", nil }
func (f *fakeTokenService) IssueTempToken(userID, purpose string) (string, error) {
	return "tmp", nil
}
func (f *fakeTokenService) Verify(token string) (*domain.TokenClaims, error) { return nil, nil }

var _ domain.PasswordService = (*fakePasswordService)(nil)

type fakePasswordService struct{}

func (f *fakePasswordService) Hash(plain string) (string, error) { return "hashed", nil }
func (f *fakePasswordService) Check(plain, hash string) bool     { return true }

func TestTokenServiceInterface(t *testing.T)   { t.Log("interfaces satisfied") }
func TestPasswordServiceInterface(t *testing.T) { t.Log("interfaces satisfied") }
