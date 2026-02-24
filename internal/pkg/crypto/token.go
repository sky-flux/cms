package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func GenerateToken(byteLen int) (string, string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	raw := hex.EncodeToString(b)
	return raw, HashToken(raw), nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func GenerateBackupCodes(n int) []string {
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	codes := make([]string, 0, n)
	seen := make(map[string]bool)
	for len(codes) < n {
		code := randomCode(charset, 4) + "-" + randomCode(charset, 4)
		if !seen[code] {
			seen[code] = true
			codes = append(codes, code)
		}
	}
	return codes
}

func randomCode(charset string, length int) string {
	b := make([]byte, length)
	rand.Read(b)
	var sb strings.Builder
	for _, v := range b {
		sb.WriteByte(charset[int(v)%len(charset)])
	}
	return sb.String()
}
