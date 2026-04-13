// Import appropriate package.
package middleware

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"letuan.com/code_demo_backend/usecases"
	"strings"
)

// Protected is a middleware that validates the JWT token and extracts user context into locals.
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract the Authorization header.
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authorization header"})
		}
		// 2. Validate the "Bearer <token>" format.
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token format"})
		}
		// 3. Parse and validate the token signature.
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return usecases.JWTSecret, nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}
		// 4. Extract claims from the payload.
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Failed to parse token claims"})
		}
		// 5. Store claims in the Fiber context.
		// Note: JWT parses numbers as float64, so explicit casting to uint is required.
		c.Locals("user_id", uint(claims["user_id"].(float64)))
		c.Locals("tenant_id", uint(claims["tenant_id"].(float64)))
		c.Locals("role", claims["role"].(string))
		// 6. Proceed to the next handler.
		return c.Next()
	}
}