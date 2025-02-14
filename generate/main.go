package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	// 配置和模板文件路径
	configFile = "config.json"
)

// 定义配置结构体，字段名必须与 JSON 匹配
type Config struct {
	Templates string     `json:"templates"`
	OutputDir string     `json:"outputDir"`
	Conf      []confData `json:"conf"`
}

// 渲染文件名及渲染信息any(模版中需要渲染的字段名需一致.类型会解析为map)
type confData struct {
	TplData  any    `json:"tplData"`
	FileName string `json:"fileName"`
}

// go run main.go
// make local
// chmod +x binaryMain
func main() {
	// 1. 读取 JSON 配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
		return
	}

	// 2. 使用 ParseGlob 解析整个目录的模板文件
	tmpl, err := template.ParseGlob(config.Templates)
	if err != nil {
		log.Fatalf("读取模板文件失败: %v", err)
		return
	}

	// 3. 渲染模板并写入文件
	for _, value := range config.Conf {
		if err := outputToFile(tmpl, value, config.OutputDir); err != nil {
			log.Fatalf("渲染模板失败: %v", err)
			return
		}
	}
}

// 渲染输出文件
func outputToFile(tmpl *template.Template, value confData, outputDir string) error {
	var (
		// 构建文件名
		tplName    = strings.Join([]string{value.FileName, `tpl`}, `.`)
		targetFile = strings.Join([]string{value.FileName, `go`}, `.`)
	)

	// 确保目录存在（如果不存在则创建）
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 在目录下:创建新文件,如果文件已存在,会清空原文件（注意数据覆盖）
	outFile, err := os.Create(filepath.Join(outputDir, targetFile))
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}

	defer func() {
		_ = outFile.Close()
	}()

	// fmt.Println(value.TplData) // 注意json配置中的模版字段大小写要一致.这个值解析的类型为map[string]interface{}

	if err = tmpl.ExecuteTemplate(outFile, tplName, value.TplData); err != nil {
		return fmt.Errorf("模板渲染失败: %w", err)
	}

	log.Println(targetFile, "渲染处理完成,结果已写入")

	return nil
}

// 读取 JSON 配置文件
func loadConfig() (Config, error) {
	var config Config

	file, err := os.Open(configFile)
	if err != nil {
		return config, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)

	return config, err
}
