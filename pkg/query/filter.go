package query

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Operator 查询操作符
type Operator string

const (
	OpEq         Operator = "eq"         // 等于
	OpNe         Operator = "ne"         // 不等于
	OpGt         Operator = "gt"         // 大于
	OpGte        Operator = "gte"        // 大于等于
	OpLt         Operator = "lt"         // 小于
	OpLte        Operator = "lte"        // 小于等于
	OpIn         Operator = "in"         // 在数组中
	OpNin        Operator = "nin"        // 不在数组中
	OpContains   Operator = "contains"   // 包含（字符串模糊查询）
	OpStartsWith Operator = "startsWith" // 以...开始
	OpEndsWith   Operator = "endsWith"   // 以...结束
	OpExists     Operator = "exists"     // 字段存在
)

// FilterCondition 单个过滤条件
type FilterCondition struct {
	Field    string      `json:"field"`            // 字段名
	Operator Operator    `json:"operator"`         // 操作符
	Value    interface{} `json:"value,omitempty"`  // 值
	Values   []string    `json:"values,omitempty"` // 值数组（用于in/nin操作符）
}

// FilterBuilder 查询条件构建器
type FilterBuilder struct {
	conditions []FilterCondition
}

// NewFilterBuilder 创建新的查询条件构建器
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		conditions: make([]FilterCondition, 0),
	}
}

// AddCondition 添加查询条件
func (fb *FilterBuilder) AddCondition(field string, op Operator, value interface{}) *FilterBuilder {
	fb.conditions = append(fb.conditions, FilterCondition{
		Field:    field,
		Operator: op,
		Value:    value,
	})
	return fb
}

// AddArrayCondition 添加数组类型的查询条件
func (fb *FilterBuilder) AddArrayCondition(field string, op Operator, values []string) *FilterBuilder {
	fb.conditions = append(fb.conditions, FilterCondition{
		Field:    field,
		Operator: op,
		Values:   values,
	})
	return fb
}

// Build 构建MongoDB的查询条件
func (fb *FilterBuilder) Build() bson.M {
	filter := bson.M{}

	// 创建一个临时映射来合并同一字段的多个条件
	fieldConditions := make(map[string]bson.M)

	for _, condition := range fb.conditions {
		// 获取或创建字段的条件映射
		fieldFilter, exists := fieldConditions[condition.Field]
		if !exists {
			fieldFilter = bson.M{}
		}

		switch condition.Operator {
		case OpEq:
			// 等于操作直接设置字段值
			filter[condition.Field] = condition.Value
		case OpNe:
			fieldFilter["$ne"] = condition.Value
		case OpGt:
			fieldFilter["$gt"] = condition.Value
		case OpGte:
			fieldFilter["$gte"] = condition.Value
		case OpLt:
			fieldFilter["$lt"] = condition.Value
		case OpLte:
			fieldFilter["$lte"] = condition.Value
		case OpIn:
			fieldFilter["$in"] = condition.Values
		case OpNin:
			fieldFilter["$nin"] = condition.Values
		case OpContains:
			if strVal, ok := condition.Value.(string); ok {
				fieldFilter["$regex"] = primitive.Regex{Pattern: strVal, Options: "i"}
			}
		case OpStartsWith:
			if strVal, ok := condition.Value.(string); ok {
				fieldFilter["$regex"] = primitive.Regex{Pattern: "^" + strVal, Options: "i"}
			}
		case OpEndsWith:
			if strVal, ok := condition.Value.(string); ok {
				fieldFilter["$regex"] = primitive.Regex{Pattern: strVal + "$", Options: "i"}
			}
		case OpExists:
			if boolVal, ok := condition.Value.(bool); ok {
				fieldFilter["$exists"] = boolVal
			}
		}

		// 如果不是等于操作，更新字段条件映射
		if condition.Operator != OpEq {
			fieldConditions[condition.Field] = fieldFilter
		}
	}

	// 将所有字段条件合并到最终过滤器中
	for field, conditions := range fieldConditions {
		if len(conditions) > 0 && filter[field] == nil {
			filter[field] = conditions
		}
	}

	return filter
}

// ParseTimeRange 解析时间范围
func ParseTimeRange(field string, startTime, endTime *time.Time) []FilterCondition {
	var conditions []FilterCondition

	if startTime != nil {
		conditions = append(conditions, FilterCondition{
			Field:    field,
			Operator: OpGte,
			Value:    startTime,
		})
	}

	if endTime != nil {
		conditions = append(conditions, FilterCondition{
			Field:    field,
			Operator: OpLte,
			Value:    endTime,
		})
	}

	return conditions
}

// ParseSearchText 解析搜索文本（支持多字段模糊搜索）
func ParseSearchText(searchText string, fields []string) []FilterCondition {
	if searchText == "" || len(fields) == 0 {
		return nil
	}

	searchText = strings.TrimSpace(searchText)
	var conditions []FilterCondition

	for _, field := range fields {
		conditions = append(conditions, FilterCondition{
			Field:    field,
			Operator: OpContains,
			Value:    searchText,
		})
	}

	return conditions
}
