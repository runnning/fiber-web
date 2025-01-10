package generator

// Field 表示模型字段
type Field struct {
	Name     string `yaml:"name"`     // 字段名
	Type     string `yaml:"type"`     // 字段类型
	Tag      string `yaml:"tag"`      // 字段标签
	Comment  string `yaml:"comment"`  // 字段注释
	SqlType  string `yaml:"sql_type"` // SQL类型（可选，默认根据Go类型推导）
	Nullable bool   `yaml:"nullable"` // 是否可为空
}

// Entity 表示实体定义
type Entity struct {
	Name      string  `yaml:"name"`       // 实体名称
	TableName string  `yaml:"table_name"` // 表名
	Fields    []Field `yaml:"fields"`     // 字段列表
	Comment   string  `yaml:"comment"`    // 表注释
}

// SQLConfig SQL生成配置
type SQLConfig struct {
	Filename         string `yaml:"filename"`          // SQL文件名（不含扩展名和时间戳）
	IncludeTimestamp bool   `yaml:"include_timestamp"` // 是否包含时间戳
	Version          string `yaml:"version"`           // 版本号
}

// ModuleConfig 表示模块配置
type ModuleConfig struct {
	Module    string    `yaml:"module"`     // 模块名称
	Entities  []Entity  `yaml:"entities"`   // 实体列表
	DbEngine  string    `yaml:"db_engine"`  // 数据库引擎（默认InnoDB）
	DbCharset string    `yaml:"db_charset"` // 数据库字符集（默认utf8mb4）
	SQLConfig SQLConfig `yaml:"sql_config"` // SQL生成配置
}
