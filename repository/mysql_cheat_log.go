// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlCheatLogRepository represents a MySQL-backed repository for managing cheat log data using GORM.
type mysqlCheatLogRepository struct {
	db *gorm.DB
}

// NewMySQLCheatLogRepository creates a new instance of CheatLogRepository.
func NewMySQLCheatLogRepository(db *gorm.DB) domain.CheatLogRepository {
	return &mysqlCheatLogRepository{db: db}
}

// Create records a new proctoring event.
func (m *mysqlCheatLogRepository) Create(log *domain.CheatLog) error {
	return m.db.Create(log).Error
}

// GetByExamAndUser retrieves all proctoring events for a specific student's exam attempt.
func (m *mysqlCheatLogRepository) GetByExamAndUser(examID uint, userID uint, tenantID uint) ([]domain.CheatLog, error) {
	var logs []domain.CheatLog
	// Fetch logs ordered by the exact time they occurred.
	err := m.db.Where("exam_id = ? AND user_id = ? AND tenant_id = ?", examID, userID, tenantID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}