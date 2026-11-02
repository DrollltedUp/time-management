package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret []byte, ttl time.Duration) *TokenManager {
	if len(secret) == 0 {
		panic("auth: secret must not be empty")
	}
	if ttl <= 0 {
		panic("auth: ttl must be positive")
	}
	return &TokenManager{
		secret: secret,
		ttl:    ttl,
	}
}

type Claims struct {
	Role Role `json:"role"`
	jwt.RegisteredClaims
}

func (m *TokenManager) Issue(userID int, role Role) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("auth: invalid user id %d", userID)
	}
	if !role.Valid() {
		return "", fmt.Errorf("auth: invalid role %q", role)
	}

	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return &Claims{}, ErrExpiredToken
		}
		return &Claims{}, ErrInvalidToken
	}
	_ = token

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil || userID <= 0 {
		return &Claims{}, ErrInvalidToken
	}
	if !claims.Role.Valid() {
		return &Claims{}, ErrInvalidToken
	}

	return claims, nil
}
