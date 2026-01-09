// scripts/get_config.go - 通用配置读取工具，支持读取ini文件的指定section和key
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getConfig 读取ini配置文件的指定section和key
func getConfig(section, key string) string {
	// 配置文件查找路径（优先config/config.ini）
	configPaths := []string{
		"config.ini",
		filepath.Join("src", "config.ini"),
		filepath.Join("config", "config.ini"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			file, err := os.Open(path)
			if err != nil {
				continue
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			currentSection := ""
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				// 跳过空行和注释
				if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
					continue
				}
				// 匹配section（如 [app]）
				if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
					currentSection = strings.Trim(line[1:len(line)-1], " ")
					continue
				}
				// 匹配指定section下的key
				if currentSection == section {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
						// 移除值两侧的引号和空格
						return strings.Trim(strings.TrimSpace(parts[1]), "\"'")
					}
				}
			}
		}
	}
	return ""
}

func main() {
	// 接收命令行参数：section key [defaultValue]
	if len(os.Args) < 3 {
		fmt.Println("用法：go run get_config.go <section> <key> [defaultValue]")
		os.Exit(1)
	}

	section := os.Args[1]
	key := os.Args[2]
	defaultValue := ""
	if len(os.Args) >= 4 {
		defaultValue = os.Args[3]
	}

	// 读取配置并输出
	value := getConfig(section, key)
	if value == "" {
		value = defaultValue
	}
	fmt.Print(value)
}
