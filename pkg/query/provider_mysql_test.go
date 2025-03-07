package query

import (
	"context"
	"errors"
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
		wantCount  int
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
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "带条件的查询",
			setup: func(db *gorm.DB) error {
				users := []User{
					{Name: "User1", Email: "user1@example.com", Age: 20, Status: "active"},
					{Name: "User2", Email: "user2@example.com", Age: 25, Status: "active"},
					{Name: "User3", Email: "user3@example.com", Age: 30, Status: "inactive"},
				}
				return db.Create(&users).Error
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
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "空结果",
			setup: func(db *gorm.DB) error {
				return nil
			},
			buildQuery: func(db *gorm.DB) QueryBuilder {
				return NewMySQLQuery(db.Model(&User{}))
			},
			req: &PageRequest{
				Page:     1,
				PageSize: 10,
			},
			wantTotal: 0,
			wantCount: 0,
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
				// 清理之前的测试数据
				if err := db.Exec("DELETE FROM user").Error; err != nil {
					t.Fatalf("Failed to cleanup previous test data: %v", err)
				}

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

					if len(users) != tt.wantCount {
						t.Errorf("Paginate() result count = %v, want %v", len(users), tt.wantCount)
					}

					if resp.Page != tt.req.Page {
						t.Errorf("Paginate() page = %v, want %v", resp.Page, tt.req.Page)
					}

					if resp.PageSize != tt.req.PageSize {
						t.Errorf("Paginate() pageSize = %v, want %v", resp.PageSize, tt.req.PageSize)
					}
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

func TestMySQLProvider_FindOne(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*gorm.DB) error
		query   func(*gorm.DB) *gorm.DB
		wantErr bool
	}{
		{
			name: "查找存在的记录",
			setup: func(db *gorm.DB) error {
				user := User{Name: "TestUser", Email: "test@example.com", Age: 25, Status: "active"}
				return db.Create(&user).Error
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("name = ?", "TestUser")
			},
			wantErr: false,
		},
		{
			name: "查找不存在的记录",
			setup: func(db *gorm.DB) error {
				return nil
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("name = ?", "NonExistentUser")
			},
			wantErr: true,
		},
	}

	WithTestDB(t, func(db *gorm.DB) {
		if err := db.AutoMigrate(&User{}); err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := db.Exec("DELETE FROM user").Error; err != nil {
					t.Fatalf("Failed to cleanup test data: %v", err)
				}

				if tt.setup != nil {
					if err := tt.setup(db); err != nil {
						t.Fatalf("Failed to setup test data: %v", err)
					}
				}

				provider := NewMySQLProvider[User](db)
				var result User
				err := provider.FindOne(context.Background(), tt.query(db), &result)

				if (err != nil) != tt.wantErr {
					t.Errorf("FindOne() error = %v, wantErr %v", err, tt.wantErr)
				}

				if err == nil && result.Name != "TestUser" {
					t.Errorf("FindOne() got = %v, want %v", result.Name, "TestUser")
				}
			})
		}
	})
}

func TestMySQLProvider_Insert(t *testing.T) {
	tests := []struct {
		name    string
		data    *User
		wantErr bool
	}{
		{
			name: "插入有效记录",
			data: &User{
				Name:   "NewUser",
				Email:  "new@example.com",
				Age:    30,
				Status: "active",
			},
			wantErr: false,
		},
		{
			name:    "插入空记录",
			data:    &User{},
			wantErr: false,
		},
	}

	WithTestDB(t, func(db *gorm.DB) {
		if err := db.AutoMigrate(&User{}); err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := db.Exec("DELETE FROM user").Error; err != nil {
					t.Fatalf("Failed to cleanup test data: %v", err)
				}

				provider := NewMySQLProvider[User](db)
				err := provider.Insert(context.Background(), tt.data)

				if (err != nil) != tt.wantErr {
					t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
				}

				if err == nil {
					var result User
					if err := db.First(&result, tt.data.ID).Error; err != nil {
						t.Errorf("Failed to verify inserted record: %v", err)
					}
				}
			})
		}
	})
}

func TestMySQLProvider_Update(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*gorm.DB) error
		query   func(*gorm.DB) *gorm.DB
		updates map[string]interface{}
		wantErr bool
	}{
		{
			name: "更新存在的记录",
			setup: func(db *gorm.DB) error {
				user := User{Name: "OldName", Email: "old@example.com", Age: 25, Status: "active"}
				return db.Create(&user).Error
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Model(&User{}).Where("name = ?", "OldName")
			},
			updates: map[string]interface{}{
				"name":  "NewName",
				"email": "new@example.com",
			},
			wantErr: false,
		},
		{
			name: "更新不存在的记录",
			setup: func(db *gorm.DB) error {
				return nil
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Model(&User{}).Where("name = ?", "NonExistent")
			},
			updates: map[string]interface{}{
				"name": "NewName",
			},
			wantErr: false,
		},
	}

	WithTestDB(t, func(db *gorm.DB) {
		if err := db.AutoMigrate(&User{}); err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := db.Exec("DELETE FROM user").Error; err != nil {
					t.Fatalf("Failed to cleanup test data: %v", err)
				}

				if tt.setup != nil {
					if err := tt.setup(db); err != nil {
						t.Fatalf("Failed to setup test data: %v", err)
					}
				}

				provider := NewMySQLProvider[User](db)
				err := provider.Update(context.Background(), tt.query(db), tt.updates)

				if (err != nil) != tt.wantErr {
					t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				}

				if err == nil && tt.name == "更新存在的记录" {
					var result User
					if err := db.Where("name = ?", "NewName").First(&result).Error; err != nil {
						t.Errorf("Failed to verify updated record: %v", err)
					}
				}
			})
		}
	})
}

func TestMySQLProvider_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*gorm.DB) error
		query   func(*gorm.DB) *gorm.DB
		wantErr bool
	}{
		{
			name: "删除存在的记录",
			setup: func(db *gorm.DB) error {
				user := User{Name: "ToDelete", Email: "delete@example.com", Age: 25, Status: "active"}
				return db.Create(&user).Error
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("name = ?", "ToDelete")
			},
			wantErr: false,
		},
		{
			name: "删除不存在的记录",
			setup: func(db *gorm.DB) error {
				return nil
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("name = ?", "NonExistent")
			},
			wantErr: false,
		},
	}

	WithTestDB(t, func(db *gorm.DB) {
		if err := db.AutoMigrate(&User{}); err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := db.Exec("DELETE FROM user").Error; err != nil {
					t.Fatalf("Failed to cleanup test data: %v", err)
				}

				if tt.setup != nil {
					if err := tt.setup(db); err != nil {
						t.Fatalf("Failed to setup test data: %v", err)
					}
				}

				provider := NewMySQLProvider[User](db)
				err := provider.Delete(context.Background(), tt.query(db))

				if (err != nil) != tt.wantErr {
					t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				}

				if err == nil && tt.name == "删除存在的记录" {
					var count int64
					if err := db.Model(&User{}).Where("name = ?", "ToDelete").Count(&count).Error; err != nil {
						t.Errorf("Failed to verify deleted record: %v", err)
					}
					if count != 0 {
						t.Errorf("Record was not deleted")
					}
				}
			})
		}
	})
}

func TestMySQLProvider_Transaction(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(context.Context, DataProvider[User]) error
		wantErr bool
	}{
		{
			name: "成功的事务",
			fn: func(ctx context.Context, provider DataProvider[User]) error {
				user := &User{Name: "TransactionUser", Email: "tx@example.com", Age: 25, Status: "active"}
				return provider.Insert(ctx, user)
			},
			wantErr: false,
		},
		{
			name: "失败的事务",
			fn: func(ctx context.Context, provider DataProvider[User]) error {
				return errors.New("transaction error")
			},
			wantErr: true,
		},
	}

	WithTestDB(t, func(db *gorm.DB) {
		if err := db.AutoMigrate(&User{}); err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// 清理之前的测试数据
				if err := db.Exec("DELETE FROM user").Error; err != nil {
					t.Fatalf("Failed to cleanup test data: %v", err)
				}

				provider := NewMySQLProvider[User](db)
				err := provider.Transaction(context.Background(), tt.fn)

				if (err != nil) != tt.wantErr {
					t.Errorf("Transaction() error = %v, wantErr %v", err, tt.wantErr)
				}

				if err == nil && tt.name == "成功的事务" {
					var user User
					if err := db.Where("name = ?", "TransactionUser").First(&user).Error; err != nil {
						t.Errorf("Failed to verify transaction: %v", err)
					}
					if user.Name != "TransactionUser" {
						t.Errorf("Transaction was not committed, user not found")
					}
				}
			})
		}
	})
}

func TestMySQLProvider_parseQuery(t *testing.T) {
	WithTestDB(t, func(db *gorm.DB) {
		provider := NewMySQLProvider[User](db)

		tests := []struct {
			name    string
			query   interface{}
			wantErr bool
		}{
			{
				name:    "nil查询",
				query:   nil,
				wantErr: false,
			},
			{
				name:    "gorm.DB查询",
				query:   db.Where("name = ?", "test"),
				wantErr: false,
			},
			{
				name:    "不支持的查询类型",
				query:   "invalid",
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := provider.parseQuery(tt.query)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseQuery() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}
