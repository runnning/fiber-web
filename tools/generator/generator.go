package generator

import (
	"fiber_web/tools/generator/templates"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// 定义常量
const (
	dirEntity     = "entity"
	dirRepository = "repository"
	dirUsecase    = "usecase"
	dirEndpoint   = "endpoint"
	dirSQL        = "sql"
)

// SQL类型映射
var typeMap = map[string]string{
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

// TemplateData 模板数据
type TemplateData struct {
	ModuleName string  // 模块名
	Name       string  // 实体名
	VarName    string  // 变量名(首字母小写)
	TableName  string  // 表名
	Fields     []Field // 字段列表
}

type Generator struct {
	config    *ModuleConfig
	templates map[string]*template.Template
	//once      sync.Once
}

// NewGenerator 创建生成器
func NewGenerator(config *ModuleConfig) *Generator {
	return &Generator{
		config:    config,
		templates: make(map[string]*template.Template),
	}
}

// Generate 生成所有文件
func (g *Generator) Generate() error {
	// 初始化模板（只执行一次）
	//g.once.Do(func() {
	if err := g.initTemplates(); err != nil {
		fmt.Printf("初始化模板失败: %v\n", err)
		os.Exit(1)
	}
	//})

	// 获取基础目录并创建目录结构
	baseDir := filepath.Join("./", strings.ToLower(g.config.Module))
	if err := g.createDirs(baseDir); err != nil {
		return err
	}

	// 为每个实体生成文件
	for _, entity := range g.config.Entities {
		if err := g.generateEntity(baseDir, entity); err != nil {
			return fmt.Errorf("生成实体 %s 失败: %v", entity.Name, err)
		}
	}

	// 生成 SQL 文件
	sqlContent := g.generateSQL()
	sqlFile := filepath.Join(baseDir, dirSQL, g.generateSQLFileName())
	if err := os.WriteFile(sqlFile, []byte(sqlContent), 0644); err != nil {
		return fmt.Errorf("生成 SQL 文件失败: %v", err)
	}

	return nil
}

// 内部辅助方法

func (g *Generator) initTemplates() error {
	templates := map[string]string{
		dirEntity:     templates.EntityTemplate,
		dirRepository: templates.RepositoryTemplate,
		dirUsecase:    templates.UseCaseTemplate,
		dirEndpoint:   templates.EndpointTemplate,
	}

	for name, content := range templates {
		tmpl, err := template.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("解析模板 %s 失败: %v", name, err)
		}
		g.templates[name] = tmpl
	}
	return nil
}

func (g *Generator) createDirs(baseDir string) error {
	dirs := []string{dirEntity, dirRepository, dirUsecase, dirEndpoint, dirSQL}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %v", dir, err)
		}
	}
	return nil
}

func (g *Generator) generateEntity(baseDir string, entity Entity) error {
	data := TemplateData{
		ModuleName: fmt.Sprintf("fiber_web/%s", strings.ToLower(g.config.Module)),
		Name:       entity.Name,
		VarName:    strings.ToLower(entity.Name[:1]) + entity.Name[1:],
		TableName:  entity.TableName,
		Fields:     entity.Fields,
	}

	files := []struct {
		tmplName string
		path     string
	}{
		{dirEntity, filepath.Join(baseDir, dirEntity, strings.ToLower(entity.Name)+".tpl")},
		{dirRepository, filepath.Join(baseDir, dirRepository, strings.ToLower(entity.Name)+"_repository.tpl")},
		{dirUsecase, filepath.Join(baseDir, dirUsecase, strings.ToLower(entity.Name)+"_usecase.tpl")},
		{dirEndpoint, filepath.Join(baseDir, dirEndpoint, strings.ToLower(entity.Name)+"_endpoint.tpl")},
	}

	for _, file := range files {
		if err := g.generateFile(file.tmplName, file.path, data); err != nil {
			return fmt.Errorf("生成文件 %s 失败: %v", file.path, err)
		}
	}

	return nil
}

func (g *Generator) generateFile(tmplName, outputPath string, data interface{}) error {
	tmpl, ok := g.templates[tmplName]
	if !ok {
		return fmt.Errorf("模板 %s 未找到", tmplName)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败 %s: %v", outputPath, err)
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// SQL 生成相关方法

func (g *Generator) generateSQLFileName() string {
	var parts []string

	if g.config.SQLConfig.Filename != "" {
		parts = append(parts, g.config.SQLConfig.Filename)
	} else {
		parts = append(parts, fmt.Sprintf("create_%s_tables", strings.ToLower(g.config.Module)))
	}

	if g.config.SQLConfig.Version != "" {
		parts = append(parts, g.config.SQLConfig.Version)
	}

	if g.config.SQLConfig.IncludeTimestamp {
		parts = append(parts, time.Now().Format("20060102_150405"))
	}

	return strings.Join(parts, "_") + ".sql"
}

func (g *Generator) generateSQL() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("-- 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("-- 模块: %s\n\n", g.config.Module))

	for _, entity := range g.config.Entities {
		if entity.Comment != "" {
			b.WriteString(fmt.Sprintf("-- %s\n", entity.Comment))
		}
		b.WriteString(g.generateCreateTableSQL(entity))
		b.WriteString("\n")
	}

	return b.String()
}

func (g *Generator) generateCreateTableSQL(entity Entity) string {
	var b strings.Builder

	engine := g.config.DbEngine
	if engine == "" {
		engine = "InnoDB"
	}
	charset := g.config.DbCharset
	if charset == "" {
		charset = "utf8mb4"
	}

	b.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", entity.TableName))

	var columns []string
	var primaryKeys []string

	// 用于收集索引信息
	indexMap := make(map[string][]string)       // 普通索引
	uniqueIndexMap := make(map[string][]string) // 唯一索引

	for _, field := range entity.Fields {
		columns = append(columns, g.generateColumnDef(field))

		// 处理 gorm 标签
		if tags := parseGormTag(field.Tag); len(tags) > 0 {
			if _, ok := tags["primarykey"]; ok {
				primaryKeys = append(primaryKeys, field.Name)
			}
		}

		// 处理普通索引
		for _, idxName := range field.Index {
			indexMap[idxName] = append(indexMap[idxName], field.Name)
		}

		// 处理唯一索引
		for _, idxName := range field.Unique {
			uniqueIndexMap[idxName] = append(uniqueIndexMap[idxName], field.Name)
		}
	}

	b.WriteString(strings.Join(columns, ",\n"))

	// 添加主键
	if len(primaryKeys) > 0 {
		b.WriteString(fmt.Sprintf(",\n  PRIMARY KEY (%s)", strings.Join(primaryKeys, ",")))
	}

	// 添加唯一索引
	for idxName, fields := range uniqueIndexMap {
		indexName := fmt.Sprintf("%s_%s_unique", entity.TableName, idxName)
		b.WriteString(fmt.Sprintf(",\n  UNIQUE KEY `%s` (%s)", indexName, strings.Join(fields, ",")))
	}

	// 添加普通索引
	for idxName, fields := range indexMap {
		indexName := fmt.Sprintf("%s_%s_idx", entity.TableName, idxName)
		b.WriteString(fmt.Sprintf(",\n  KEY `%s` (%s)", indexName, strings.Join(fields, ",")))
	}

	b.WriteString(fmt.Sprintf("\n) ENGINE=%s DEFAULT CHARSET=%s", engine, charset))

	if entity.Comment != "" {
		b.WriteString(fmt.Sprintf(" COMMENT='%s'", entity.Comment))
	}

	b.WriteString(";\n")
	return b.String()
}

func (g *Generator) generateColumnDef(field Field) string {
	var parts []string

	sqlType := field.SqlType
	if sqlType == "" {
		sqlType = typeMap[field.Type]
		if sqlType == "" {
			sqlType = "VARCHAR(255)"
		}
	}
	parts = append(parts, fmt.Sprintf("  %s %s", field.Name, sqlType))

	if !field.Nullable && !strings.Contains(strings.ToUpper(sqlType), "DATETIME") {
		parts = append(parts, "NOT NULL")
	}

	if tags := parseGormTag(field.Tag); len(tags) > 0 {
		if val, ok := tags["default"]; ok {
			parts = append(parts, fmt.Sprintf("DEFAULT %s", val))
		}
		if _, ok := tags["primarykey"]; ok {
			parts = append(parts, "AUTO_INCREMENT")
		}
	}

	if field.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", field.Comment))
	}

	return strings.Join(parts, " ")
}

func parseGormTag(tag string) map[string]string {
	result := make(map[string]string)
	if !strings.Contains(tag, "gorm") {
		return result
	}

	gormTag := strings.Split(strings.Split(tag, "gorm:\"")[1], "\"")[0]
	for _, pair := range strings.Split(gormTag, ";") {
		if pair == "" {
			continue
		}
		if i := strings.IndexByte(pair, ':'); i > 0 {
			result[pair[:i]] = pair[i+1:]
		} else {
			result[pair] = "true"
		}
	}

	return result
}
