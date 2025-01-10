package generator

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig 从YAML文件加载配置
func LoadConfig(filename string) (*ModuleConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config ModuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ValidateConfig 验证配置
func ValidateConfig(config *ModuleConfig) error {
	if config.Module == "" {
		return ErrInvalidConfig{Message: "module name is required"}
	}
	if len(config.Entities) == 0 {
		return ErrInvalidConfig{Message: "at least one entity is required"}
	}
	for _, entity := range config.Entities {
		if entity.Name == "" {
			return ErrInvalidConfig{Message: "entity name is required"}
		}
		if len(entity.Fields) == 0 {
			return ErrInvalidConfig{Message: "at least one field is required for entity " + entity.Name}
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
