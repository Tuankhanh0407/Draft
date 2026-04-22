// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"context"
	"gorm.io/gorm"
	"time"
)

// User represents a user entity in the database.
// The combination of TenantID and Email must be unique.
type User struct {
	ID			uint			`json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID	uint			`json:"tenant_id" gorm:"uniqueIndex:idx_tenant_email;not null"`
	Email		string			`json:"email" gorm:"uniqueIndex:idx_tenant_email;size:255;not null"`
	Password	string			`json:"-" gorm:"not null"` // "-" ensures password is never exported in JSON.
	Role		string			`json:"role" gorm:"size:50;default:'student'"`
	CreatedAt	time.Time		`json:"created_at"`
	UpdatedAt	time.Time		`json:"updated_at"`
	DeletedAt	gorm.DeletedAt	`json:"-" gorm:"index"`
}

// RegisterRequest represents the payload for user registration.
type RegisterRequest struct {
	TenantID	uint			`json:"tenant_id" validate:"required"`
	Email		string			`json:"email" validate:"required,email"`
	Password	string			`json:"password" validate:"required,min=8"`
	Role		string			`json:"role"`
}

// LoginRequest represents the payload for user authentication.
type LoginRequest struct {
	TenantID	uint			`json:"tenant_id" validate:"required"`
	Email		string			`json:"email" validate:"required,email"`
	Password	string			`json:"password" validate:"required"`
}

// LoginResponse represents the payload returned upon successful login.
type LoginResponse struct {
	Token		string			`json:"token"`
	User		User			`json:"user"`
}

// UserRepository defines the contract for user database operations.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmailAndTenant(ctx context.Context, email string, tenantID uint) (User, error)
	DeleteAccount(ctx context.Context, tenantID, userID uint) error
}

// UserUsecase defines the business logic for user authentication.
type UserUsecase interface {
	Register(ctx context.Context, req *RegisterRequest) (User, error)
	Login(ctx context.Context, req *LoginRequest) (LoginResponse, error)
	DeleteAccount(ctx context.Context, tenantID, userID uint) error
}