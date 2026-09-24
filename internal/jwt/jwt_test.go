package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidate(t *testing.T) {
	rawToken, err := Generate("secret", "alice")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	username, err := Validate(rawToken, "secret")
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if username != "alice" {
		t.Fatalf("username = %q, want alice", username)
	}
}

func TestValidateRejectsWrongSecretAndAlgorithm(t *testing.T) {
	rawToken, err := Generate("secret", "alice")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if _, err := Validate(rawToken, "wrong-secret"); err == nil {
		t.Fatal("expected wrong secret to be rejected")
	}

	claims := jwtlib.RegisteredClaims{
		Subject:   "alice",
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}
	wrongAlgorithmToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS384, claims)
	rawToken, err = wrongAlgorithmToken.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign wrong algorithm token: %v", err)
	}
	if _, err := Validate(rawToken, "secret"); err == nil {
		t.Fatal("expected wrong algorithm to be rejected")
	}
}
