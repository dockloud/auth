package userauthn

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var (
	JWTTooShortError    = errors.New("JWT secret is too short")
	InvalidTokenError   = errors.New("invalid token")
	MissingSubjectError = errors.New("missing subject")
)

type Claims struct {
	jwt.RegisteredClaims
}

type Verifier struct {
	secret []byte
}

func NewVerifier(secret string) (*Verifier, error) {
	if len(secret) < 32 {
		return nil, JWTTooShortError
	}

	return &Verifier{
		secret: []byte(secret),
	}, nil
}

func (v *Verifier) Verify(rawToken string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(t *jwt.Token) (any, error) {
			return v.secret, nil
		},
		jwt.WithValidMethods(
			[]string{jwt.SigningMethodHS256.Alg()},
		),
		jwt.WithIssuer("dockloud-auth"),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", InvalidTokenError
	}
	if claims.Subject == "" {
		return "", MissingSubjectError
	}

	return claims.Subject, nil
}
