// Import appropriate package.
package middleware

// Import necessary libraries.
import (
	"github.com/casbin/casbin/v2"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

// RoleBasedAuth uses Casbin Enforcer to check if the current user role has permission to access the route.
// It MUST be placed after the Protected() JWT middleware.
func RoleBasedAuth(enforcer *casbin.Enforcer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Get role from JWT context.
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Role not found in token"})
		}
		// Capitalize role to match Casbin policies (in example, "student" -> "Student").
		// We use cases.Title to replace the deprecated strings.Title method.
		// formattedRole := strings.Title(strings.ToLower(role))
		caser := cases.Title(language.English)
		formattedRole := caser.String(strings.ToLower(role))
		path := c.Path()
		method := c.Method()
		// 2. Check permission via Casbin engine.
		allowed, err := enforcer.Enforce(formattedRole, path, method)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify access permissions"})
		}
		// 3. Block access if not allowed.
		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied. Insufficient permissions"})
		}
		// 4. Continue to the actual handler.
		return c.Next()
	}
}