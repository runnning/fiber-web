package query

import "gorm.io/gorm"

// PageOption 分页选项
type PageOption struct {
	Page     int  // 当前页码
	PageSize int  // 每页大小
	NoCount  bool // 是否不统计总数
}

// Apply 实现 Option 接口
func (p PageOption) Apply(db *gorm.DB) *gorm.DB {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}

	offset := (p.Page - 1) * p.PageSize
	return db.Offset(offset).Limit(p.PageSize)
}

// NewPageOption 创建分页选项
func NewPageOption(page, pageSize int, noCount bool) PageOption {
	return PageOption{
		Page:     page,
		PageSize: pageSize,
		NoCount:  noCount,
	}
}

// Result 查询结果
type Result[T any] struct {
	Data       T     `json:"data"`        // 数据列表
	Total      int64 `json:"total"`       // 总记录数
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页大小
	TotalPages int   `json:"total_pages"` // 总页数
}

// NewResult 创建查询结果
func NewResult[T any](data T, total int64, page, pageSize int) *Result[T] {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &Result[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
