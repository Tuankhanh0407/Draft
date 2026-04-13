// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"time"
)

// User represents a system user (student or admin) tied to a specific tenant.
type User struct {
	ID				uint			`json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID		uint			`json:"tenant_id" gorm:"not null;index"` // Multi-tenant compliance.
	Username		string			`json:"username" gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_username"`
	Email			string			`json:"email" gorm:"type:varchar(150);not null;uniqueIndex:idx_tenant_email"` // Required for password recovery.
	Password		string			`json:"-" gorm:"type:varchar(255);not null"` // Hyphen "-" hides password from JSON responses.
	Role			string			`json:"role" gorm:"type:varchar(20);default:'student'"`
	CreatedAt		time.Time		`json:"created_at"`
	UpdatedAt		time.Time		`json:"updated_at"`
	DeletedAt		gorm.DeletedAt	`json:"-" gorm:"index"` // Soft delete.
}

// AuthRequest represents the payload for login and registration requests.
type AuthRequest struct {
	TenantID 		uint 			`json:"tenant_id" validate:"required"`
	Username 		string 			`json:"username" validate:"required,min=4"`
	Email			string			`json:"email,omitempty" validate:"omitempty,email"` // Optional for login, required for registration.
	Password 		string 			`json:"password" validate:"required,min=6"`
}

// ForgetPasswordRequest is the payload to request an OTP.
type ForgotPasswordRequest struct {
	TenantID		uint			`json:"tenant_id" validate:"required"`
	Email			string			`json:"email" validate:"required,email"`
}

// ResetPasswordRequest is the payload to request an OTP.
type ResetPasswordRequest struct {
	TenantID		uint			`json:"tenant_id" validate:"required"`
	Email			string			`json:"email" validate:"required,email"`
	OTP				string			`json:"otp" validate:"required,len=6"`
	NewPassword		string			`json:"new_password" validate:"required,min=6"`
}

// UserRepository defines database operations for a User.
type UserRepository interface {
	// Create saves a new user record to the database.
	Create(user *User) error
	// GetByUsernameAndTenant finds a user based on their username within a specific tenant.
	GetByUsernameAndTenant(username string, tenantID uint) (*User, error)
	// GetByEmailAndTenant gets user by email to send OTP.
	GetByEmailAndTenant(email string, tenantID uint) (*User, error)
	// Update updates user details (in example, changing password).
	Update(user *User) error
}

// AuthUseCase defines business logic for authentication.
type AuthUseCase interface {
	// Register validates inputs and creates a new user account.
	Register(req *AuthRequest) (*User, error)
	// Login authenticates a user and returns a JWT token upon success.
	Login(req *AuthRequest) (string, error)
	// RequestPasswordReset requests an OTP to the user's email.
	RequestPasswordReset(req *ForgotPasswordRequest) error
	// ResetPassword validates OTP and applies the new password.
	ResetPassword(req *ResetPasswordRequest) error
}