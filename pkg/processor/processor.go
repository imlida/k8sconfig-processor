package processor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k8sconfig-processor/pkg/parser"
	"github.com/k8sconfig-processor/pkg/utils"
)

// 主处理器
type MainProcessor struct {
	// YAML解析器
	Parser *parser.YAMLParser
	// 工作负载处理器
	WorkloadProcessor *WorkloadProcessor
	// 配置缓存
	ConfigCache *utils.ConfigCache
	// 处理报告
	Report *utils.ProcessReport
	// 处理选项
	Options *utils.ProcessOptions
	// 缓存是否已初始化
	CacheInitialized bool
}

// 创建新的主处理器
func NewMainProcessor(options *utils.ProcessOptions) *MainProcessor {
	configCache := utils.NewConfigCache()
	report := utils.NewProcessReport()

	yamlParser := parser.NewYAMLParser(configCache, report)
	workloadProcessor := NewWorkloadProcessor(configCache, report)

	return &MainProcessor{
		Parser:            yamlParser,
		WorkloadProcessor: workloadProcessor,
		ConfigCache:       configCache,
		Report:            report,
		Options:           options,
		CacheInitialized:  false,
	}
}

// 处理单个文件
func (p *MainProcessor) ProcessFile(filePath string) error {
	// 解析文件
	resources, err := p.Parser.ParseFile(filePath)
	if err != nil {
		p.Report.Errors = append(p.Report.Errors,
			fmt.Sprintf("处理文件失败: %s: %v", filePath, err))
		return err
	}

	// 只处理工作负载资源，不更新缓存
	var modified bool
	var modifiedResources []utils.KubeResource

	for i := range resources {
		resource := &resources[i]

		// 处理工作负载
		resourceModified, err := p.WorkloadProcessor.ProcessWorkload(resource)
		if err != nil {
			p.Report.Errors = append(p.Report.Errors,
				fmt.Sprintf("处理资源失败: %s: %v", filePath, err))
			continue
		}

		modifiedResources = append(modifiedResources, *resource)
		if resourceModified {
			modified = true
		}
	}

	// 如果文件被修改，写入输出
	if modified {
		p.Report.ProcessedFiles++
		return p.writeOutput(filePath, modifiedResources)
	}

	return nil
}

// 写入输出
func (p *MainProcessor) writeOutput(filePath string, resources []utils.KubeResource) error {
	// 转换为YAML
	yamlData, err := p.Parser.EncodeToYAML(resources)
	if err != nil {
		return err
	}

	// 根据输出模式处理
	var outputPath string

	if p.Options.Mode == utils.ModeSafe {
		// 安全模式：输出到新目录
		relPath, err := filepath.Rel(p.Options.InputDir, filePath)
		if err != nil {
			relPath = filepath.Base(filePath)
		}

		outputPath = filepath.Join(p.Options.OutputDir, relPath)

		// 确保输出目录存在
		outDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}

	} else if p.Options.Mode == utils.ModeOverwrite {
		// 覆盖模式：直接写回原文件
		outputPath = filePath
	} else if p.Options.Mode == utils.ModeDryRun {
		// 干运行模式：输出差异
		fmt.Printf("--- 原文件: %s\n", filePath)
		fmt.Printf("+++ 处理后:\n")
		fmt.Println(string(yamlData))
		return nil
	}

	// 写入文件
	return os.WriteFile(outputPath, yamlData, 0644)
}

// 初始化配置缓存
func (p *MainProcessor) InitializeCache(yamlFiles []string) error {
	// 如果缓存已初始化，则跳过
	if p.CacheInitialized {
		return nil
	}

	// 加载所有配置文件到缓存中
	for _, file := range yamlFiles {
		resources, err := p.Parser.ParseFile(file)
		if err != nil {
			fmt.Printf("解析文件 %s 时出错: %v\n", file, err)
			continue
		}

		BuildConfigCache(resources, p.ConfigCache)
	}

	p.CacheInitialized = true

	// 调试：打印缓存内容
	if p.Options.Mode == utils.ModeDryRun {
		fmt.Println("\n=== ConfigMap缓存内容 ===")
		for namespace, namespaceConfigs := range p.ConfigCache.ConfigMaps {
			fmt.Printf("  命名空间: %s\n", namespace)
			for name, data := range namespaceConfigs {
				fmt.Printf("    ConfigMap: %s\n", name)
				for key, value := range data {
					fmt.Printf("      %s: %s\n", key, value)
				}
			}
		}

		fmt.Println("\n=== Secret缓存内容 ===")
		for namespace, namespaceSecrets := range p.ConfigCache.Secrets {
			fmt.Printf("  命名空间: %s\n", namespace)
			for name, data := range namespaceSecrets {
				fmt.Printf("    Secret: %s\n", name)
				for key, value := range data {
					fmt.Printf("      %s: %s\n", key, value)
				}
			}
		}
	}

	return nil
}

// 执行处理
func (p *MainProcessor) Execute() error {
	// 扫描目录
	yamlFiles, err := p.Parser.ScanDirectory(p.Options.InputDir)
	if err != nil {
		return err
	}

	fmt.Printf("找到 %d 个YAML文件\n", len(yamlFiles))

	// 初始化配置缓存
	if err := p.InitializeCache(yamlFiles); err != nil {
		return err
	}

	// 处理所有文件
	for _, file := range yamlFiles {
		if err := p.ProcessFile(file); err != nil {
			fmt.Printf("处理文件 %s 时出错: %v\n", file, err)
		}
	}

	// 输出报告
	p.PrintReport()

	return nil
}

// 打印报告
func (p *MainProcessor) PrintReport() {
	fmt.Println("\n===== 处理报告 =====")
	fmt.Printf("文件总数: %d\n", p.Report.TotalFiles)
	fmt.Printf("处理的文件数: %d\n", p.Report.ProcessedFiles)
	fmt.Printf("成功更新的资源数: %d\n", p.Report.SuccessfulUpdates)

	if len(p.Report.Warnings) > 0 {
		fmt.Println("\n警告:")
		for _, warning := range p.Report.Warnings {
			fmt.Printf("- %s\n", warning)
		}
	}

	if len(p.Report.Errors) > 0 {
		fmt.Println("\n错误:")
		for _, err := range p.Report.Errors {
			fmt.Printf("- %s\n", err)
		}
	}
}
