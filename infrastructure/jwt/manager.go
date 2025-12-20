package jwt

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	authservice "github.com/mzulfanw/boilerplate-go-fiber/internal/service/auth"
)

var ErrInvalidToken = errors.New("jwt: invalid token")

type Manager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

func NewManager(secret, issuer string, accessTTL time.Duration) (*Manager, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("jwt: secret is empty")
	}
	if accessTTL <= 0 {
		return nil, errors.New("jwt: access token ttl must be positive")
	}
	return &Manager{
		secret:    []byte(secret),
		issuer:    strings.TrimSpace(issuer),
		accessTTL: accessTTL,
	}, nil
}

func (m *Manager) GenerateAccessToken(userID string, roles, permissions []string, tokenVersion int) (string, int64, error) {
	if m == nil || len(m.secret) == 0 {
		return "", 0, ErrInvalidToken
	}
	if strings.TrimSpace(userID) == "" {
		return "", 0, ErrInvalidToken
	}

	now := time.Now()
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
		Roles:        roles,
		Permissions:  permissions,
		TokenVersion: tokenVersion,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}

	return signed, int64(m.accessTTL.Seconds()), nil
}

func (m *Manager) ParseAccessToken(tokenString string) (authservice.AccessClaims, error) {
	if m == nil || len(m.secret) == 0 {
		return authservice.AccessClaims{}, ErrInvalidToken
	}
	if strings.TrimSpace(tokenString) == "" {
		return authservice.AccessClaims{}, ErrInvalidToken
	}

	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	}
	if m.issuer != "" {
		options = append(options, jwt.WithIssuer(m.issuer))
	}

	claims := accessClaims{}
	parser := jwt.NewParser(options...)
	token, err := parser.ParseWithClaims(tokenString, &claims, func(_ *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return authservice.AccessClaims{}, ErrInvalidToken
	}

	result := authservice.AccessClaims{
		Subject:      claims.Subject,
		Issuer:       claims.Issuer,
		Roles:        claims.Roles,
		Permissions:  claims.Permissions,
		TokenVersion: claims.TokenVersion,
	}
	if claims.ExpiresAt != nil {
		result.ExpiresAt = claims.ExpiresAt.Time
	}
	return result, nil
}

type accessClaims struct {
	jwt.RegisteredClaims
	Roles        []string `json:"roles,omitempty"`
	Permissions  []string `json:"perms,omitempty"`
	TokenVersion int      `json:"token_version,omitempty"`
}
