// Import appropriate package.
package usecases

// Import necessary libraries.
import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/datatypes"
	"letuan.com/code_demo_backend/domain"
)

// questionUsecase implements the business logic for managing questions.
type questionUsecase struct {
	repo domain.QuestionRepository
}

// NewQuestionUsecase initializes the question usecase with its required repository.
func NewQuestionUsecase(repo domain.QuestionRepository) domain.QuestionUsecase {
	return &questionUsecase{repo}
}

// validateAST verifies the structural integrity of question content, ensuring that gap nodes match the provided answers.
func validateAST(req *domain.QuestionRequest) error {
	if req.Type == "GAP_FILL" {
		gapCount := 0
		for _, node := range req.Content.Nodes {
			if node.Type == "gap" {
				if node.ID == "" || node.Size <= 0 {
					return errors.New("Gap node must have a valid 'id' and 'size' > 0")
				}
				gapCount++
			}
		}
		if gapCount != len(req.CorrectData) {
			return errors.New("Number of gap nodes does not match correct_data")
		}
	}
	return nil
}

// mapToResponse converts a database entity into a client-safe response, intentionally omitting sensitive fields like CorrectData to prevent cheating.
func mapToResponse(q domain.Question) domain.QuestionResponse {
	var content domain.QuestionContent
	var tags []string
	_ = json.Unmarshal(q.Content, &content)
	_ = json.Unmarshal(q.Tags, &tags)
	return domain.QuestionResponse{
		ID:			q.ID,
		Type:		q.Type,
		Tags:		tags,
		Content:	content,
		CreatedAt:	q.CreatedAt,
	}
}

// mapToEntity transforms a question request payload into a database-ready entity, serializing JSON fields.
func mapToEntity(tenantID uint, req *domain.QuestionRequest) (domain.Question, error) {
	contentBytes, _ := json.Marshal(req.Content)
	correctBytes, _ := json.Marshal(req.CorrectData)
	tagsBytes, _ := json.Marshal(req.Tags)
	return domain.Question{
		TenantID:		tenantID,
		Type:			req.Type,
		Tags:			datatypes.JSON(tagsBytes),
		Content:		datatypes.JSON(contentBytes),
		CorrectData: 	datatypes.JSON(correctBytes),
	}, nil
}

// GetAll retrieves a paginated and optionally filtered list of safe question responses for a specific tenant.
func (u *questionUsecase) GetAll(ctx context.Context, tenantID uint, qType, tag string, page, limit int) ([]domain.QuestionResponse, error) {
	offset := (page - 1) * limit
	questions, err := u.repo.Fetch(ctx, tenantID, qType, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	var res []domain.QuestionResponse
	for _, q := range questions {
		res = append(res, mapToResponse(q))
	}
	return res, nil
}

// GetByID fetches a single question by its ID and ensures it belongs to the specified tenant.
func (u *questionUsecase) GetByID(ctx context.Context, tenantID, id uint) (domain.QuestionResponse, error) {
	question, err := u.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return domain.QuestionResponse{}, errors.New("Question not found")
	}
	return mapToResponse(question), nil
}

// Create validates the payload structure and registers a new question into the system.
func (u *questionUsecase) Create(ctx context.Context, tenantID uint, req *domain.QuestionRequest) (domain.QuestionResponse, error) {
	if err := validateAST(req); err != nil {
		return domain.QuestionResponse{}, err
	}
	entity, _ := mapToEntity(tenantID, req)
	if err := u.repo.Create(ctx, &entity); err != nil {
		return domain.QuestionResponse{}, err
	}
	return mapToResponse(entity), nil
}

// CreateBulk validates and inserts multiple questions into the database in a single batch.
func (u *questionUsecase) CreateBulk(ctx context.Context, tenantID uint, reqs []domain.QuestionRequest) error {
	var entities []domain.Question
	for _, req := range reqs {
		if err := validateAST(&req); err != nil {
			return errors.New("Validation failed for one or more items: " + err.Error())
		}
		entity, _ := mapToEntity(tenantID, &req)
		entities = append(entities, entity)
	}
	return u.repo.CreateBulk(ctx, entities)
}

// Update modifies an existing question after fully validating the new payload data.
func (u *questionUsecase) Update(ctx context.Context, tenantID, id uint, req *domain.QuestionRequest) (domain.QuestionResponse, error) {
	if err := validateAST(req); err != nil {
		return domain.QuestionResponse{}, err
	}
	existing, err := u.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return domain.QuestionResponse{}, errors.New("Question not found")
	}
	updates, _ := mapToEntity(tenantID, req)
	existing.Type = updates.Type
	existing.Tags = updates.Tags
	existing.Content = updates.Content
	existing.CorrectData = updates.CorrectData
	if err := u.repo.Update(ctx, &existing); err != nil {
		return domain.QuestionResponse{}, err
	}
	return mapToResponse(existing), nil
}

// Delete removes a question from the specified tenant's data.
func (u *questionUsecase) Delete(ctx context.Context, tenantID, id uint) error {
	return u.repo.Delete(ctx, tenantID, id)
}