package main

import (
	"fiber_web/tools/generator"
	"flag"
	"fmt"
	"os"
)

func main() {
	entityName := flag.String("name", "", "实体名称")
	module := flag.String("module", "core", "模块名称 (例如: core, user, product)")
	flag.Parse()

	if *entityName == "" {
		fmt.Println("请提供实体名称，例如: -name Product -module Core")
		os.Exit(1)
	}

	gen := generator.NewGenerator(*entityName, *module)
	if err := gen.Generate(); err != nil {
		fmt.Printf("生成失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功在模块 %s 中生成 %s 的所有文件\n", *module, *entityName)
}
