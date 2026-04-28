package main

import (
	"fmt"
	"log"

	"backend/internal/auth/services"
)

func main() {
	// Use the same secret as the server (from GetEnvJWTSecret)
	secret := "taskify-dev-secret-change-in-production"

	// Create JWT service with 24 hour expiration
	jwtService := services.NewJWTService(secret, 24)

	// Generate token for user-1
	userID := "user-1"
	email := "test@example.com"

	token, err := jwtService.GenerateToken(userID, email)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	fmt.Println("Generated JWT Token:")
	fmt.Println(token)
	fmt.Println()
	fmt.Println("Use with curl:")
	fmt.Printf("curl -H \"Authorization: Bearer %s\" http://localhost:8080/api/projects/1/members\n", token)
}
