// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"letuan.com/code_demo_backend/domain"
	"math"
	"time"
)

// questionUseCase implements the domain.QuestionUseCase interface.
type questionUseCase struct {
	questionRepo	domain.QuestionRepository
	redisClient 	*redis.Client
}

// NewQuestionUseCase injects the repository and Redis client.
func NewQuestionUseCase(repo domain.QuestionRepository, rdb *redis.Client) domain.QuestionUseCase {
	return &questionUseCase{
		questionRepo:	repo,
		redisClient:	rdb,
	}
}

// CreateQuestion validates and inserts a single question.
func (u *questionUseCase) CreateQuestion(question *domain.Question) error {
	return u.questionRepo.Create(question)
}

// GetQuestionForClient fetches a question via Redis cache and strips sensitive correct answers/explanations.
func (u *questionUseCase) GetQuestionForClient(id uint, tenantID uint) (*domain.Question, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("question:%d:%d", tenantID, id)
	var question domain.Question
	// 1. Try fetching from cache.
	cachedData, err := u.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		json.Unmarshal([]byte(cachedData), &question)
	} else {
		// 2. Fetch from database on cache miss.
		q, err := u.questionRepo.GetByIDAndTenant(id, tenantID)
		if err != nil {
			return nil, errors.New("Question not found")
		}
		question = *q
		// 3. Update cache.
		questionBytes, _ := json.Marshal(question)
		u.redisClient.Set(ctx, cacheKey, questionBytes, time.Hour)
	}
	// SECURITY: Ensure the student cannot cheat by inspecting the API response.
	question.CorrectData = nil
	question.Explanation = "" // Hide explanation before they submit.
	return &question, nil
}

// CreateQuestionsBulk processes and imports large arrays of questions efficiently.
func (u *questionUseCase) CreateQuestionsBulk(questions []domain.Question) error {
	if len(questions) == 0 {
		return errors.New("The question list is empty")
	}
	batchSize := 100
	err := u.questionRepo.CreateInBatches(questions, batchSize)
	if err != nil {
		return errors.New("Failed to import questions: " + err.Error())
	}
	return nil
}

// ListQuestions retrieves a paginated and optionally filtered list of questions.
func (u *questionUseCase) ListQuestions(tenantID uint, page int, limit int, tag string) (*domain.PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	questions, total, err := u.questionRepo.List(tenantID, limit, offset, tag)
	if err != nil {
		return nil, err
	}
	// Admin might need explanations, but usually we strip sensitive data for standard lists.
	for i := range questions {
		questions[i].CorrectData = nil
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := domain.PaginationMeta{
		CurrentPage:	page,
		PageSize:		limit,
		TotalItems:		total,
		TotalPages:		totalPages,
		HasNextPage:	page < totalPages,
		HasPrevPage:	page > 1,
	}
	return &domain.PaginatedResult{
		Data: 			questions,
		Meta: 			meta,
	}, nil
}