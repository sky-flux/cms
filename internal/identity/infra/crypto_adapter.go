package infra

import pkgcrypto "github.com/sky-flux/cms/internal/pkg/crypto"

// CryptoAdapter adapts pkg/crypto to domain.PasswordService.
type CryptoAdapter struct{}

func NewCryptoAdapter() *CryptoAdapter { return &CryptoAdapter{} }

func (c *CryptoAdapter) Hash(plain string) (string, error) {
	return pkgcrypto.HashPassword(plain)
}

func (c *CryptoAdapter) Check(plain, hash string) bool {
	return pkgcrypto.CheckPassword(plain, hash)
}
