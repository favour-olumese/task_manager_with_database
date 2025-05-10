package infrastructure

import (
	"errors"
	"fmt"
	domain "task_manager/Domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("ahnljdbjiohwebljnsknpihdbuo")

type jwtService struct{}

func NewJWTService() domain.JWTService {
	return &jwtService{}
}

// Creates a new JWT for a given username and role
func (service *jwtService) GenerateToken(username, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	claims := domain.CustomClaims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username, // Unique identifier for the subject
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // Using HS256 signing method.

	// Sign the token with the secret key.
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", errors.New("failed to sign token")
	}

	return tokenString, nil
}

// Parses and validates a JWT
func (service *jwtService) ValidateToken(tokenString string) (*domain.CustomClaims, error) {
	claims := &domain.CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil // Provide the secret key for validation
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}
