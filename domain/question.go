// Import appropriate package.
package domain

// Import necessary libraries.
import (
	"gorm.io/gorm"
	"time"
)

// Node represents a single element in the AST (abstract syntax tree) for flexible UI rendering.
type Node struct {
	Type			string					`json:"type"` // In example: "text", "gap", "image"...
	Value			string					`json:"value,omitempty"` // Used for "text" nodes.
	ID				string					`json:"id,omitempty"` // Identifier for interactive nodes like gaps.
	Size			int						`json:"size,omitempty"` // Determine input width for gaps.
	Options			map[string]string		`json:"options,omitempty"` // Used for multiple choice or matching options.
}

// QuestionContent holds the instructions, media, and parsed AST nodes.
type QuestionContent struct {
	Instruction		string					`json:"instruction"`
	MediaURL		string					`json:"media_url,omitempty"`
	Nodes			[]Node					`json:"nodes"`
}

// Question represents the universal model for an assessment question.
type Question struct {
	ID				uint					`json:"question_id" gorm:"primaryKey;autoIncrement"`
	TenantID		uint					`json:"tenant_id" gorm:"not null;index" validate:"required"`
	Type			string					`json:"question_type" gorm:"type:varchar(20);not null" validate:"required"` // In example: "GAP_FILL".
	Content			QuestionContent			`json:"content" gorm:"type:json;serializer:json" validate:"required"` // Auto-converted to JSON in MySQL.
	CorrectData		map[string][]string		`json:"correct_data,omitempty" gorm:"type:json;serializer:json"` // Hidden in client responses.
	Tags			[]string				`json:"tags,omitempty" gorm:"type:json;serializer:json"` // Array of tags for categorization/filtering.
	Explanation		string					`json:"explanation,omitempty" gorm:"type:text"`
	CreatedAt		time.Time				`json:"created_at"`
	UpdatedAt		time.Time				`json:"updated_at"`
	DeletedAt		gorm.DeletedAt			`json:"-" gorm:"index"` // Soft delete.
}

// QuestionRepository defines database operations for a question.
type QuestionRepository interface {
	// Create inserts a single question into the database.
	Create(question *Question) error
	// GetByIDAndTenant fetches a specific question ensuring tenant isolation.
	GetByIDAndTenant(id uint, tenantID uint) (*Question, error)
	// List retrieves a paginated array of questions, optionally filtered by a tag.
	List(tenantID uint, limit int, offset int, tag string) ([]Question, int64, error)
	// CreateInBatches inserts multiple questions simultaneously to optimize database performance.
	CreateInBatches(questions []Question, batchSize int) error
}

// QuestionUseCase defines business logic for questions.
type QuestionUseCase interface {
	// CreateQuestion validates and processes a new question.
	CreateQuestion(question *Question) error
	// GetQuestionForClient fetches a question but strips out sensitive correct answers and explanations.
	GetQuestionForClient(id uint, tenantID uint) (*Question, error)
	// ListQuestions handles pagination logic and formatting for question lists.
	ListQuestions(tenantID uint, page int, limit int, tag string) (*PaginatedResult, error)
	// CreateQuestionsBulk validates and imports an array of questions in one transaction.
	CreateQuestionsBulk(questions []Question) error
}