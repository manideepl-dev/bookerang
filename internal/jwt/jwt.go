package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func Generate(secret, username string) (string, error) {
	claims := jwtlib.RegisteredClaims{
		Subject:   username,
		IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Validate(rawToken, secret string) (string, error) {
	claims := &jwtlib.RegisteredClaims{}
	token, err := jwtlib.ParseWithClaims(rawToken, claims, func(token *jwtlib.Token) (interface{}, error) {
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}

	if claims.Subject == "" {
		return "", errors.New("missing subject")
	}
	return claims.Subject, nil
}
