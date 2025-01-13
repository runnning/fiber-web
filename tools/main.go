package main

import (
	"fiber_web/tools/generator"
	"flag"
	"fmt"
	"os"
)

func main() {
	configFile := flag.String("config", "", "配置文件路径")
	flag.Parse()

	if *configFile == "" {
		fmt.Println("请提供配置文件路径，例如: -config configs/admin/user.yaml")
		os.Exit(1)
	}

	// 加载配置
	err := generator.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	// 创建生成器并生成代码
	gen := generator.NewGenerator(generator.Data)
	if err := gen.Generate(); err != nil {
		fmt.Printf("生成失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功在模块 %s 中生成所有文件\n", generator.Data.Module)
}
