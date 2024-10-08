package utils

import (
	"log"
	"os"
	"time"

	"github.com/Ukkenjijo/trendtrek/models"
	"github.com/golang-jwt/jwt/v5"
)

// Custom claims struct for JWT
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

   

// Load JWT secret key from environment variable
var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))


// generateJWT generates a JWT token for the authenticated user
func GenerateJWT(user models.User) (string, error) {
	// Set expiration time for the token (e.g., 24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the JWT claims, including the user ID and email
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // Use the new method for expiration
		},

	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	log.Println("tokengenerated",token)
	log.Println(jwtSecretKey)
	

	// Sign the token with the secret key
	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}
