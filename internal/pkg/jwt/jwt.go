package jwt

import (
	"context"
	"fmt"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Claims struct {
	Subject string
	JTI     string
	Purpose string
}

type registeredClaims struct {
	jwtgo.RegisteredClaims
	Purpose string `json:"purpose,omitempty"`
}

type Manager struct {
	secret    []byte
	accessTTL time.Duration
	tempTTL   time.Duration
	rdb       *redis.Client
}

func NewManager(secret string, accessTTL, tempTTL time.Duration, rdb *redis.Client) *Manager {
	return &Manager{
		secret:    []byte(secret),
		accessTTL: accessTTL,
		tempTTL:   tempTTL,
		rdb:       rdb,
	}
}

func (m *Manager) SignAccessToken(userID string) (string, error) {
	return m.sign(userID, "", m.accessTTL)
}

func (m *Manager) SignTempToken(userID, purpose string) (string, error) {
	return m.sign(userID, purpose, m.tempTTL)
}

func (m *Manager) sign(userID, purpose string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := registeredClaims{
		RegisteredClaims: jwtgo.RegisteredClaims{
			Subject:   userID,
			ID:        uuid.NewString(),
			IssuedAt:  jwtgo.NewNumericDate(now),
			ExpiresAt: jwtgo.NewNumericDate(now.Add(ttl)),
		},
		Purpose: purpose,
	}
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwtgo.ParseWithClaims(tokenStr, &registeredClaims{}, func(t *jwtgo.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtgo.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	rc, ok := token.Claims.(*registeredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return &Claims{
		Subject: rc.Subject,
		JTI:     rc.ID,
		Purpose: rc.Purpose,
	}, nil
}

func (m *Manager) Blacklist(ctx context.Context, jti string, ttl time.Duration) error {
	return m.rdb.Set(ctx, "auth:blacklist:"+jti, "1", ttl).Err()
}

func (m *Manager) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := m.rdb.Exists(ctx, "auth:blacklist:"+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
