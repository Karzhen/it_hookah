package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

type RefreshTokenManager struct {
	ttl    time.Duration
	pepper string
}

func NewRefreshTokenManager(ttl time.Duration, pepper string) *RefreshTokenManager {
	return &RefreshTokenManager{
		ttl:    ttl,
		pepper: pepper,
	}
}

func (m *RefreshTokenManager) Generate() (plainToken string, tokenHash string, expiresAt time.Time, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", time.Time{}, err
	}

	plainToken = base64.RawURLEncoding.EncodeToString(raw)
	tokenHash = m.Hash(plainToken)
	expiresAt = time.Now().Add(m.ttl)
	return plainToken, tokenHash, expiresAt, nil
}

func (m *RefreshTokenManager) Hash(token string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", token, m.pepper)))
	return hex.EncodeToString(sum[:])
}
