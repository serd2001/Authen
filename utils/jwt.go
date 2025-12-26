package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Custom errors
var (
	ErrMissingJWTSecret = errors.New("JWT_SECRET environment variable is not set")
	ErrWeakJWTSecret    = errors.New("JWT_SECRET must be at least 32 characters for HS256 security")
	ErrInvalidToken     = errors.New("invalid or malformed JWT token")
	ErrExpiredToken     = errors.New("JWT token has expired")
	ErrInvalidClaims    = errors.New("invalid token claims structure")
	ErrTokenNotYetValid = errors.New("JWT token not yet valid")
)

// JWTClaims defines the custom claims structure
type JWTClaims struct {
	UserID      uint   `json:"user_id"`
	Email       string `json:"email"`
	FirebaseUID string `json:"firebase_uid"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration (cached for performance)
type JWTConfig struct {
	Secret      string
	ExpiryHours int
	Issuer      string
}

// TokenWithExpiry holds token and its expiration info
type TokenWithExpiry struct {
	Token     string
	ExpiresIn int64 // seconds until expiration
	ExpiresAt time.Time
}

var (
	jwtConfig     *JWTConfig
	configOnce    sync.Once
	configLoadErr error
)

// loadJWTConfig loads and validates JWT configuration from environment
func loadJWTConfig() (*JWTConfig, error) {
	configOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			configLoadErr = ErrMissingJWTSecret
			return
		}

		// Validate secret strength (critical for HS256)
		if len(secret) < 32 {
			configLoadErr = ErrWeakJWTSecret
			return
		}

		expiryHours, err := strconv.Atoi(os.Getenv("JWT_EXPIRY"))
		if err != nil || expiryHours <= 0 {
			expiryHours = 24 // default to 24 hours
		}

		issuer := os.Getenv("JWT_ISSUER")
		if issuer == "" {
			issuer = "my-app" // default issuer
		}

		jwtConfig = &JWTConfig{
			Secret:      secret,
			ExpiryHours: expiryHours,
			Issuer:      issuer,
		}
	})

	return jwtConfig, configLoadErr
}

// GenerateJWT creates and signs a JWT token string with user info
func GenerateJWT(userID uint, email, firebaseUID, role string) (string, error) {
	config, err := loadJWTConfig()
	if err != nil {
		return "", err
	}

	now := time.Now()
	expiresAt := now.Add(time.Hour * time.Duration(config.ExpiryHours))

	claims := JWTClaims{
		UserID:      userID,
		Email:       email,
		FirebaseUID: firebaseUID,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Issuer,
			Subject:   strconv.FormatUint(uint64(userID), 10),
			Audience:  jwt.ClaimStrings{config.Issuer},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Secret))
}

// ValidateJWT parses and validates a JWT token string
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	config, err := loadJWTConfig()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method to prevent algorithm confusion attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Secret), nil
	})

	if err != nil {
		// Distinguish between different error types for better UX
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, ErrTokenNotYetValid
		default:
			return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
		}
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// GenerateJWTWithExpiry creates a JWT and returns token with expiry info
// This is a convenience function that returns additional expiration metadata
func GenerateJWTWithExpiry(userID uint, email, firebaseUID, role string) (*TokenWithExpiry, error) {
	config, err := loadJWTConfig()
	if err != nil {
		return nil, err
	}

	token, err := GenerateJWT(userID, email, firebaseUID, role)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Hour * time.Duration(config.ExpiryHours))
	expiresIn := int64(config.ExpiryHours * 60 * 60) // seconds

	return &TokenWithExpiry{
		Token:     token,
		ExpiresIn: expiresIn,
		ExpiresAt: expiresAt,
	}, nil
}

// RefreshJWT generates a new token from an existing valid or expired token
func RefreshJWT(tokenString string) (string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		// Allow refresh even if token is expired
		if !errors.Is(err, ErrExpiredToken) {
			return "", err
		}
		// Parse without validation to get claims from expired token
		token, parseErr := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			config, configErr := loadJWTConfig()
			if configErr != nil {
				return nil, configErr
			}
			return []byte(config.Secret), nil
		})
		if parseErr != nil {
			return "", fmt.Errorf("%w: cannot refresh", ErrInvalidToken)
		}
		claims, _ = token.Claims.(*JWTClaims)
	}

	// Generate new token with same user info
	return GenerateJWT(claims.UserID, claims.Email, claims.FirebaseUID, claims.Role)
}

// GetUserIDFromToken extracts user ID from a validated token
func GetUserIDFromToken(tokenString string) (uint, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// IsTokenExpired checks if a token is expired
func IsTokenExpired(tokenString string) bool {
	_, err := ValidateJWT(tokenString)
	return errors.Is(err, ErrExpiredToken)
}

// GetTimeUntilExpiry returns duration until token expires
func GetTimeUntilExpiry(tokenString string) (time.Duration, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return 0, err
	}
	if claims.ExpiresAt == nil {
		return 0, errors.New("token has no expiration time")
	}
	return time.Until(claims.ExpiresAt.Time), nil
}