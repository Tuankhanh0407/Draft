// Import appropriate package.
package http

// Import necessary libraries.
import (
	"github.com/gofiber/fiber/v2"
	"letuan.com/code_demo_backend/domain"
)

// UserHandler manages incoming API requests and HTTP responses for users.
type UserHandler struct {
	Usecase domain.UserUsecase
}

// NewUserHandler registers routes for user authentication.
func NewUserHandler(app *fiber.App, us domain.UserUsecase) {
	handler := &UserHandler{Usecase: us}
	api := app.Group("/api/v1/auth")

	api.Post("/register", handler.Register)
	api.Post("/login", handler.Login)
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