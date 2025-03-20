package parser

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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

// Deployment工作负载的YAML模板
const deploymentTemplate = `apiVersion: {{.APIVersion}}
kind: {{.Kind}}
metadata:
  name: {{.Metadata.Name}}
  namespace: {{.Metadata.Namespace}}
spec:
{{- if .Spec.replicas}}
  replicas: {{.Spec.replicas}}
{{- end}}
{{- if .Spec.selector}}
  selector:
    matchLabels:
    {{- range $key, $value := .Spec.selector.matchLabels}}
      {{$key}}: {{$value}}
    {{- end}}
{{- end}}
{{- if .Spec.template}}
  template:
    metadata:
      labels:
      {{- range $key, $value := .Spec.template.metadata.labels}}
        {{$key}}: {{$value}}
      {{- end}}
    spec:
      containers:
      {{- range .Spec.template.spec.containers}}
      - name: {{.name}}
        image: {{.image}}
        {{- if .env}}
        env:
        {{- range .env}}
        - name: {{.name}}
          {{- if .value}}
          value: {{.value}}
          {{- end}}
          {{- if .valueFrom}}
          valueFrom:
            {{- if .valueFrom.configMapKeyRef}}
            configMapKeyRef:
              name: {{.valueFrom.configMapKeyRef.name}}
              key: {{.valueFrom.configMapKeyRef.key}}
            {{- end}}
            {{- if .valueFrom.secretKeyRef}}
            secretKeyRef:
              name: {{.valueFrom.secretKeyRef.name}}
              key: {{.valueFrom.secretKeyRef.key}}
            {{- end}}
          {{- end}}
        {{- end}}
        {{- end}}
      {{- end}}
{{- end}}
`

// 将资源编码为YAML
func (p *YAMLParser) EncodeToYAML(resources []utils.KubeResource) ([]byte, error) {
	var resultBuf bytes.Buffer

	for i, resource := range resources {
		// 只对Deployment类型使用模板，其他类型使用通用方法
		if resource.Kind == utils.DeploymentKind {
			// 解析模板
			tmpl, err := template.New("deployment").Parse(deploymentTemplate)
			if err != nil {
				return nil, fmt.Errorf("解析模板失败: %v", err)
			}

			// 执行模板
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, resource); err != nil {
				return nil, fmt.Errorf("执行模板失败: %v", err)
			}

			resultBuf.Write(buf.Bytes())
		} else {
			// 对于非Deployment资源，使用常规encoding
			var buf bytes.Buffer
			encoder := yaml.NewEncoder(&buf)
			encoder.SetIndent(2)

			if err := encoder.Encode(resource); err != nil {
				return nil, err
			}

			resultBuf.Write(buf.Bytes())
		}

		// 除了最后一个资源外，在每个资源后添加分隔符
		if i < len(resources)-1 {
			resultBuf.WriteString("---\n")
		}
	}

	return resultBuf.Bytes(), nil
}
