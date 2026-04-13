// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlQuestionRepository represents a MySQL-backed repository for managing question data using GORM.
type mysqlQuestionRepository struct {
	db *gorm.DB
}

// NewMySQLQuestionRepository creates a new instance of QuestionRepository.
func NewMySQLQuestionRepository(db *gorm.DB) domain.QuestionRepository {
	return &mysqlQuestionRepository{db: db}
}

// Create inserts a new question into the database.
func (m *mysqlQuestionRepository) Create(question *domain.Question) error {
	return m.db.Create(question).Error
}

// GetByIDAndTenant fetches a question ensuring it belongs to the specified tenant.
func (m *mysqlQuestionRepository) GetByIDAndTenant(id uint, tenantID uint) (*domain.Question, error) {
	var question domain.Question
	// Ensure tenant_id is strictly checked.
	err := m.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// CreateInBatches inserts an array of questions efficiently in a single database transaction.
func (m *mysqlQuestionRepository) CreateInBatches(questions []domain.Question, batchSize int) error {
	// GORM's automatically chunks the data and builds a single BULK INSERT SQL statement.
	return m.db.CreateInBatches(questions, batchSize).Error
}

// List retrieves paginated questions, optionally filtering by a specific JSON tag.
func (m *mysqlQuestionRepository) List(tenantID uint, limit int, offset int, tag string) ([]domain.Question, int64, error) {
	var questions []domain.Question
	var total int64
	// Start building the query.
	query := m.db.Model(&domain.Question{}).Where("tenant_id = ?", tenantID)
	// Apply tag filter if a tag is provided.
	// MySQL's JSON_CONTAINS requires the search value to be a valid JSON string (wrapped in double quotes).
	if tag != "" {
		jsonTag := `"` + tag + `"`
		query = query.Where("JSON_CONTAINS(tags, ?)", jsonTag)
	}
	// Count total records matching the filter.
	query.Count(&total)
	// Fetch paginated data sorted by the newest first.
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&questions).Error
	return questions, total, err
}