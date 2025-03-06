package query

import (
	"testing"
	"time"
)

// TestMySQLQuery_Select 测试字段选择功能
func TestMySQLQuery_Select(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	query := NewMySQLQuery(db)

	// 测试单个字段
	query.Select("name")
	if len(query.fields) != 1 || query.fields[0] != "name" {
		t.Errorf("Expected fields to contain ['name'], got %v", query.fields)
	}

	// 测试多个字段
	query.Select("age", "email")
	if len(query.fields) != 3 || query.fields[2] != "email" {
		t.Errorf("Expected fields to contain ['name', 'age', 'email'], got %v", query.fields)
	}
}

// TestMySQLQuery_Join 测试连接查询功能
func TestMySQLQuery_Join(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	query := NewMySQLQuery(db)

	query.Join("orders", "users.id = orders.user_id")
	if len(query.joins) != 1 {
		t.Errorf("Expected joins to contain 1 item, got %d", len(query.joins))
	}
}

// TestMySQLQuery_WhereSimple 测试简单条件功能
func TestMySQLQuery_WhereSimple(t *testing.T) {
	tests := []struct {
		name  string
		field string
		op    Operator
		value interface{}
	}{
		{
			name:  "等于条件",
			field: "status",
			op:    OpEq,
			value: "active",
		},
		{
			name:  "大于等于条件",
			field: "age",
			op:    OpGte,
			value: 18,
		},
		{
			name:  "包含条件",
			field: "name",
			op:    OpContains,
			value: "john",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewTestHelper(t).GetDB()
			query := NewMySQLQuery(db)
			query.WhereSimple(tt.field, tt.op, tt.value)

			if len(query.conditions) != 1 {
				t.Fatalf("Expected 1 condition, got %d", len(query.conditions))
			}

			cond, ok := query.conditions[0].(*SimpleCondition)
			if !ok {
				t.Fatalf("Expected SimpleCondition, got %T", query.conditions[0])
			}

			if cond.Field != tt.field {
				t.Errorf("Expected field %s, got %s", tt.field, cond.Field)
			}
			if cond.Operator != tt.op {
				t.Errorf("Expected operator %s, got %s", tt.op, cond.Operator)
			}
			if cond.Value != tt.value {
				t.Errorf("Expected value %v, got %v", tt.value, cond.Value)
			}
		})
	}
}

// TestMySQLQuery_WhereGroup 测试条件组功能
func TestMySQLQuery_WhereGroup(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	query := NewMySQLQuery(db)

	// 创建OR条件组
	cond1 := NewEqCondition("status", "active")
	cond2 := NewGteCondition("age", 18)
	query.WhereGroup(LogicOr, cond1, cond2)

	if len(query.conditions) != 1 {
		t.Fatalf("Expected 1 condition, got %d", len(query.conditions))
	}

	group, ok := query.conditions[0].(*GroupCondition)
	if !ok {
		t.Fatalf("Expected GroupCondition, got %T", query.conditions[0])
	}

	if group.Logic != LogicOr {
		t.Errorf("Expected logic OR, got %s", group.Logic)
	}

	if len(group.Conditions) != 2 {
		t.Errorf("Expected 2 sub-conditions, got %d", len(group.Conditions))
	}
}

// TestMySQLQuery_OrderBy 测试排序功能
func TestMySQLQuery_OrderBy(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	query := NewMySQLQuery(db)

	query.OrderBy("created_at", "DESC")
	if len(query.orders) != 1 {
		t.Fatalf("Expected 1 order, got %d", len(query.orders))
	}

	if query.orders[0] != "created_at DESC" {
		t.Errorf("Expected order 'created_at DESC', got '%s'", query.orders[0])
	}
}

// TestMySQLQuery_Limit 测试限制功能
func TestMySQLQuery_Limit(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	query := NewMySQLQuery(db)

	query.Limit(10)
	if query.limit != 10 {
		t.Errorf("Expected limit 10, got %d", query.limit)
	}
}

// TestMySQLQuery_Offset 测试偏移功能
func TestMySQLQuery_Offset(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	query := NewMySQLQuery(db)

	query.Offset(20)
	if query.offset != 20 {
		t.Errorf("Expected offset 20, got %d", query.offset)
	}
}

// TestMySQLQuery_Build 测试构建查询功能
func TestMySQLQuery_Build(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	query := NewMySQLQuery(db)

	query.WhereSimple("status", OpEq, "active")
	query.WhereSimple("age", OpGte, 18)
	query.OrderBy("created_at", "DESC")
	query.Limit(10)
	query.Offset(20)

	result := query.Build()
	if result == nil {
		t.Error("Expected non-nil result from Build()")
	}
}

// TestNewTimeRangeCondition 测试时间范围条件创建
func TestNewTimeRangeCondition(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// 测试两个时间点都提供
	condition := NewTimeRangeCondition("created_at", &yesterday, &now)
	if condition == nil {
		t.Fatal("Expected non-nil condition")
	}

	group, ok := condition.(*GroupCondition)
	if !ok {
		t.Fatalf("Expected GroupCondition, got %T", condition)
	}

	if group.Logic != LogicAnd {
		t.Errorf("Expected logic AND, got %s", group.Logic)
	}

	if len(group.Conditions) != 2 {
		t.Errorf("Expected 2 sub-conditions, got %d", len(group.Conditions))
	}

	// 测试只提供开始时间
	condition = NewTimeRangeCondition("created_at", &yesterday, nil)
	if condition == nil {
		t.Fatal("Expected non-nil condition")
	}

	simple, ok := condition.(*SimpleCondition)
	if !ok {
		t.Fatalf("Expected SimpleCondition, got %T", condition)
	}

	if simple.Operator != OpGte {
		t.Errorf("Expected operator GTE, got %s", simple.Operator)
	}
}

// TestNewSearchCondition 测试搜索条件创建
func TestNewSearchCondition(t *testing.T) {
	// 测试单字段搜索
	condition := NewSearchCondition("john", []string{"name"})
	if condition == nil {
		t.Fatal("Expected non-nil condition")
	}

	simple, ok := condition.(*SimpleCondition)
	if !ok {
		t.Fatalf("Expected SimpleCondition, got %T", condition)
	}

	if simple.Operator != OpContains {
		t.Errorf("Expected operator CONTAINS, got %s", simple.Operator)
	}

	// 测试多字段搜索
	condition = NewSearchCondition("john", []string{"name", "email"})
	if condition == nil {
		t.Fatal("Expected non-nil condition")
	}

	group, ok := condition.(*GroupCondition)
	if !ok {
		t.Fatalf("Expected GroupCondition, got %T", condition)
	}

	if group.Logic != LogicOr {
		t.Errorf("Expected logic OR, got %s", group.Logic)
	}

	if len(group.Conditions) != 2 {
		t.Errorf("Expected 2 sub-conditions, got %d", len(group.Conditions))
	}
}
