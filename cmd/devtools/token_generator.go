package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Use the same secret as in docker-compose.yml
	secret := "your-development-secret"
	
	// Create the claims
	claims := jwt.MapClaims{
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"nbf":   time.Now().Unix(),
		"iss":   "dev-auth",
		"roles": []string{"admin"},
		"uid":   "test-user",
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Fatalf("Error signing token: %v", err)
	}

	fmt.Printf("Token: %s\n", tokenString)
} 