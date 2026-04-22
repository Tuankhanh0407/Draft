// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"context"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// Constants for strictly allowed question types.
const (
	TypeGapFilling = "GAP_FILLING"
	TypeMultipleChoice = "MULTIPLE_CHOICE"
	TypeMatching = "MATCHING"
)

// Question represents a test question in the database.
type Question struct {
	ID				uint					`json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID		uint					`json:"-" gorm:"index:idx_tenant_type;not null"` // Hidden from JSON.
	Type			string					`json:"type" gorm:"index:idx_tenant_type;size:50;not null"`
	Tags			datatypes.JSON			`json:"tags" gorm:"type:json"`
	Content			datatypes.JSON			`json:"content" gorm:"type:json;not null"`
	CorrectData 	datatypes.JSON			`json:"-" gorm:"type:json;not null"` // Strictly hidden from frontend clients.
	CreatedAt		time.Time				`json:"created_at"`
	UpdatedAt		time.Time				`json:"updated_at"`
	DeletedAt		gorm.DeletedAt			`json:"-" gorm:"index"`
}

// ASTNode represents a single element in the question's content array.
type ASTNode struct {
	Type			string					`json:"type"`
	Value			string					`json:"value,omitempty"`
	ID				string					`json:"id,omitempty"`
	Size			int						`json:"size,omitempty"`
}

// QuestionContent represents the structured payload of the question text.
type QuestionContent struct {
	Instruction		string					`json:"instruction"`
	MediaURL		string					`json:"media_url,omitempty"`
	Nodes			[]ASTNode				`json:"nodes"`
}

// QuestionRequest is the payload expected from the frontend when creating or updating.
type QuestionRequest struct {
	Type			string					`json:"type" validate:"required"`
	Tags			[]string				`json:"tags"`
	Content			QuestionContent			`json:"content" validate:"required"`
	CorrectData		map[string]interface{}	`json:"correct_data" validate:"required"`
}

// QuestionResponse is the safe payload returned to the frontend (anti-cheat).
type QuestionResponse struct {
	ID				uint					`json:"id"`
	Type			string					`json:"type"`
	Tags			[]string				`json:"tags"`
	Content			QuestionContent			`json:"content"`
	CreatedAt		time.Time				`json:"created_at"`
}

// QuestionRepository defines data access methods for questions.
type QuestionRepository interface {
	Fetch(ctx context.Context, tenantID uint, qType, tag string, limit, offset int) ([]Question, error)
	GetByID(ctx context.Context, tenantID, id uint) (Question, error)
	Create(ctx context.Context, question *Question) error
	CreateBulk(ctx context.Context, questions []Question) error
	Update(ctx context.Context, question *Question) error
	Delete(ctx context.Context, tenantID, id uint) error
}

// QuestionUsecase defines business logic for question management.
type QuestionUsecase interface {
	GetAll(ctx context.Context, tenantID uint, qType, tag string, page, limit int) ([]QuestionResponse, error)
	GetByID(ctx context.Context, tenantID, id uint) (QuestionResponse, error)
	Create(ctx context.Context, tenantID uint, req *QuestionRequest) (QuestionResponse, error)
	CreateBulk(ctx context.Context, tenantID uint, reqs []QuestionRequest) error
	Update(ctx context.Context, tenantID, id uint, req *QuestionRequest) (QuestionResponse, error)
	Delete(ctx context.Context, tenantID, id uint) error
}