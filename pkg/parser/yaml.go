package parser

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/k8sconfig-processor/pkg/utils"
	"gopkg.in/yaml.v3"
)

// YAML文件解析器
type YAMLParser struct {
	// 配置缓存
	ConfigCache *utils.ConfigCache
	// 处理报告
	Report *utils.ProcessReport
}

// 创建新的YAML解析器
func NewYAMLParser(configCache *utils.ConfigCache, report *utils.ProcessReport) *YAMLParser {
	return &YAMLParser{
		ConfigCache: configCache,
		Report:      report,
	}
}

// 扫描目录中的所有YAML文件
func (p *YAMLParser) ScanDirectory(directory string) ([]string, error) {
	var yamlFiles []string

	// 递归遍历目录
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if ext == utils.YamlExt || ext == utils.YmlExt {
			yamlFiles = append(yamlFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	p.Report.TotalFiles = len(yamlFiles)
	return yamlFiles, nil
}

// 解析单个YAML文件中的所有文档
func (p *YAMLParser) ParseFile(filePath string) ([]utils.KubeResource, error) {
	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析文件中的多个YAML文档
	var resources []utils.KubeResource
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var resource utils.KubeResource
		err := decoder.Decode(&resource)
		if err == io.EOF {
			break
		}
		if err != nil {
			p.Report.Errors = append(p.Report.Errors,
				"解析文件失败: "+filePath+": "+err.Error())
			continue
		}

		// 确保命名空间字段有值
		if resource.Metadata.Namespace == "" {
			resource.Metadata.Namespace = utils.DefaultNamespace
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// 将资源编码为YAML
func (p *YAMLParser) EncodeToYAML(resources []utils.KubeResource) ([]byte, error) {
	var resultBuf bytes.Buffer

	// 遍历所有资源
	for i, resource := range resources {
		// 使用yaml.Marshal来编码资源
		yamlData, err := yaml.Marshal(resource)
		if err != nil {
			return nil, err
		}

		// 将编码后的数据写入结果缓冲区
		resultBuf.Write(yamlData)

		// 除了最后一个资源外，在每个资源后添加分隔符
		if i < len(resources)-1 {
			resultBuf.WriteString("---\n")
		}
	}

	return resultBuf.Bytes(), nil
}
