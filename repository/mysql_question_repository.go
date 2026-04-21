// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"context"
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
)

// mysqlQuestionRepository provides MySQL data access for Question entities using GORM.
type mysqlQuestionRepository struct {
	db *gorm.DB
}

// NewMysqlQuestionRepository creates a new instance.
func NewMysqlQuestionRepository(db *gorm.DB) domain.QuestionRepository {
	return &mysqlQuestionRepository{db}
}

// Fetch retrieves a list of questions with pagination and filters.
// SQL: SELECT * FROM questions WHERE tenant_id = ? AND deleted_at IS NULL [AND type = ?] [AND JSON_CONTAINS(tags, ?)] LIMIT ? OFFSET ?;
func (r *mysqlQuestionRepository) Fetch(ctx context.Context, tenantID uint, qType, tag string, limit, offset int) ([]domain.Question, error) {
	var questions []domain.Question
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if qType != "" {
		query = query.Where("type = ?", qType)
	}
	if tag != "" {
		// JSON_CONTAINS check for MySQL JSON arrays.
		query = query.Where("JSON_CONTAINS(tags, ?)", `"` + tag + `"`)
	}
	err := query.Limit(limit).Offset(offset).Find(&questions).Error
	return questions, err
}

// GetByID retrieves a single question belonging to the specific tenant.
// SQL: SELECT * FROM questions WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL LIMIT 1;
func (r *mysqlQuestionRepository) GetByID(ctx context.Context, tenantID, id uint) (domain.Question, error) {
	var question domain.Question
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&question).Error
	return question, err
}

// Create inserts a single question.
// SQL: INSERT INTO questions (tenant_id, type, tags, content, correct_data, created_at, updated_at) VALUES (...);
func (r *mysqlQuestionRepository) Create(ctx context.Context, question *domain.Question) error {
	return r.db.WithContext(ctx).Create(question).Error
}

// CreateBulk inserts multiple questions inside a transaction.
// SQL: BEGIN; INSERT INTO questions (...) VALUES (...), (...); COMMIT;
func (r *mysqlQuestionRepository) CreateBulk(ctx context.Context, questions []domain.Question) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&questions).Error; err != nil {
			return err // Rollback on any error.
		}
		return nil // Commit.
	})
}

// Update saves modifications to an existing question.
// SQL: UPDATE questions SET type = ?, tags = ?, content = ?, correct_data = ?, updated_at = ? WHERE id = ? AND tenant_id = ?;
func (r *mysqlQuestionRepository) Update(ctx context.Context, question *domain.Question) error {
	return r.db.WithContext(ctx).Save(question).Error
}

// Delete performs a soft delete.
// SQL: UPDATE questions SET deleted_at = NOW() WHERE id = ? AND tenant_id = ?;
func (r *mysqlQuestionRepository) Delete(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&domain.Question{}, id).Error
}