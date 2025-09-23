package query

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page,omitempty"`      // 页码，从1开始
	PageSize int `json:"page_size,omitempty"` // 每页大小
	Offset   int `json:"-"`                   // 偏移量，内部计算
}

// Validate 验证分页参数
func (p *Pagination) Validate() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	p.Offset = (p.Page - 1) * p.PageSize
}

// GetLimit 获取 SQL LIMIT
func (p *Pagination) GetLimit() int {
	return p.PageSize
}

// GetOffset 获取 SQL OFFSET
func (p *Pagination) GetOffset() int {
	return p.Offset
}

// PaginationResult 分页结果
type PaginationResult struct {
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页大小
	Total      int64 `json:"total"`       // 总记录数
	TotalPages int   `json:"total_pages"` // 总页数
}

// NewPaginationResult 创建分页结果
func NewPaginationResult(page, pageSize int, total int64) *PaginationResult {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &PaginationResult{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}