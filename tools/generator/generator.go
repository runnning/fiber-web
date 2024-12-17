package generator

import (
	"fiber_web/tools/generator/templates"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateData 模板数据
type TemplateData struct {
	ModuleName string // 模块名
	Name       string // 实体名
	VarName    string // 变量名(首字母小写)
}

type Generator struct {
	data TemplateData
}

// NewGenerator 创建生成器
func NewGenerator(name, module string) *Generator {
	varName := strings.ToLower(name[:1]) + name[1:]
	return &Generator{
		data: TemplateData{
			ModuleName: module,
			Name:       name,
			VarName:    varName,
		},
	}
}

func (g *Generator) Generate() error {
	// 获取基础目录
	baseDir := filepath.Join("./", strings.ToLower(g.data.ModuleName))

	// 创建模块目录结构
	dirs := []string{
		filepath.Join(baseDir, "entity"),
		filepath.Join(baseDir, "repository"),
		filepath.Join(baseDir, "usecase"),
		filepath.Join(baseDir, "endpoint"),
	}

	// 创建必要的目录
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %v", dir, err)
		}
	}

	// 生成各个文件
	files := []struct {
		template string
		path     string
	}{
		{templates.EntityTemplate, filepath.Join(baseDir, "entity", strings.ToLower(g.data.Name)+".tpl")},
		{templates.RepositoryTemplate, filepath.Join(baseDir, "repository", strings.ToLower(g.data.Name)+"_repository.tpl")},
		{templates.UseCaseTemplate, filepath.Join(baseDir, "usecase", strings.ToLower(g.data.Name)+"_usecase.tpl")},
		{templates.EndpointTemplate, filepath.Join(baseDir, "endpoint", strings.ToLower(g.data.Name)+"_endpoint.tpl")},
	}

	for _, file := range files {
		if err := g.generateFile(file.template, file.path); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateFile(templateContent string, outputPath string) error {
	tmpl, err := template.New("file").Parse(templateContent)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 更新 ModuleName 以包含完整的导入路径
	g.data.ModuleName = fmt.Sprintf("fiber_web/%s", strings.ToLower(g.data.ModuleName))

	// 直接使用 TemplateData
	return tmpl.Execute(file, g.data)
}
