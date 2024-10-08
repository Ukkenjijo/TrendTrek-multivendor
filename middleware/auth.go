package middleware

import (
	"log"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware() fiber.Handler {
	jwtSecretKey := os.Getenv("JWT_SECRET_KEY") // Load secret key from environment variable
	log.Println("mid", jwtSecretKey)

	// Return the middleware handler
	return jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(jwtSecretKey)}, // Secret key for signing JWTs
		TokenLookup: "header:Authorization",                        // Look for the token in the Authorization header
		AuthScheme:  "Bearer",                                      // Use Bearer token scheme
		ErrorHandler: func(c *fiber.Ctx, err error) error {

			log.Println(err, []byte(jwtSecretKey))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			// Extract claims directly from Locals (set by jwtware middleware)
			userToken := c.Locals("user").(*jwt.Token)
			claims, ok := userToken.Claims.(jwt.MapClaims)
			if !ok {
				log.Println("Failed to extract claims from context")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Failed to parse token claims",
				})
			}

			// Log the extracted claims for debugging
			log.Printf("Extracted claims: %v\n", claims)

			// Store user ID and role in the context for future use
			c.Locals("user_id", claims["user_id"])
			c.Locals("role", claims["role"])

			return c.Next() // Proceed to the next handler
		},
	})
}

// SellerRoleMiddleware ensures that only users with the 'seller' role can access the route
func SellerRoleMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the user's role from the context (set in JWT middleware)
		role := c.Locals("role")

		// Check if the role is 'seller'
		if role != "seller" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Only sellers can access this route.",
			})
		}

		return c.Next() // Proceed to the next handler if the role is seller
	}
}

// SellerRoleMiddleware ensures that only users with the 'seller' role can access the route
func AdminRoleMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the user's role from the context (set in JWT middleware)
		role := c.Locals("role")
		log.Println(role)

		// Check if the role is 'seller'
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Only admins can access this route.",
			})
		}

		return c.Next() // Proceed to the next handler if the role is seller
	}
}
