package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(jwtSecret)

type claims struct {
	jwt.RegisteredClaims
}

func generateToken() (string, error) {
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rukkie",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(jwtKey)
}

func validateToken(tokenStr string) error {
	if tokenStr == "" {
		return errors.New("not logged in — run `rukkie login`")
	}
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	if err != nil {
		return errors.New("session expired or invalid — run `rukkie login`")
	}
	if !token.Valid {
		return errors.New("invalid token — run `rukkie login`")
	}
	return nil
}

// Login validates the password, generates a JWT, and stores it.
func Login(password string) error {
	if password != hardcodedPassword {
		return errors.New("incorrect password")
	}
	token, err := generateToken()
	if err != nil {
		return err
	}
	return saveToken(token)
}

// Logout clears the stored token.
func Logout() error {
	return clearToken()
}

// RequireAuth returns an error if the user is not logged in.
func RequireAuth() error {
	stored, err := loadStored()
	if err != nil {
		return errors.New("not logged in — run `rukkie login`")
	}
	return validateToken(stored.Token)
}
