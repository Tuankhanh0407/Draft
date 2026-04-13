// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlUserRepository represents a MySQL-backed repository for managing user data using GORM.
type mysqlUserRepository struct {
	db *gorm.DB
}

// NewMySQLUserRepository creates a new instance of UserRepository.
func NewMySQLUserRepository(db *gorm.DB) domain.UserRepository {
	return &mysqlUserRepository{db: db}
}

// Create inserts a new user account into the database.
func (m *mysqlUserRepository) Create(user *domain.User) error {
	return m.db.Create(user).Error
}

// GetByUsernameAndTenant fetches a user ensuring strict multi-tenant isolation.
func (m *mysqlUserRepository) GetByUsernameAndTenant(username string, tenantID uint) (*domain.User, error) {
	var user domain.User
	// Always include tenant_id in queries to prevent cross-leak data.
	err := m.db.Where("username = ? AND tenant_id = ?", username, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmailAndTenant fetches a user by their email address.
func (m *mysqlUserRepository) GetByEmailAndTenant(email string, tenantID uint) (*domain.User, error) {
	var user domain.User
	err := m.db.Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update modifies an existing user record (used for password reset).
func (m *mysqlUserRepository) Update(user *domain.User) error {
	return m.db.Save(user).Error
}