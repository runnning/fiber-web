package generator

// TemplateData 模板渲染数据
type TemplateData struct {
	ModuleName string  // 模块名
	Name       string  // 实体名
	VarName    string  // 变量名(首字母小写)
	TableName  string  // 表名
	Fields     []Field // 字段列表
}

// Index 表示索引定义
type Index struct {
	Name    string   `mapstructure:"name"`    // 索引名称
	Fields  []string `mapstructure:"fields"`  // 索引字段
	Unique  bool     `mapstructure:"unique"`  // 是否是唯一索引
	Comment string   `mapstructure:"comment"` // 索引注释
}

// Field 表示模型字段
type Field struct {
	Name       string `mapstructure:"name"`        // 字段名
	Type       string `mapstructure:"type"`        // 字段类型
	Tag        string `mapstructure:"tag"`         // 字段标签
	Comment    string `mapstructure:"comment"`     // 字段注释
	SqlType    string `mapstructure:"sql_type"`    // SQL类型（可选，默认根据Go类型推导）
	Nullable   bool   `mapstructure:"nullable"`    // 是否可为空
	PrimaryKey bool   `mapstructure:"primary_key"` // 是否为主键
	AutoIncr   bool   `mapstructure:"auto_incr"`   // 是否自增
	Default    string `mapstructure:"default"`     // 默认值
}

// Entity 表示实体定义
type Entity struct {
	Name      string  `mapstructure:"name"`       // 实体名称
	TableName string  `mapstructure:"table_name"` // 表名
	Fields    []Field `mapstructure:"fields"`     // 字段列表
	Indexes   []Index `mapstructure:"indexes"`    // 索引列表
	Comment   string  `mapstructure:"comment"`    // 表注释
}

// SQLConfig SQL生成配置
type SQLConfig struct {
	Filename         string `mapstructure:"filename"`          // SQL文件名（不含扩展名和时间戳）
	IncludeTimestamp bool   `mapstructure:"include_timestamp"` // 是否包含时间戳
	Version          string `mapstructure:"version"`           // 版本号
}

// ModuleConfig 表示模块配置
type ModuleConfig struct {
	Module    string    `mapstructure:"module"`     // 模块名称
	Entities  []Entity  `mapstructure:"entities"`   // 实体列表
	DbEngine  string    `mapstructure:"db_engine"`  // 数据库引擎（默认InnoDB）
	DbCharset string    `mapstructure:"db_charset"` // 数据库字符集（默认utf8mb4）
	SQLConfig SQLConfig `mapstructure:"sql_config"` // SQL生成配置
}
