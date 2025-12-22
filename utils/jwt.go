package utils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID      uint   `json:"user_id"`
	Email       string `json:"email"`
	FirebaseUID string `json:"firebase_uid"`
	Name        string `json:"name"`
	Lastname    string `json:"lastname"`
	Gender      string `json:"gender"`
	Age         int    `json:"age"`
	Dob         string `json:"dob"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(
	userID uint,
	email string,
	firebaseUID string,
	name string,
	lastname string,
	gender string,
	age int,
	dob string,
	role string,
) (string, error) {

	expiryHours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY"))
	if expiryHours == 0 {
		expiryHours = 24
	}

	claims := JWTClaims{
		UserID:      userID,
		Email:       email,
		FirebaseUID: firebaseUID,
		Name:        name,
		Lastname:    lastname,
		Gender:      gender,
		Age:         age,
		Dob:         dob,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expiryHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}


func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
func GenerateJWTWithExpiry(userID uint, email string, firebaseUID string) (string, int64, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // 24 hours
	expiresIn := int64(24 * 60 * 60)                 // 86400 seconds

	claims := jwt.MapClaims{
		"user_id":      userID,
		"email":        email,
		"firebase_uid": firebaseUID,
		"exp":          expirationTime.Unix(),
		"iat":          time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, expiresIn, err
}