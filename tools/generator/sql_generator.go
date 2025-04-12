package generator

import (
	"fiber_web/pkg/utils/str"
	"fmt"
	"strings"
	"time"
)

// SQLGenerator SQL生成器
type SQLGenerator struct {
	config *ModuleConfig
}

// NewSQLGenerator 创建SQL生成器
func NewSQLGenerator(config *ModuleConfig) *SQLGenerator {
	return &SQLGenerator{
		config: config,
	}
}

// GenerateSQL 生成SQL内容
func (g *SQLGenerator) GenerateSQL() string {
	var b strings.Builder

	// 添加头部注释
	g.writeHeaderComment(&b)

	// 生成每个实体的建表语句
	for _, entity := range g.config.Entities {
		if entity.Comment != "" {
			b.WriteString(fmt.Sprintf("-- %s\n", entity.Comment))
		}
		b.WriteString(g.generateCreateTableSQL(entity))
		b.WriteString("\n")
	}

	return b.String()
}

func (g *SQLGenerator) writeHeaderComment(b *strings.Builder) {
	b.WriteString(fmt.Sprintf("-- 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("-- 模块: %s\n\n", g.config.Module))
}

func (g *SQLGenerator) generateCreateTableSQL(entity Entity) string {
	var b strings.Builder

	// 写入建表头部
	g.writeTableHeader(&b, entity)

	// 写入列定义
	columns, primaryKeys := g.generateColumns(entity.Fields)
	b.WriteString(strings.Join(columns, ",\n"))

	// 写入主键
	g.writePrimaryKey(&b, primaryKeys)

	// 写入索引
	g.writeIndexes(&b, entity)

	// 写入表选项
	g.writeTableOptions(&b, entity)

	return b.String()
}

func (g *SQLGenerator) writeTableHeader(b *strings.Builder, entity Entity) {
	b.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", entity.TableName))
}

func (g *SQLGenerator) generateColumns(fields []Field) ([]string, []string) {
	columns := make([]string, 0, len(fields))
	var primaryKeys []string

	for _, field := range fields {
		columns = append(columns, g.generateColumnDef(field))
		if field.PrimaryKey {
			primaryKeys = append(primaryKeys, field.Name)
		}
	}
	return columns, primaryKeys
}

func (g *SQLGenerator) writePrimaryKey(b *strings.Builder, primaryKeys []string) {
	if len(primaryKeys) > 0 {
		b.WriteString(fmt.Sprintf(",\n  PRIMARY KEY (%s)", strings.Join(primaryKeys, ",")))
	}
}

func (g *SQLGenerator) writeIndexes(b *strings.Builder, entity Entity) {
	for _, idx := range entity.Indexes {
		indexName := fmt.Sprintf("%s_%s_%s", entity.TableName, idx.Name, indexSuffix[idx.Unique])

		b.WriteString(fmt.Sprintf(",\n  %s `%s` (%s)",
			indexType[idx.Unique],
			indexName,
			strings.Join(idx.Fields, ","),
		))

		if idx.Comment != "" {
			b.WriteString(fmt.Sprintf(" COMMENT '%s'", idx.Comment))
		}
	}
}

func (g *SQLGenerator) writeTableOptions(b *strings.Builder, entity Entity) {
	engine := g.config.DbEngine
	if engine == "" {
		engine = "InnoDB"
	}
	charset := g.config.DbCharset
	if charset == "" {
		charset = "utf8mb4"
	}

	b.WriteString(fmt.Sprintf("\n) ENGINE=%s DEFAULT CHARSET=%s", engine, charset))

	if entity.Comment != "" {
		b.WriteString(fmt.Sprintf(" COMMENT='%s'", entity.Comment))
	}

	b.WriteString(";\n")
}

func (g *SQLGenerator) generateColumnDef(field Field) string {
	fieldName := field.Name
	if len(fieldName) > 2 {
		fieldName = str.SnakeCase(fieldName)
	} else {
		fieldName = strings.ToLower(fieldName)
	}
	parts := []string{
		fmt.Sprintf("  %s %s", fieldName, g.getColumnType(field)),
	}

	parts = append(parts, g.getColumnConstraints(field)...)

	if field.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", field.Comment))
	}

	return strings.Join(parts, " ")
}

func (g *SQLGenerator) getColumnType(field Field) string {
	if field.SqlType != "" {
		return field.SqlType
	}
	if sqlType, ok := typeMap[field.Type]; ok {
		return sqlType
	}
	return "VARCHAR(255)"
}

func (g *SQLGenerator) getColumnConstraints(field Field) []string {
	var constraints []string

	if !field.Nullable {
		constraints = append(constraints, "NOT NULL")
	}

	if field.AutoIncr {
		constraints = append(constraints, "AUTO_INCREMENT")
	}

	if field.Default != "" {
		constraints = append(constraints, fmt.Sprintf("DEFAULT %s", field.Default))
	}

	return constraints
}

func (g *SQLGenerator) GenerateSQLFileName() string {
	parts := []string{
		g.config.SQLConfig.Filename,
		g.config.SQLConfig.Version,
	}

	if parts[0] == "" {
		parts[0] = fmt.Sprintf("create_%s_tables", strings.ToLower(g.config.Module))
	}

	if g.config.SQLConfig.IncludeTimestamp {
		parts = append(parts[:2], time.Now().Format("20060102_150405"))
	}

	return strings.Join(parts, "_") + ".sql"
}
