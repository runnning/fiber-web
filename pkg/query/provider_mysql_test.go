package query

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
)

// 模拟用户结构
type User struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"size:100"`
	Email     string `gorm:"size:100"`
	Age       int
	Status    string `gorm:"size:20"`
	CreatedAt time.Time
}

func TestMySQLPaginate(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*gorm.DB) error
		buildQuery func(*gorm.DB) QueryBuilder
		req        *PageRequest
		wantTotal  int64
		wantErr    bool
	}{
		{
			name: "正常分页",
			setup: func(db *gorm.DB) error {
				users := []User{
					{Name: "User1", Email: "user1@example.com", Age: 20, Status: "active"},
					{Name: "User2", Email: "user2@example.com", Age: 25, Status: "active"},
					{Name: "User3", Email: "user3@example.com", Age: 30, Status: "inactive"},
				}
				return db.Create(&users).Error
			},
			buildQuery: func(db *gorm.DB) QueryBuilder {
				return NewMySQLQuery(db.Model(&User{}))
			},
			req: &PageRequest{
				Page:     1,
				PageSize: 2,
				OrderBy:  "id",
				Order:    "DESC",
			},
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name: "带条件的查询",
			setup: func(db *gorm.DB) error {
				// 使用上一个测试的数据
				return nil
			},
			buildQuery: func(db *gorm.DB) QueryBuilder {
				query := NewMySQLQuery(db.Model(&User{}))
				query.WhereSimple("status", OpEq, "active")
				return query
			},
			req: &PageRequest{
				Page:     1,
				PageSize: 10,
				OrderBy:  "age",
				Order:    "ASC",
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name: "空结果",
			setup: func(db *gorm.DB) error {
				return db.Exec("DELETE FROM users").Error
			},
			buildQuery: func(db *gorm.DB) QueryBuilder {
				return NewMySQLQuery(db.Model(&User{}))
			},
			req: &PageRequest{
				Page:     1,
				PageSize: 10,
			},
			wantTotal: 0,
			wantErr:   false,
		},
	}

	WithTestDB(t, func(db *gorm.DB) {
		// 创建测试表
		if err := db.AutoMigrate(&User{}); err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// 准备测试数据
				if tt.setup != nil {
					if err := tt.setup(db); err != nil {
						t.Fatalf("Failed to setup test data: %v", err)
					}
				}

				var users []User
				query := tt.buildQuery(db)
				provider := NewMySQLProvider[User](db)

				ctx := context.Background()
				resp, err := Paginate(ctx, query, provider, tt.req, &users)

				if (err != nil) != tt.wantErr {
					t.Errorf("Paginate() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if err == nil {
					if resp.Total != tt.wantTotal {
						t.Errorf("Paginate() total = %v, want %v", resp.Total, tt.wantTotal)
					}

					if resp.Page != tt.req.Page {
						t.Errorf("Paginate() page = %v, want %v", resp.Page, tt.req.Page)
					}

					if resp.PageSize != tt.req.PageSize {
						t.Errorf("Paginate() pageSize = %v, want %v", resp.PageSize, tt.req.PageSize)
					}

					if len(resp.List) == 0 && tt.wantTotal > 0 {
						t.Error("Paginate() expected non-empty result")
					}
				}

				// 清理测试数据
				if err := db.Exec("DELETE FROM users").Error; err != nil {
					t.Errorf("Failed to cleanup test data: %v", err)
				}
			})
		}
	})
}

func TestNewPageRequest(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{
			name:         "正常参数",
			page:         2,
			pageSize:     15,
			wantPage:     2,
			wantPageSize: 15,
		},
		{
			name:         "页码为0",
			page:         0,
			pageSize:     10,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name:         "页大小为0",
			page:         1,
			pageSize:     0,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name:         "负数参数",
			page:         -1,
			pageSize:     -5,
			wantPage:     1,
			wantPageSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewPageRequest(tt.page, tt.pageSize)

			if req.Page != tt.wantPage {
				t.Errorf("NewPageRequest() page = %v, want %v", req.Page, tt.wantPage)
			}

			if req.PageSize != tt.wantPageSize {
				t.Errorf("NewPageRequest() pageSize = %v, want %v", req.PageSize, tt.wantPageSize)
			}
		})
	}
}

func TestPageRequest_Offset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		want     int
	}{
		{
			name:     "第一页",
			page:     1,
			pageSize: 10,
			want:     0,
		},
		{
			name:     "第二页",
			page:     2,
			pageSize: 10,
			want:     10,
		},
		{
			name:     "自定义页大小",
			page:     3,
			pageSize: 15,
			want:     30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &PageRequest{
				Page:     tt.page,
				PageSize: tt.pageSize,
			}

			if got := req.Offset(); got != tt.want {
				t.Errorf("PageRequest.Offset() = %v, want %v", got, tt.want)
			}
		})
	}
}
