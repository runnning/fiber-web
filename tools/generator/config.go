package generator

import (
	"fmt"

	"github.com/spf13/viper"
)

var Data = new(ModuleConfig)

// LoadConfig 从配置文件加载配置
func LoadConfig(configFile string) error {
	v := viper.New()
	v.SetConfigFile(configFile)

	// 设置默认值
	v.SetDefault("db_engine", "InnoDB")
	v.SetDefault("db_charset", "utf8mb4")
	v.SetDefault("sql_config.include_timestamp", true)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := v.Unmarshal(Data); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	// 验证配置
	if err := ValidateConfig(Data); err != nil {
		return err
	}

	return nil
}

// ValidateConfig 验证配置
func ValidateConfig(config *ModuleConfig) error {
	if config.Module == "" {
		return fmt.Errorf("module 不能为空")
	}
	if len(config.Entities) == 0 {
		return fmt.Errorf("至少需要一个实体定义")
	}

	// 验证每个实体
	for _, entity := range config.Entities {
		if entity.Name == "" {
			return fmt.Errorf("实体名称不能为空")
		}
		if entity.TableName == "" {
			return fmt.Errorf("实体 %s 的表名不能为空", entity.Name)
		}
		if len(entity.Fields) == 0 {
			return fmt.Errorf("实体 %s 至少需要一个字段", entity.Name)
		}

		// 验证字段
		for _, field := range entity.Fields {
			if field.Name == "" {
				return fmt.Errorf("实体 %s 的字段名不能为空", entity.Name)
			}
			if field.Type == "" {
				return fmt.Errorf("实体 %s 的字段 %s 类型不能为空", entity.Name, field.Name)
			}
		}
	}

	return nil
}

// ErrInvalidConfig 表示配置验证错误
type ErrInvalidConfig struct {
	Message string
}

func (e ErrInvalidConfig) Error() string {
	return e.Message
}
