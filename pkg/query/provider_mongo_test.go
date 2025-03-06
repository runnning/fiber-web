package query

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoUser 用户文档结构
type MongoUser struct {
	ID        string    `bson:"_id,omitempty"`
	Name      string    `bson:"name"`
	Email     string    `bson:"email"`
	Age       int       `bson:"age"`
	Status    string    `bson:"status"`
	CreatedAt time.Time `bson:"created_at"`
}

func TestMongoPaginate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(context.Context, *mongo.Collection) error
		req       *PageRequest
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "正常分页",
			setup: func(ctx context.Context, coll *mongo.Collection) error {
				users := []interface{}{
					MongoUser{
						Name:      "User1",
						Email:     "user1@example.com",
						Age:       20,
						Status:    "active",
						CreatedAt: time.Now(),
					},
					MongoUser{
						Name:      "User2",
						Email:     "user2@example.com",
						Age:       25,
						Status:    "active",
						CreatedAt: time.Now(),
					},
					MongoUser{
						Name:      "User3",
						Email:     "user3@example.com",
						Age:       30,
						Status:    "inactive",
						CreatedAt: time.Now(),
					},
				}
				_, err := coll.InsertMany(ctx, users)
				return err
			},
			req: &PageRequest{
				Page:     1,
				PageSize: 2,
				OrderBy:  "age",
				Order:    "DESC",
			},
			wantTotal: 3,
			wantErr:   false,
		},
		{
			name: "空结果",
			setup: func(ctx context.Context, coll *mongo.Collection) error {
				return coll.Drop(ctx)
			},
			req: &PageRequest{
				Page:     1,
				PageSize: 10,
			},
			wantTotal: 0,
			wantErr:   false,
		},
	}

	// 使用测试辅助函数创建MongoDB实例
	WithTestMongoDB(t, func(db *mongo.Database) {
		ctx := context.Background()
		coll := db.Collection("users")

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// 准备测试数据
				if tt.setup != nil {
					if err := tt.setup(ctx, coll); err != nil {
						t.Fatalf("Failed to setup test data: %v", err)
					}
				}

				var users []MongoUser
				query := NewMongoQuery()
				provider := NewMongoProvider[MongoUser](coll)

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
			})
		}
	})
}

func TestMongoSearchQuery(t *testing.T) {
	// 使用测试辅助函数创建MongoDB实例
	WithTestMongoDB(t, func(db *mongo.Database) {
		ctx := context.Background()
		coll := db.Collection("users")

		// 准备测试数据
		users := []interface{}{
			MongoUser{
				Name:   "John Doe",
				Email:  "john@example.com",
				Status: "active",
			},
			MongoUser{
				Name:   "Jane Smith",
				Email:  "jane@example.com",
				Status: "inactive",
			},
		}
		_, err := coll.InsertMany(ctx, users)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}

		// 确保测试结束后清理数据
		defer func() {
			if err := coll.Drop(ctx); err != nil {
				t.Errorf("Failed to cleanup test data: %v", err)
			}
		}()

		tests := []struct {
			name      string
			search    string
			fields    []string
			wantCount int64
		}{
			{
				name:      "按名字搜索",
				search:    "John",
				fields:    []string{"name"},
				wantCount: 1,
			},
			{
				name:      "按邮箱搜索",
				search:    "example.com",
				fields:    []string{"email"},
				wantCount: 2,
			},
			{
				name:      "多字段搜索",
				search:    "jane",
				fields:    []string{"name", "email"},
				wantCount: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				conditions := ParseSearchText(tt.search, tt.fields)
				filter := NewFilterBuilder()
				for _, condition := range conditions {
					filter.AddCondition(condition.Field, condition.Operator, condition.Value)
				}

				count, err := coll.CountDocuments(ctx, filter.Build())
				if err != nil {
					t.Errorf("Failed to count documents: %v", err)
				}

				if count != tt.wantCount {
					t.Errorf("Search count = %v, want %v", count, tt.wantCount)
				}
			})
		}
	})
}

func TestMongoTimeRangeQuery(t *testing.T) {
	// 使用测试辅助函数创建MongoDB实例
	WithTestMongoDB(t, func(db *mongo.Database) {
		ctx := context.Background()
		coll := db.Collection("users")

		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		tomorrow := now.Add(24 * time.Hour)

		// 准备测试数据
		users := []interface{}{
			MongoUser{
				Name:      "User1",
				CreatedAt: yesterday,
			},
			MongoUser{
				Name:      "User2",
				CreatedAt: now,
			},
			MongoUser{
				Name:      "User3",
				CreatedAt: tomorrow,
			},
		}
		_, err := coll.InsertMany(ctx, users)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}

		// 确保测试结束后清理数据
		defer func() {
			if err := coll.Drop(ctx); err != nil {
				t.Errorf("Failed to cleanup test data: %v", err)
			}
		}()

		tests := []struct {
			name      string
			start     *time.Time
			end       *time.Time
			wantCount int64
		}{
			{
				name:      "全部时间范围",
				start:     &yesterday,
				end:       &tomorrow,
				wantCount: 3,
			},
			{
				name:      "只有开始时间",
				start:     &now,
				end:       nil,
				wantCount: 2,
			},
			{
				name:      "只有结束时间",
				start:     nil,
				end:       &now,
				wantCount: 2,
			},
			{
				name:      "精确时间范围",
				start:     &yesterday,
				end:       &now,
				wantCount: 2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				conditions := ParseTimeRange("created_at", tt.start, tt.end)
				filter := NewFilterBuilder()
				for _, condition := range conditions {
					filter.AddCondition(condition.Field, condition.Operator, condition.Value)
				}

				count, err := coll.CountDocuments(ctx, filter.Build())
				if err != nil {
					t.Errorf("Failed to count documents: %v", err)
				}

				if count != tt.wantCount {
					t.Errorf("Time range count = %v, want %v", count, tt.wantCount)
				}
			})
		}
	})
}
