// Import appropriate package.
package repository

// Import necessary libraries.
import (
	"context"
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
	"strings"
)

// mysqlExamRepository provides MySQL data access for Exam entities using GORM.
type mysqlExamRepository struct {
	db *gorm.DB
}

// NewMysqlExamRepository creates a new instance of ExamRepository.
func NewMysqlExamRepository(db *gorm.DB) domain.ExamRepository {
	return &mysqlExamRepository{db}
}

// Fetch retrieves paginated exams.
// SQL: SELECT * FROM exams WHERE tenant_id = ? AND deleted_at IS NULL LIMIT ? OFFSET ?;
func (r *mysqlExamRepository) Fetch(ctx context.Context, tenantID uint, limit, offset int) ([]domain.Exam, error) {
	var exams []domain.Exam
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Limit(limit).Offset(offset).Find(&exams).Error
	return exams, err
}

// GetByID fetches an exam and its associated questions.
func (r *mysqlExamRepository) GetByID(ctx context.Context, tenantID, id uint) (domain.Exam, []domain.Question, error) {
	// SQL 1: SELECT * FROM exams WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL LIMIT 1;
	var exam domain.Exam
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&exam).Error; err != nil {
		return exam, nil, err
	}
	// SQL 2: SELECT q.* FROM questions q JOIN exam_questions eq ON q.id = eq.question_id WHERE eq.exam_id = ?;
	var questions []domain.Question
	r.db.WithContext(ctx).Joins("JOIN exam_questions eq ON questions.id = eq.question_id").Where("eq.exam_id = ?", id).Find(&questions)
	return exam, questions, nil
}

// Create uses a transaction to insert an exam and link its questions.
func (r *mysqlExamRepository) Create(ctx context.Context, exam *domain.Exam, questionIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// SQL 1: INSERT INTO exams (...) VALUES (...);
		if err := tx.Create(exam).Error; err != nil {
			return err
		}
		var mappings []domain.ExamQuestion
		for _, qID := range questionIDs {
			mappings = append(mappings, domain.ExamQuestion{ExamID: exam.ID, QuestionID: qID})
		}
		// SQL 2: INSERT INTO exam_questions (exam_id, question_id) VALUES (...), (...);
		if err := tx.Create(&mappings).Error; err != nil {
			// Catch MySQL unique constraint error for reused questions.
			if strings.Contains(err.Error(), "Duplicate entry") {
				return gorm.ErrDuplicatedKey
			}
			return err
		}
		return nil
	})
}

// Update modifies exam metadata and replaces question mappings within a transaction.
func (r *mysqlExamRepository) Update(ctx context.Context, exam *domain.Exam, questionIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// SQL 1: UPDATE exams SET title = ?, duration = ?, max_attempts = ?, valid_from = ?, valid_to = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL;
		if err := tx.Model(exam).Omit("created_at").Updates(exam).Error; err != nil {
			return err
		}
		// SQL 2: DELETE FROM exam_questions WHERE exam_id = ?;
		if err := tx.Where("exam_id = ?", exam.ID).Delete(&domain.ExamQuestion{}).Error; err != nil {
			return err
		}
		var mappings []domain.ExamQuestion
		for _, qID := range questionIDs {
			mappings = append(mappings, domain.ExamQuestion{ExamID: exam.ID, QuestionID: qID})
		}
		// SQL 3: INSERT INTO exam_questions (exam_id, question_id) VALUES (?, ?), (?, ?),...;
		if err := tx.Create(&mappings).Error; err != nil {
			return err
		}
		return nil
	})
}

// Delete performs a soft delete on the exam.
// SQL: UPDATE exams SET deleted_at = NOW() WHERE id = ? AND tenant_id = ?;
func (r *mysqlExamRepository) Delete(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&domain.Exam{}, id).Error
}

// GetAnalytics aggregates submission data for a specific exam.
// SQL: SELECT COUNT(id), AVG(score) FROM submissions WHERE exam_id = ? AND tenant_id = ?;
func (r *mysqlExamRepository) GetAnalytics(ctx context.Context, tenantID, id uint) (domain.AnalyticsResponse, error) {
	var stats domain.AnalyticsResponse
	stats.ExamID = id
	var count int64
	var avgScore float64
	// Assuming a Submissions table exists.
	r.db.WithContext(ctx).Model(&domain.Submission{}).Where("exam_id = ? AND tenant_id = ?", id, tenantID).Count(&count)
	// Avoid division by zero error in database by checking count first.
	if count > 0 {
		r.db.WithContext(ctx).Model(&domain.Submission{}).Where("exam_id = ? AND tenant_id = ?", id, tenantID).Select("AVG(score)").Row().Scan(&avgScore)
	}
	stats.TotalSubmissions = count
	stats.AverageScore = avgScore
	return stats, nil
}