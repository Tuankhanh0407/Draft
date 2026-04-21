// Import appropriate package.
package middleware

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
)

// AuthProtected validates the JWT token and extracts user info into the request context.
func AuthProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHandler := c.Get("Authorization")
		if authHandler == "" || !strings.HasPrefix(authHandler, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid token"})
		}
		tokenString := strings.TrimPrefix(authHandler, "Bearer ")
		secretKey := os.Getenv("JWT_SECRET")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token expired or invalid"})
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
		}
		// Store claims in Fiber context for downstream handlers.
		c.Locals("user_id", claims["user_id"])
		c.Locals("tenant_id", claims["tenant_id"])
		c.Locals("role", claims["role"])
		return c.Next()
	}
}

// RoleProtected ensures the user has one of the allowed roles.
func RoleProtected(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Role not found"})
		}
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied. Insufficient permissions"})
	}
}