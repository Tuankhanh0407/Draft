// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"context"
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlUserRepository provides MySQL data access for User entities using GORM.
type mysqlUserRepository struct {
	db *gorm.DB
}

// NewMysqlUserRepository creates a new instance of UserRepository.
func NewMysqlUserRepository(db *gorm.DB) domain.UserRepository {
	return &mysqlUserRepository{db}
}

// Create inserts a new user record into the database.
// SQL: INSERT INTO users (tenant_id, email, password, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?);
func (r *mysqlUserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByEmailAndTenant retrieves a user based on their exact email and associated tenant.
// SQL: SELECT * FROM users WHERE email = ? AND tenant_id = ? LIMIT 1;
func (r *mysqlUserRepository) GetByEmailAndTenant(ctx context.Context, email string, tenantID uint) (domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
	return user, err
}