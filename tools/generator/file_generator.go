package generator

import (
	"fiber_web/pkg/utils/str"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileGenerator 文件生成器
type FileGenerator struct {
	config          *ModuleConfig
	templateManager *TemplateManager
}

// NewFileGenerator 创建文件生成器
func NewFileGenerator(config *ModuleConfig, templateManager *TemplateManager) *FileGenerator {
	return &FileGenerator{
		config:          config,
		templateManager: templateManager,
	}
}

// CreateDirs 创建所需的目录结构
func (fg *FileGenerator) CreateDirs(baseDir string) error {
	dirs := []string{dirEntity, dirRepository, dirUsecase, dirEndpoint, dirSQL}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %v", dir, err)
		}
	}
	return nil
}

// GenerateEntityFiles 生成实体相关的所有文件
func (fg *FileGenerator) GenerateEntityFiles(baseDir string, entity Entity) error {
	// 转换字段名称为驼峰命名
	fields := make([]Field, 0, len(entity.Fields))
	for _, field := range entity.Fields {
		field.Name = str.PascalCase(field.Name)
		fields = append(fields, field)
	}
	data := TemplateData{
		ModuleName: fmt.Sprintf("fiber_web/%s", strings.ToLower(fg.config.Module)),
		Name:       entity.Name,
		VarName:    str.SnakeCase(entity.Name),
		TableName:  entity.TableName,
		Fields:     fields,
	}

	files := []struct {
		tmplName string
		path     string
	}{
		{dirEntity, filepath.Join(baseDir, dirEntity, str.SnakeCase(entity.Name)+".tpl")},
		{dirRepository, filepath.Join(baseDir, dirRepository, str.SnakeCase(entity.Name)+"_repository.tpl")},
		{dirUsecase, filepath.Join(baseDir, dirUsecase, str.SnakeCase(entity.Name)+"_usecase.tpl")},
		{dirEndpoint, filepath.Join(baseDir, dirEndpoint, str.SnakeCase(entity.Name)+"_endpoint.tpl")},
	}

	for _, file := range files {
		if err := fg.generateFile(file.tmplName, file.path, data); err != nil {
			return fmt.Errorf("生成文件 %s 失败: %v", file.path, err)
		}
	}

	return nil
}

// generateFile 生成单个文件
func (fg *FileGenerator) generateFile(tmplName, outputPath string, data interface{}) error {
	tmpl, ok := fg.templateManager.GetTemplate(tmplName)
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
