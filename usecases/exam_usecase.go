// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"letuan.com/code_demo_backend/domain"
	"time"
)

// examUsecase implements the business logic for managing exams.
type examUsecase struct {
	repo			domain.ExamRepository
	redisClient		*redis.Client
}

// NewExamUsecase initializes the business logic layer for exams.
func NewExamUsecase(repo domain.ExamRepository, rdb *redis.Client) domain.ExamUsecase {
	return &examUsecase{repo, rdb}
}

// validateExamRules enforces the 10 questions minimum rule.
func validateExamRules(req *domain.ExamRequest) error {
	if len(req.QuestionIDs) < 10 {
		return errors.New("422: Exam must contain at least 10 questions to ensure quality")
	}
	return nil
}

// mapExamToResponse packages the exam and formats questions based on role.
func mapExamToResponse(exam domain.Exam, questions []domain.Question, role string) domain.ExamResponse {
	var qResponses []domain.QuestionResponse
	for _, q := range questions {
		var content domain.QuestionContent
		var tags []string
		_ = json.Unmarshal(q.Content, &content)
		_ = json.Unmarshal(q.Tags, &tags)
		qRes := domain.QuestionResponse{
			ID:			q.ID,
			Type:		q.Type,
			Tags:		tags,
			Content:	content,
			CreatedAt:	q.CreatedAt,
		}
		// Strictly sanitize for students.
		if role == "teacher" || role == "admin" {
			var correctData map[string]interface{}
			_ = json.Unmarshal(q.CorrectData, &correctData)
			qRes.CorrectData = correctData
		}
		qResponses = append(qResponses, qRes)
	}
	return domain.ExamResponse{
		ID:			exam.ID,
		Title:		exam.Title,
		Duration:	exam.Duration,
		Questions:	qResponses,
		CreatedAt:	exam.CreatedAt,
	}
}

// GetAll retrieves a paginated list of exams for a tenant, returning lightweight overviews without deeply loading questions.
func (u *examUsecase) GetAll(ctx context.Context, tenantID uint, page, limit int) ([]domain.ExamResponse, error) {
	offset := (page - 1) * limit
	exams, err := u.repo.Fetch(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	var res []domain.ExamResponse
	for _, e := range exams {
		res = append(res, mapExamToResponse(e, nil, "student")) // List does not need to load questions deeply.
	}
	return res, nil
}

// GetByID fetches full exam details, utilizing a Redis caching strategy for students to optimize performance while serving real-time data for teachers.
func (u *examUsecase) GetByID(ctx context.Context, tenantID, id uint, role string) (domain.ExamResponse, error) {
	cacheKey := fmt.Sprintf("exam:%d:%d:student_view", tenantID, id)
	// 1. Redis cache strategy for students.
	if role == "student" {
		val, err := u.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedRes domain.ExamResponse
			if json.Unmarshal([]byte(val), &cachedRes) == nil {
				return cachedRes, nil // Cache hit.
			}
		}
	}
	// 2. Cache miss or teacher request: Fetch from database.
	exam, questions, err := u.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return domain.ExamResponse{}, errors.New("Exam not found")
	}
	res := mapExamToResponse(exam, questions, role)
	// 3. Save to Redis if it is a student view.
	if role == "student" {
		if bytes, err := json.Marshal(res); err == nil {
			// TTL 1 hour.
			u.redisClient.Set(ctx, cacheKey, bytes, time.Hour)
		}
	}
	return res, nil
}

// Create sets up a new exam and links its specified questions.
func (u *examUsecase) Create(ctx context.Context, tenantID uint, req *domain.ExamRequest) (domain.ExamResponse, error) {
	if err := validateExamRules(req); err != nil {
		return domain.ExamResponse{}, err
	}
	exam := domain.Exam{
		TenantID:		tenantID,
		Title:			req.Title,
		Duration:		req.Duration,
		MaxAttempts:	req.MaxAttempts,
		ValidFrom:		req.ValidFrom,
		ValidTo:		req.ValidTo,
	}
	if err := u.repo.Create(ctx, &exam, req.QuestionIDs); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ExamResponse{}, errors.New("409: One or more questions are already assigned to another exam")
		}
		return domain.ExamResponse{}, err
	}
	return u.GetByID(ctx, tenantID, exam.ID, "teacher")
}

// Update modifies an exam's details and question list, importantly invalidating the student cache to prevent stale data.
func (u *examUsecase) Update(ctx context.Context, tenantID, id uint, req *domain.ExamRequest) (domain.ExamResponse, error) {
	if err := validateExamRules(req); err != nil {
		return domain.ExamResponse{}, err
	}
	exam, _, err := u.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return domain.ExamResponse{}, errors.New("Exam not found")
	}
	exam.Title = req.Title
	exam.Duration = req.Duration
	exam.MaxAttempts = req.MaxAttempts
	exam.ValidFrom = req.ValidFrom
	exam.ValidTo = req.ValidTo
	if err := u.repo.Update(ctx, &exam, req.QuestionIDs); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ExamResponse{}, errors.New("409: One or more questions are already assigned to another exam")
		}
		return domain.ExamResponse{}, err
	}
	// Invalidate cache when exam is updated.
	cacheKey := fmt.Sprintf("exam:%d:%d:student_view", tenantID, id)
	u.redisClient.Del(ctx, cacheKey)
	return u.GetByID(ctx, tenantID, exam.ID, "teacher")
}

// Delete safety removes an exam from the specified tenant's records.
func (u *examUsecase) Delete(ctx context.Context, tenantID, id uint) error {
	return u.repo.Delete(ctx, tenantID, id)
}

// GetAnalytics retrieves performance metrics and statistical data for a specific exam.
func (u *examUsecase) GetAnalytics(ctx context.Context, tenantID, id uint) (domain.AnalyticsResponse, error) {
	return u.repo.GetAnalytics(ctx, tenantID, id)
}