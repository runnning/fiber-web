package generator

// Field 表示模型字段
type Field struct {
	Name    string `yaml:"name"`    // 字段名
	Type    string `yaml:"type"`    // 字段类型
	Tag     string `yaml:"tag"`     // 字段标签
	Comment string `yaml:"comment"` // 字段注释
}

// Entity 表示实体定义
type Entity struct {
	Name      string  `yaml:"name"`       // 实体名称
	TableName string  `yaml:"table_name"` // 表名
	Fields    []Field `yaml:"fields"`     // 字段列表
}

// ModuleConfig 表示模块配置
type ModuleConfig struct {
	Module   string   `yaml:"module"`   // 模块名称
	Entities []Entity `yaml:"entities"` // 实体列表
}
