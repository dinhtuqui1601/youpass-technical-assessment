package auth

import (
	"errors"
	"testing"
	"time"
)

func TestTokenManagerGenerateAndParse(t *testing.T) {
	tm := NewTokenManager("test-secret", 15*time.Minute, "youpass", "youpass-web")

	token, expiresIn, err := tm.Generate("user-123", "student")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if expiresIn != 900 {
		t.Fatalf("expiresIn = %d, want 900", expiresIn)
	}

	claims, err := tm.Parse(token)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("Subject = %q, want user-123", claims.Subject)
	}
	if claims.Role != "student" {
		t.Errorf("Role = %q, want student", claims.Role)
	}
	if claims.Issuer != "youpass" {
		t.Errorf("Issuer = %q, want youpass", claims.Issuer)
	}
}

func TestTokenManagerExpired(t *testing.T) {
	tm := NewTokenManager("test-secret", -time.Minute, "youpass", "youpass-web")
	token, _, err := tm.Generate("user-123", "student")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if _, err := tm.Parse(token); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("Parse error = %v, want ErrTokenExpired", err)
	}
}

func TestTokenManagerInvalid(t *testing.T) {
	tm := NewTokenManager("test-secret", time.Minute, "youpass", "youpass-web")
	if _, err := tm.Parse("not-a-jwt"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("Parse error = %v, want ErrTokenInvalid", err)
	}
}

func TestTokenManagerWrongSecret(t *testing.T) {
	tm := NewTokenManager("secret-a", time.Minute, "youpass", "youpass-web")
	other := NewTokenManager("secret-b", time.Minute, "youpass", "youpass-web")

	token, _, err := tm.Generate("user", "student")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if _, err := other.Parse(token); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("Parse with wrong secret = %v, want ErrTokenInvalid", err)
	}
}

func TestNewRefreshTokenOpaqueAndHashed(t *testing.T) {
	token, hash, err := newRefreshToken()
	if err != nil {
		t.Fatalf("newRefreshToken error: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
	if hash == token {
		t.Fatal("hash must not equal plaintext token")
	}
	if hashRefreshToken(token) != hash {
		t.Fatal("hashRefreshToken mismatch")
	}

	token2, hash2, _ := newRefreshToken()
	if token2 == token || hash2 == hash {
		t.Fatal("tokens are not unique")
	}
}
