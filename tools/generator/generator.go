package generator

import (
	"fiber_web/tools/generator/templates"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// Field 表示模型字段
//type Field struct {
//	Name    string // 字段名
//	Type    string // 字段类型
//	Tag     string // 字段标签
//	Comment string // 字段注释
//}

// TemplateData 模板数据
type TemplateData struct {
	ModuleName string  // 模块名
	Name       string  // 实体名
	VarName    string  // 变量名(首字母小写)
	TableName  string  // 表名
	Fields     []Field // 字段列表
}

type Generator struct {
	config    *ModuleConfig
	templates map[string]*template.Template
	once      sync.Once
}

// NewGenerator 创建生成器
func NewGenerator(config *ModuleConfig) *Generator {
	return &Generator{
		config:    config,
		templates: make(map[string]*template.Template),
	}
}

// initTemplates 初始化并预编译所有模板
func (g *Generator) initTemplates() error {
	templateMap := map[string]string{
		"entity":     templates.EntityTemplate,
		"repository": templates.RepositoryTemplate,
		"usecase":    templates.UseCaseTemplate,
		"endpoint":   templates.EndpointTemplate,
	}

	for name, content := range templateMap {
		tmpl, err := template.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("解析模板 %s 失败: %v", name, err)
		}
		g.templates[name] = tmpl
	}
	return nil
}

func (g *Generator) Generate() error {
	// 初始化模板（只执行一次）
	g.once.Do(func() {
		if err := g.initTemplates(); err != nil {
			fmt.Printf("初始化模板失败: %v\n", err)
			os.Exit(1)
		}
	})

	// 获取基础目录
	baseDir := filepath.Join("./", strings.ToLower(g.config.Module))

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

	// 为每个实体生成文件
	for _, entity := range g.config.Entities {
		if err := g.generateEntity(baseDir, entity); err != nil {
			return fmt.Errorf("生成实体 %s 失败: %v", entity.Name, err)
		}
	}

	return nil
}

func (g *Generator) generateEntity(baseDir string, entity Entity) error {
	// 准备模板数据
	data := TemplateData{
		ModuleName: fmt.Sprintf("fiber_web/%s", strings.ToLower(g.config.Module)),
		Name:       entity.Name,
		VarName:    strings.ToLower(entity.Name[:1]) + entity.Name[1:],
		TableName:  entity.TableName,
		Fields:     entity.Fields,
	}

	// 生成各个文件
	files := []struct {
		tmplName string
		path     string
	}{
		{"entity", filepath.Join(baseDir, "entity", strings.ToLower(entity.Name)+".tpl")},
		{"repository", filepath.Join(baseDir, "repository", strings.ToLower(entity.Name)+"_repository.tpl")},
		{"usecase", filepath.Join(baseDir, "usecase", strings.ToLower(entity.Name)+"_usecase.tpl")},
		{"endpoint", filepath.Join(baseDir, "endpoint", strings.ToLower(entity.Name)+"_endpoint.tpl")},
	}

	for _, file := range files {
		if err := g.generateFile(file.tmplName, file.path, data); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateFile(tmplName, outputPath string, data TemplateData) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl, ok := g.templates[tmplName]
	if !ok {
		return fmt.Errorf("模板 %s 未找到", tmplName)
	}

	return tmpl.Execute(file, data)
}
