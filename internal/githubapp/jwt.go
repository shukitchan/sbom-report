package githubapp

import (
	"crypto/rsa"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type nowFunc func() time.Time

func NewAppJWT(appID string, privateKeyPEM string, now nowFunc) (string, error) {
	key, err := parseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", err
	}

	n := now().UTC()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(n.Add(-30 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(n.Add(9 * time.Minute)),
		Issuer:    appID,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := t.SignedString(key)
	if err != nil {
		return "", err
	}
	return ss, nil
}

func parseRSAPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	pemStr = strings.TrimSpace(pemStr)
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid PEM: could not decode private key")
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pemStr))
	if err != nil {
		return nil, fmt.Errorf("parse rsa private key: %w", err)
	}
	return key, nil
}