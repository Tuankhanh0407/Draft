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
	"log"
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

// GetQuestionForClient fetches a question, prioritizing Redis cache with graceful MySQL fallback.
func (u *questionUseCase) GetQuestionForClient(id uint, tenantID uint) (*domain.Question, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("question:%d:%d", tenantID, id)
	// 1. Attempt to fetch from Redis.
	cached, err := u.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit.
		var q domain.Question
		json.Unmarshal([]byte(cached), &q)
		return &q, nil
	} else if err != redis.Nil {
		// GRACEFUL FALLBACK: Redis is down or network error occurred.
		// Instead of returning HTTP 500, we log the error and let the app query MySQL.
		log.Printf("Warning: Redis cache failed for key %s. Falling back to MySQL. Error: %v\n", cacheKey, err)
	}
	// Note: If err == redis.Nil, it is just a cache miss (normal behavior).
	// 2. Fetch from MySQL (fallback/cache miss).
	question, err := u.questionRepo.GetByIDAndTenant(id, tenantID)
	if err != nil {
		return nil, err // MySQL failed too, return error.
	}
	// 3. Strip sensitive correct answers before returning to the student.
	question.CorrectData = nil
	// 4. Attempt to populate cache for future requests.
	// Fire-and-forget logic: Ignore errors so Redis downtime does not affect user experience.
	go func() {
		qJSON, _ := json.Marshal(question)
		u.redisClient.Set(context.Background(), cacheKey, qJSON, 1 * time.Hour)
	}()
	return question, nil
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