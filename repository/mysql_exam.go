// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"fmt"
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlExamRepository represents a MySQL-backed repository for managing exam data using GORM.
type mysqlExamRepository struct {
	db *gorm.DB
}

// NewMySQLExamRepository creates a new instance of ExamRepository.
func NewMySQLExamRepository(db *gorm.DB) domain.ExamRepository {
	return &mysqlExamRepository{db: db}
}

// Create inserts a new exam record into the database.
func (m *mysqlExamRepository) Create(exam *domain.Exam) error {
	return m.db.Create(exam).Error
}

// GetByIDAndTenant fetches an exam securely by checking the tenant ownership.
func (m *mysqlExamRepository) GetByIDAndTenant(id uint, tenantID uint) (*domain.Exam, error) {
	var exam domain.Exam
	err := m.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&exam).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

// CheckQuestionsInUse validates whether any of the provided questions IDs are already assigned to an existing exam.
func (m *mysqlExamRepository) CheckQuestionsInUse(tenantID uint, questionIDs []uint) (bool, error) {
	for _, qID := range questionIDs {
		var count int64
		// MySQL JSON_CONTAINS expects the search value to be stringified.
		queryVal := fmt.Sprintf("%d", qID)
		err := m.db.Model(&domain.Exam{}).
			Where("tenant_id = ? AND JSON_CONTAINS(question_ids, ?)", tenantID, queryVal).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		// If count is greater than 0, it means this question is already inside another exam's JSON array.
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}