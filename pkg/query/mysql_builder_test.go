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

	joinSQL := "LEFT JOIN orders ON users.id = orders.user_id"
	query.Join(joinSQL)

	if len(query.joins) != 1 || query.joins[0] != joinSQL {
		t.Errorf("Expected joins to contain [%s], got %v", joinSQL, query.joins)
	}
}

// TestMySQLQuery_Conditions 测试查询条件功能
func TestMySQLQuery_Conditions(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*MySQLQuery)
		wantField string
		wantOp    Operator
		wantValue interface{}
	}{
		{
			name: "等于条件",
			setup: func(q *MySQLQuery) {
				q.AddCondition("status", OpEq, "active")
			},
			wantField: "status",
			wantOp:    OpEq,
			wantValue: "active",
		},
		{
			name: "大于等于条件",
			setup: func(q *MySQLQuery) {
				q.AddCondition("age", OpGte, 18)
			},
			wantField: "age",
			wantOp:    OpGte,
			wantValue: 18,
		},
		{
			name: "包含条件",
			setup: func(q *MySQLQuery) {
				q.AddCondition("name", OpContains, "john")
			},
			wantField: "name",
			wantOp:    OpContains,
			wantValue: "john",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewTestHelper(t).GetDB()
			query := NewMySQLQuery(db)
			tt.setup(query)

			if len(query.conditions) != 1 {
				t.Fatalf("Expected 1 condition, got %d", len(query.conditions))
			}

			cond := query.conditions[0]
			if cond.Field != tt.wantField {
				t.Errorf("Expected field %s, got %s", tt.wantField, cond.Field)
			}
			if cond.Operator != tt.wantOp {
				t.Errorf("Expected operator %s, got %s", tt.wantOp, cond.Operator)
			}
			if cond.Value != tt.wantValue {
				t.Errorf("Expected value %v, got %v", tt.wantValue, cond.Value)
			}
		})
	}
}

// TestBuildSearchQuery 测试搜索查询构建功能
func TestBuildSearchQuery(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	searchText := "john"
	fields := []string{"name", "email"}

	result := BuildSearchQuery(db, searchText, fields)
	if result == nil {
		t.Error("Expected non-nil result from BuildSearchQuery")
	}
}

// TestBuildTimeRangeQuery 测试时间范围查询构建功能
func TestBuildTimeRangeQuery(t *testing.T) {
	db := NewTestHelper(t).GetDB()
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	endTime := now

	result := BuildTimeRangeQuery(db, "created_at", startTime, endTime)
	if result == nil {
		t.Error("Expected non-nil result from BuildTimeRangeQuery")
	}
}
