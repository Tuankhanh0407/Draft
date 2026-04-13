// Import appropriate package.
package domain

// PaginationMeta holds the metadata required for frontend paginators.
type PaginationMeta struct {
	CurrentPage		int				`json:"current_page"`
	PageSize		int				`json:"page_size"`
	TotalItems		int64			`json:"total_items"`
	TotalPages		int				`json:"total_pages"`
	HasNextPage		bool			`json:"has_next_page"`
	HasPrevPage		bool			`json:"has_prev_page"`
}

// PaginatedResult is the standard, unified response wrapper for all list/array APIs.
type PaginatedResult struct {
	Data			interface{}		`json:"data"` // Hold the generic slice of entities (Questions, Exams...).
	Meta			PaginationMeta	`json:"meta"` // Hold the pagination context.
}