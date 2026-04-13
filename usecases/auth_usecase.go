// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"letuan.com/code_demo_backend/domain"
	"math/big"
	"net/smtp"
	"os"
	"time"
)

// JWTSecret is the cryptographic key used to sign JWT tokens.
var JWTSecret = []byte("super-secret-assessment-engine-key")

// authUseCase implements the domain.AuthUseCase interface.
type authUseCase struct {
	userRepo		domain.UserRepository
	redisClient		*redis.Client // Need Redis to store OTPs.
}

// NewAuthUseCase creates a new instance of AuthUseCase with Redis injected.
func NewAuthUseCase(repo domain.UserRepository, rdb *redis.Client) domain.AuthUseCase {
	return &authUseCase{
		userRepo: 		repo,
		redisClient:	rdb,
	}
}

// Register validates input, hashes the password securely, and persists the new user.
func (u *authUseCase) Register(req *domain.AuthRequest) (*domain.User, error) {
	if req.Email == "" {
		return nil, errors.New("Email is required for registration")
	}
	// 1. Check if username uniqueness within the tenant.
	existingUser, _ := u.userRepo.GetByUsernameAndTenant(req.Username, req.TenantID)
	if existingUser != nil {
		return nil, errors.New("Username already exists in this tenant")
	}
	// Check email uniqueness.
	existingEmail, _ := u.userRepo.GetByEmailAndTenant(req.Email, req.TenantID)
	if existingEmail != nil {
		return nil, errors.New("Email already registered in this tenant")
	}
	// 2. Hash the password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Failed to hash password")
	}
	// 3. Construct user entity.
	newUser := &domain.User{
		TenantID:	req.TenantID,
		Username:	req.Username,
		Email:		req.Email,
		Password:	string(hashedPassword),
		Role:		"student", // Default role assignment.
	}
	// 4. Save to the database.
	err = u.userRepo.Create(newUser)
	return newUser, nil
}

// Login verifies user credentials and issues a valid JWT token.
func (u *authUseCase) Login(req *domain.AuthRequest) (string, error) {
	// 1. Retrieve user securely by tenant.
	user, err := u.userRepo.GetByUsernameAndTenant(req.Username, req.TenantID)
	if err != nil {
		return "", errors.New("Invalid username or password")
	}
	// 2. Validate password.
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", errors.New("Invalid username of password")
	}
	// 3. Generate JWT containing integer IDs.
	claims := jwt.MapClaims{
		"user_id":		user.ID,
		"tenant_id":	user.TenantID,
		"role":			user.Role,
		"exp":			time.Now().Add(time.Hour * 72).Unix(), // 72 hours expiration.
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// RequestPasswordReset generates a 6-digit one-time password (OTP), saves it in Redis, and sends an email.
func (u *authUseCase) RequestPasswordReset(req *domain.ForgotPasswordRequest) error {
	// 1. Verify user exists.
	_, err := u.userRepo.GetByEmailAndTenant(req.Email, req.TenantID)
	if err != nil {
		return errors.New("No account found with this email")
	}
	// 2. Generate a 6-digit secure random OTP.
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return errors.New("Failed to generate secure OTP")
	}
	otp := fmt.Sprintf("%06d", n.Int64())
	// 3. Store in Redis with 5 minutes time to live (TTL).
	ctx := context.Background()
	key := fmt.Sprintf("pwd_reset:%s", req.Email)
	err = u.redisClient.Set(ctx, key, otp, 5 * time.Minute).Err()
	if err != nil {
		return errors.New("Failed to save OTP to database")
	}
	// 4. Send the OTP via email asynchronously.
	go u.sendOTPEmail(req.Email, otp)
	return nil
}

// ResetPassword validates the OTP from Redis and securely updates the user's password.
func (u *authUseCase) ResetPassword(req *domain.ResetPasswordRequest) error {
	ctx := context.Background()
	key := fmt.Sprintf("pwd_reset:%s", req.Email)
	// 1. Fetch OTP from Redis.
	storedOTP, err := u.redisClient.Get(ctx, key).Result()
	if err == redis.Nil || storedOTP != req.OTP {
		return errors.New("Invalid or expired OTP")
	}
	// 2. Fetch the user.
	user, err := u.userRepo.GetByEmailAndTenant(req.Email, req.TenantID)
	if err != nil {
		return errors.New("User not found")
	}
	// 3. Hash the new password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Failed to secure the new password")
	}
	// 4. Update database.
	user.Password = string(hashedPassword)
	if err := u.userRepo.Update(user); err != nil {
		return errors.New("Failed to reset password")
	}
	// 5. Clean up the OTP from Redis so it cannot be reused.
	u.redisClient.Del(ctx, key)
	return nil
}

// sendOTPEmail handles the simple mail transfer protocol (SMTP) communication using environment variables.
func (u *authUseCase) sendOTPEmail(toEmail string, otp string) {
	// Load credentials dynamically from the environment.
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	// Validate configuration presence to prevent silent failures.
	if from == "" || password == "" || smtpHost == "" || smtpPort == "" {
		fmt.Println("Error: SMTP configuration is missing in the environment variables.")
		return
	}
	message := []byte(fmt.Sprintf("Subject: Password reset request\r\n\r\n" +
		"Your OTP for resetting your password is: %s\r\n" + 
		"This code is valid for 5 minutes. Do not share it with anyone.", otp))
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost + ":" + smtpPort, auth, from, []string{toEmail}, message)
	if err != nil {
		fmt.Printf("Failed to send OTP to %s: %v\n", toEmail, err)
	}
}