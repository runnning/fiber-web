package generator

import (
	"fmt"
	"strings"
	"time"
)

// SQL类型映射
var goToSQLTypeMap = map[string]string{
	"string":         "VARCHAR(255)",
	"int":            "INT",
	"int32":          "INT",
	"int8":           "TINYINT",
	"int16":          "SMALLINT",
	"int64":          "BIGINT",
	"uint":           "INT UNSIGNED",
	"uint32":         "INT UNSIGNED",
	"uint8":          "TINYINT UNSIGNED",
	"uint16":         "SMALLINT UNSIGNED",
	"uint64":         "BIGINT UNSIGNED",
	"float32":        "FLOAT",
	"float64":        "DOUBLE",
	"bool":           "TINYINT(1)",
	"time.Time":      "DATETIME",
	"gorm.DeletedAt": "DATETIME",
}

// goTypeToSQLType 将 Go 类型转换为 SQL 类型
func goTypeToSQLType(goType string) string {
	if sqlType, ok := goToSQLTypeMap[goType]; ok {
		return sqlType
	}
	return "VARCHAR(255)"
}

// parseGormTag 解析 gorm 标签
func parseGormTag(tag string) map[string]string {
	result := make(map[string]string)
	if !strings.Contains(tag, "gorm") {
		return result
	}

	// 提取 gorm 标签内容
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

// generateColumnDef 生成列定义
func generateColumnDef(field Field, gormTags map[string]string) string {
	var parts []string

	// SQL类型
	sqlType := field.SqlType
	if sqlType == "" {
		sqlType = goTypeToSQLType(field.Type)
	}
	parts = append(parts, fmt.Sprintf("  %s %s", field.Name, sqlType))

	// 可空性
	if !field.Nullable && !strings.Contains(strings.ToUpper(sqlType), "DATETIME") {
		parts = append(parts, "NOT NULL")
	}

	// 默认值
	if defaultVal, ok := gormTags["default"]; ok {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", defaultVal))
	}

	// 自增
	if _, ok := gormTags["primarykey"]; ok {
		parts = append(parts, "AUTO_INCREMENT")
	}

	// 注释
	if field.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", field.Comment))
	}

	return strings.Join(parts, " ")
}

// generateCreateTableSQL 生成建表 SQL
func generateCreateTableSQL(entity Entity, config ModuleConfig) string {
	var builder strings.Builder

	// 设置默认值
	engine := config.DbEngine
	if engine == "" {
		engine = "InnoDB"
	}
	charset := config.DbCharset
	if charset == "" {
		charset = "utf8mb4"
	}

	// 开始建表语句
	builder.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", entity.TableName))

	// 收集索引信息
	var (
		columns     []string
		primaryKeys []string
		uniqueKeys  []string
		indexes     []string
	)

	// 处理字段
	for _, field := range entity.Fields {
		gormTags := parseGormTag(field.Tag)

		// 生成列定义
		columns = append(columns, generateColumnDef(field, gormTags))

		// 收集索引信息
		if _, ok := gormTags["primarykey"]; ok {
			primaryKeys = append(primaryKeys, field.Name)
		}
		if _, ok := gormTags["unique"]; ok {
			uniqueKeys = append(uniqueKeys, field.Name)
		}
		if _, ok := gormTags["index"]; ok {
			indexes = append(indexes, field.Name)
		}
	}

	// 写入列定义
	builder.WriteString(strings.Join(columns, ",\n"))

	// 添加主键
	if len(primaryKeys) > 0 {
		builder.WriteString(fmt.Sprintf(",\n  PRIMARY KEY (%s)", strings.Join(primaryKeys, ",")))
	}

	// 添加唯一索引
	for _, key := range uniqueKeys {
		builder.WriteString(fmt.Sprintf(",\n  UNIQUE KEY %s_%s_unique (%s)",
			entity.TableName, strings.ToLower(key), key))
	}

	// 添加普通索引
	for _, key := range indexes {
		builder.WriteString(fmt.Sprintf(",\n  KEY %s_%s_index (%s)",
			entity.TableName, strings.ToLower(key), key))
	}

	// 添加表选项
	builder.WriteString(fmt.Sprintf("\n) ENGINE=%s DEFAULT CHARSET=%s", engine, charset))

	// 添加表注释
	if entity.Comment != "" {
		builder.WriteString(fmt.Sprintf(" COMMENT='%s'", entity.Comment))
	}

	builder.WriteString(";\n")
	return builder.String()
}

// GenerateSQL 为整个模块生成 SQL
func (g *Generator) GenerateSQL() string {
	var builder strings.Builder

	// 添加头部注释
	builder.WriteString(fmt.Sprintf("-- 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("-- 模块: %s\n\n", g.config.Module))

	// 生成每个实体的建表语句
	for _, entity := range g.config.Entities {
		if entity.Comment != "" {
			builder.WriteString(fmt.Sprintf("-- %s\n", entity.Comment))
		}
		builder.WriteString(generateCreateTableSQL(entity, *g.config))
		builder.WriteString("\n")
	}

	return builder.String()
}
