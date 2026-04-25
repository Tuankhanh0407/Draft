// Import appropriate package.
package http

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
	"letuan.com/code_demo_backend/delivery/http/middleware"
	"letuan.com/code_demo_backend/domain"
)

// UserHandler manages incoming API requests and HTTP responses for users.
type UserHandler struct {
	Usecase domain.UserUsecase
}

// NewUserHandler registers routes for user authentication and account management.
func NewUserHandler(app *fiber.App, us domain.UserUsecase) {
	handler := &UserHandler{Usecase: us}
	
	// Public routes.
	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/register", handler.Register)
	authGroup.Post("/login", handler.Login)

	// Protected route for account deletion (JWT required).
	protectedGroup := app.Group("/api/v1/account", middleware.AuthProtected())
	protectedGroup.Delete("/", handler.DeleteAccount)
}

// Registers registers a new user in the system.
// @Summary Register a new user
// @Description Create a new user account bound to a specific tenant
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Registration info"
// @Success 201 {object} domain.User "Successfully registered user"
// @Failure 400 {object} map[string]string "Bad request: Invalid request body or validation failed"
// @Failure 409 {onject} map[string]string "Conflict: User already exists in the specified tenant"
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	user, err := h.Usecase.Register(c.Context(), &req)
	if err != nil {
		// Differentiate between validation errors and database conflicts.
		if err.Error() == "User with this email already exists in the specified tenant" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

// Login authenticates a user and provides a token.
// @Summary User login
// @Description Authenticate a user and return a JWT
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login info"
// @Success 200 {object} domain.LoginResponse "Successfully authenticated"
// @Failure 400 {object} map[string]string "Bad request: Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized: Invalid credentials"
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	res, err := h.Usecase.Login(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

// DeleteAccount removes the currently authenticated user's account.
// @Summary Delete current user account.
// @Description Allow a logged-in user to (soft) delete their own account. It requires a valid JWT token
// @Tags Account
// @Security BearerAuth
// @Produce json
// @Success 204 "No content: Account successfully deleted"
// @Failure 401 {object} map[string]string "Unauthorized: Missing or invalid token"
// @Failure 500 {object} map[string]string "Internal server error: Failed to delete account"
// @Router /api/v1/account [delete]
func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	tenantID := uint(c.Locals("tenant_id").(float64))
	userID := uint(c.Locals("user_id").(float64))
	if err := h.Usecase.DeleteAccount(c.Context(), tenantID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}