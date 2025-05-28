package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Command line flags
	role := flag.String("role", "user", "Role to include in token (admin/user/viewer)")
	userId := flag.String("user", "test-user", "User ID to include in token")
	secret := flag.String("secret", "your-development-secret", "JWT secret key")
	issuer := flag.String("issuer", "dev-auth", "Token issuer")
	duration := flag.Duration("duration", 1*time.Hour, "Token duration")
	flag.Parse()

	// Create claims
	now := time.Now()
	claims := jwt.MapClaims{
		"uid":   *userId,
		"roles": []string{*role},
		"iss":   *issuer,
		"exp":   now.Add(*duration).Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(*secret))
	if err != nil {
		log.Fatalf("Error signing token: %v", err)
	}

	// Create output
	output := struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
		UserID    string    `json:"user_id"`
		Roles     []string  `json:"roles"`
	}{
		Token:     tokenString,
		ExpiresAt: time.Unix(claims["exp"].(int64), 0),
		UserID:    claims["uid"].(string),
		Roles:     []string{*role},
	}

	// Print as JSON
	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(jsonOutput))
} 