package jwt

import (
	"errors"
	"fmt"
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
	token, err := jwtlib.Parse(rawToken, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwtlib.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	subject, ok := claims["sub"].(string)
	if !ok || subject == "" {
		return "", errors.New("missing subject")
	}
	return subject, nil
}
