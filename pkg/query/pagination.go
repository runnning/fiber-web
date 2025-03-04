package query

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// 分页相关常量
const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// PageOption 分页选项
type PageOption struct {
	Page     int  // 当前页码
	PageSize int  // 每页大小
	NoCount  bool // 是否不统计总数
}

// NewPageOption 创建分页选项
func NewPageOption(page, pageSize int, noCount bool) *PageOption {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return &PageOption{
		Page:     page,
		PageSize: pageSize,
		NoCount:  noCount,
	}
}

// Apply 应用MySQL分页
func (p *PageOption) Apply(db *gorm.DB) *gorm.DB {
	offset := (p.Page - 1) * p.PageSize
	return db.Offset(offset).Limit(p.PageSize)
}

// ApplyMongo 应用MongoDB分页
func (p *PageOption) ApplyMongo(opts *options.FindOptions, _ *bson.D) {
	skip := int64((p.Page - 1) * p.PageSize)
	limit := int64(p.PageSize)
	opts.SetSkip(skip)
	opts.SetLimit(limit)
}

// Result 查询结果
type Result[T any] struct {
	Data      interface{} `json:"data"`       // 数据列表
	Total     int64       `json:"total"`      // 总记录数
	Page      int         `json:"page"`       // 当前页码
	PageSize  int         `json:"page_size"`  // 每页大小
	TotalPage int         `json:"total_page"` // 总页数
}

// NewResult 创建查询结果
func NewResult[T any](data T, total int64, page, pageSize int) *Result[T] {
	totalPage := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPage++
	}

	return &Result[T]{
		Data:      data,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}
}
