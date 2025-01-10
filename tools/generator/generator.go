package generator

import (
	"fiber_web/tools/generator/templates"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

// 定义常量
const (
	dirEntity     = "entity"
	dirRepository = "repository"
	dirUsecase    = "usecase"
	dirEndpoint   = "endpoint"
	dirSQL        = "sql"
)

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
		dirEntity:     templates.EntityTemplate,
		dirRepository: templates.RepositoryTemplate,
		dirUsecase:    templates.UseCaseTemplate,
		dirEndpoint:   templates.EndpointTemplate,
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

// createDirs 创建必要的目录
func (g *Generator) createDirs(baseDir string) error {
	dirs := []string{dirEntity, dirRepository, dirUsecase, dirEndpoint, dirSQL}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %v", dir, err)
		}
	}
	return nil
}

// generateSQLFileName 生成SQL文件名
func (g *Generator) generateSQLFileName() string {
	var parts []string

	// 基础文件名
	if g.config.SQLConfig.Filename != "" {
		parts = append(parts, g.config.SQLConfig.Filename)
	} else {
		parts = append(parts, fmt.Sprintf("create_%s_tables", strings.ToLower(g.config.Module)))
	}

	// 版本号
	if g.config.SQLConfig.Version != "" {
		parts = append(parts, g.config.SQLConfig.Version)
	}

	// 时间戳
	if g.config.SQLConfig.IncludeTimestamp {
		parts = append(parts, time.Now().Format("20060102_150405"))
	}

	return strings.Join(parts, "_") + ".sql"
}

func (g *Generator) Generate() error {
	// 初始化模板（只执行一次）
	g.once.Do(func() {
		if err := g.initTemplates(); err != nil {
			fmt.Printf("初始化模板失败: %v\n", err)
			os.Exit(1)
		}
	})

	// 获取基础目录并创建目录结构
	baseDir := filepath.Join("./", strings.ToLower(g.config.Module))
	if err := g.createDirs(baseDir); err != nil {
		return err
	}

	// 为每个实体生成文件
	for _, entity := range g.config.Entities {
		if err := g.generateEntity(baseDir, entity); err != nil {
			return fmt.Errorf("生成实体 %s 失败: %v", entity.Name, err)
		}
	}

	// 生成 SQL 文件
	sqlContent := g.GenerateSQL()
	sqlFile := filepath.Join(baseDir, dirSQL, g.generateSQLFileName())
	if err := os.WriteFile(sqlFile, []byte(sqlContent), 0644); err != nil {
		return fmt.Errorf("生成 SQL 文件失败: %v", err)
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

	// 定义要生成的文件
	files := []struct {
		tmplName string
		path     string
	}{
		{dirEntity, filepath.Join(baseDir, dirEntity, strings.ToLower(entity.Name)+".tpl")},
		{dirRepository, filepath.Join(baseDir, dirRepository, strings.ToLower(entity.Name)+"_repository.tpl")},
		{dirUsecase, filepath.Join(baseDir, dirUsecase, strings.ToLower(entity.Name)+"_usecase.tpl")},
		{dirEndpoint, filepath.Join(baseDir, dirEndpoint, strings.ToLower(entity.Name)+"_endpoint.tpl")},
	}

	// 生成文件
	for _, file := range files {
		if err := g.generateFile(file.tmplName, file.path, data); err != nil {
			return fmt.Errorf("生成文件 %s 失败: %v", file.path, err)
		}
	}

	return nil
}

func (g *Generator) generateFile(tmplName, outputPath string, data TemplateData) error {
	tmpl, ok := g.templates[tmplName]
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
