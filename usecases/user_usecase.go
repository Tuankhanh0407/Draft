// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"letuan.com/code_demo_backend/domain"
	"os"
	"regexp"
	"strings"
	"time"
)

// userUsecase implements the business logic for user operations.
type userUsecase struct {
	repo domain.UserRepository
}

// NewUserUsecase initializes the business logic layer for users.
func NewUserUsecase(repo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{repo}
}

// validateEmail checks if the provided string matches a standard email format.
func validateEmail(email string) bool {
	regex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return regex.MatchString(email)
}

// Register handles business logic for creating a new user account.
func (u *userUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (domain.User, error) {
	// 1. Data sanitization.
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	// 2. Data validation.
	if !validateEmail(req.Email) {
		return domain.User{}, errors.New("Invalid email format")
	}
	if len(req.Password) < 8 {
		return domain.User{}, errors.New("Password must be at least 8 characters long")
	}
	// 3. Hash password.
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, errors.New("Failed to hash password")
	}
	role := "student"
	if req.Role != "" {
		role = req.Role
	}
	user := domain.User{
		TenantID:	req.TenantID,
		Email:		req.Email,
		Password:	string(hashedBytes),
		Role:		role,
	}
	// 4. Save to database.
	if err := u.repo.Create(ctx, &user); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return domain.User{}, errors.New("User with this email already exists in the specified tenant")
		}
		return domain.User{}, err
	}
	return user, nil
}

// Login handles business logic for authenticating a user and issuing a JWT.
func (u *userUsecase) Login(ctx context.Context, req *domain.LoginRequest) (domain.LoginResponse, error) {
	// 1. Data sanitization.
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	// 2. Fetch user.
	user, err := u.repo.GetByEmailAndTenant(ctx, req.Email, req.TenantID)
	if err != nil {
		// Generic error message to prevent enumeration attacks.
		return domain.LoginResponse{}, errors.New("Invalid credentials")
	}
	// 3. Compare password.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Generic error message.
		return domain.LoginResponse{}, errors.New("Invalid credentials")
	}
	// 4. Generate JWT.
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return domain.LoginResponse{}, errors.New("Internal server error: Missing token secret")
	}
	claims := jwt.MapClaims{
		"user_id":		user.ID,
		"tenant_id":	user.TenantID,
		"role":			user.Role,
		"exp":			time.Now().Add(time.Hour * 24).Unix(), // Expire in 24 hours.
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return domain.LoginResponse{}, errors.New("Failed to generate token")
	}
	return domain.LoginResponse{
		Token:	t,
		User:	user,
	}, nil
}