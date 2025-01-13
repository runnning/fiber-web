package generator

import (
	"fiber_web/tools/generator/templates"
	"fmt"
	"text/template"
)

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]*template.Template
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*template.Template),
	}
}

// InitTemplates 初始化所有模板
func (tm *TemplateManager) InitTemplates() error {
	templateStrings := map[string]string{
		dirEntity:     templates.EntityTemplate,
		dirRepository: templates.RepositoryTemplate,
		dirUsecase:    templates.UseCaseTemplate,
		dirEndpoint:   templates.EndpointTemplate,
	}

	for name, content := range templateStrings {
		tmpl, err := template.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("解析模板 %s 失败: %v", name, err)
		}
		tm.templates[name] = tmpl
	}
	return nil
}

// GetTemplate 获取指定名称的模板
func (tm *TemplateManager) GetTemplate(name string) (*template.Template, bool) {
	tmpl, ok := tm.templates[name]
	return tmpl, ok
}
