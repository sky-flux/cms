package domain

// TokenClaims is the parsed, trusted result of a verified JWT.
type TokenClaims struct {
	UserID  string
	JTI     string
	Purpose string
}

// TokenService is the port for JWT operations.
// The infra adapter wraps pkg/jwt.Manager.
type TokenService interface {
	IssueAccessToken(userID string) (string, error)
	IssueTempToken(userID, purpose string) (string, error)
	Verify(token string) (*TokenClaims, error)
}

// PasswordService is the port for bcrypt operations.
// The infra adapter wraps pkg/crypto functions.
type PasswordService interface {
	Hash(plain string) (string, error)
	Check(plain, hash string) bool
}
