package generator

// Field 表示模型字段
type Field struct {
	Name    string // 字段名
	Type    string // 字段类型
	Tag     string // 字段标签
	Comment string // 字段注释
}

// Entity 表示实体定义
type Entity struct {
	Name      string  // 实体名称
	TableName string  // 表名
	Fields    []Field // 字段列表
}

// ModuleConfig 表示模块配置
type ModuleConfig struct {
	Module   string   // 模块名称
	Entities []Entity // 实体列表
}
