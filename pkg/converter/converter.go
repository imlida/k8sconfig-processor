package converter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/k8sconfig-processor/pkg/utils"
	"gopkg.in/yaml.v3"
)

// 从.env文件生成Kubernetes资源
func GenerateFromEnvFile(filePath string, resourceType string, resourceName string) error {
	// 读取并解析.env文件
	envVars, err := parseEnvFile(filePath)
	if err != nil {
		return fmt.Errorf("无法解析.env文件: %w", err)
	}

	// 检查是否有任何环境变量
	if len(envVars) == 0 {
		return fmt.Errorf(".env文件为空或格式不正确")
	}

	// 创建对应的资源对象
	var resource utils.KubeResource

	// 设置通用字段
	resource.APIVersion = "v1"
	resource.Metadata.Name = resourceName

	// 根据类型设置特定字段
	switch resourceType {
	case "cm":
		resource.Kind = "ConfigMap"
		resource.Data = envVars
	case "secret":
		resource.Kind = "Secret"
		resource.StringData = envVars
		resource.Type = "Opaque"

		// 添加来源标签
		resource.Metadata.Labels = map[string]string{
			"source-file": filepath.Base(filePath),
		}
	}

	// 转换为YAML
	yamlData, err := yaml.Marshal(resource)
	if err != nil {
		return fmt.Errorf("无法生成YAML: %w", err)
	}

	// 输出生成的YAML
	fmt.Println(string(yamlData))

	// 保存到文件（选项）
	outputFile := getOutputFileName(resourceName, resourceType)
	err = os.WriteFile(outputFile, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("无法保存到文件: %w", err)
	}

	fmt.Printf("已生成 %s 并保存到 %s\n", resourceType, outputFile)
	return nil
}

// 解析.env文件内容
func parseEnvFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 分割键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // 跳过格式不正确的行
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 去除值中的引号
		value = strings.Trim(value, "\"'")

		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// 获取输出文件名
func getOutputFileName(resourceName string, resourceType string) string {
	var suffix string
	if resourceType == "cm" {
		suffix = "configmap"
	} else {
		suffix = "secret"
	}

	return fmt.Sprintf("%s-%s.yaml", resourceName, suffix)
}
