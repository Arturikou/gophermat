package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const tokenExp = time.Minute * 15

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

type Manager struct {
	secretKey []byte
}

func NewManager(secretKey string) *Manager {
	return &Manager{secretKey: []byte(secretKey)}
}

func (m *Manager) BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

func (m *Manager) GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secretKey, nil
		})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}
