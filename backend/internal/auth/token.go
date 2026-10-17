package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrTokenExpired indicates the access token is well-formed but expired.
	ErrTokenExpired = errors.New("access token expired")
	// ErrTokenInvalid indicates the access token failed verification.
	ErrTokenInvalid = errors.New("access token invalid")
)

// Claims is the JWT payload of an access token.
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies access tokens.
type TokenManager struct {
	secret   []byte
	ttl      time.Duration
	issuer   string
	audience string
}

// NewTokenManager builds a TokenManager.
func NewTokenManager(secret string, ttl time.Duration, issuer, audience string) *TokenManager {
	return &TokenManager{
		secret:   []byte(secret),
		ttl:      ttl,
		issuer:   issuer,
		audience: audience,
	}
}

// Generate issues a new HS256 access token for the given user.
func (t *TokenManager) Generate(userID, role string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(t.ttl)

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    t.issuer,
			Audience:  jwt.ClaimStrings{t.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.NewString(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	if err != nil {
		return "", 0, err
	}
	return token, int64(t.ttl.Seconds()), nil
}

// Parse verifies an access token and returns its claims. It distinguishes
// expired tokens from invalid ones so callers can map to the right error code.
func (t *TokenManager) Parse(raw string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (any, error) {
		return t.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// newRefreshToken generates an opaque random refresh token and its SHA-256 hash.
// Only the hash is persisted; the plaintext is returned once to the caller.
func newRefreshToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashRefreshToken(token), nil
}

// hashRefreshToken returns the SHA-256 hex digest of a refresh token.
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
