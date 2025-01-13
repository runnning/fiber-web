package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 定义常量
const (
	dirEntity     = "entity"
	dirRepository = "repository"
	dirUsecase    = "usecase"
	dirEndpoint   = "endpoint"
	dirSQL        = "sql"
)

var (
	// SQL类型映射
	typeMap = map[string]string{
		// 基础类型
		"string":  "VARCHAR(255)",
		"int":     "INT",
		"uint":    "INT UNSIGNED",
		"int8":    "TINYINT",
		"uint8":   "TINYINT UNSIGNED",
		"int16":   "SMALLINT",
		"uint16":  "SMALLINT UNSIGNED",
		"int32":   "INT",
		"uint32":  "INT UNSIGNED",
		"int64":   "BIGINT",
		"uint64":  "BIGINT UNSIGNED",
		"float32": "FLOAT",
		"float64": "DOUBLE",
		"bool":    "TINYINT(1)",

		// 时间相关
		"time.Time":      "DATETIME",
		"gorm.DeletedAt": "DATETIME",

		// 字节切片
		"[]byte": "BLOB",

		// JSON
		"json.RawMessage": "JSON",
	}

	// 索引后缀
	indexSuffix = map[bool]string{true: "unique", false: "idx"}
	// 索引类型
	indexType = map[bool]string{true: "UNIQUE KEY", false: "KEY"}
)

// Generator 代码生成器
type Generator struct {
	config          *ModuleConfig
	templateManager *TemplateManager
	fileGenerator   *FileGenerator
	sqlGenerator    *SQLGenerator
}

// NewGenerator 创建生成器
func NewGenerator(config *ModuleConfig) *Generator {
	templateManager := NewTemplateManager()
	return &Generator{
		config:          config,
		templateManager: templateManager,
		fileGenerator:   NewFileGenerator(config, templateManager),
		sqlGenerator:    NewSQLGenerator(config),
	}
}

// Generate 生成所有文件
func (g *Generator) Generate() error {
	// 初始化模板
	if err := g.templateManager.InitTemplates(); err != nil {
		fmt.Printf("初始化模板失败: %v\n", err)
		os.Exit(1)
	}

	// 创建基础目录
	baseDir := filepath.Join("./", strings.ToLower(g.config.Module))
	if err := g.fileGenerator.CreateDirs(baseDir); err != nil {
		return err
	}

	// 为每个实体生成文件
	for _, entity := range g.config.Entities {
		if err := g.fileGenerator.GenerateEntityFiles(baseDir, entity); err != nil {
			return fmt.Errorf("生成实体 %s 失败: %v", entity.Name, err)
		}
	}

	// 生成 SQL 文件
	sqlContent := g.sqlGenerator.GenerateSQL()
	sqlFile := filepath.Join(baseDir, dirSQL, g.sqlGenerator.GenerateSQLFileName())
	if err := os.WriteFile(sqlFile, []byte(sqlContent), 0644); err != nil {
		return fmt.Errorf("生成 SQL 文件失败: %v", err)
	}

	return nil
}
