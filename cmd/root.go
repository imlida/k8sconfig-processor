package cmd

import (
	"fmt"
	"os"

	"github.com/k8sconfig-processor/pkg/processor"
	"github.com/k8sconfig-processor/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	// 输入目录
	inputDir string
	// 输出目录
	outputDir string
	// 处理模式
	mode string
	// 是否强制覆盖
	force bool
	// 是否执行预检查
	precheck bool
)

// rootCmd 表示没有调用子命令时的基础命令
var rootCmd = &cobra.Command{
	Use:   "k8sconfig-processor",
	Short: "Kubernetes配置处理工具",
	Long: `Kubernetes配置处理工具，用于自动化处理K8s配置文件中的环境变量。
从ConfigMap和Secret中自动查找匹配的配置来填充未设置值的环境变量。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 创建处理选项
		options := &utils.ProcessOptions{
			InputDir:  inputDir,
			OutputDir: outputDir,
			Mode:      mode,
			Force:     force,
			Precheck:  precheck,
		}

		// 验证选项
		if err := validateOptions(options); err != nil {
			fmt.Println("选项无效:", err)
			os.Exit(1)
		}

		// 创建处理器
		mainProcessor := processor.NewMainProcessor(options)

		// 执行处理
		if err := mainProcessor.Execute(); err != nil {
			fmt.Println("执行失败:", err)
			os.Exit(1)
		}
	},
}

// 验证处理选项
func validateOptions(options *utils.ProcessOptions) error {
	// 验证输入目录
	if options.InputDir == "" {
		return fmt.Errorf("输入目录不能为空")
	}

	// 验证输入目录存在
	if _, err := os.Stat(options.InputDir); os.IsNotExist(err) {
		return fmt.Errorf("输入目录不存在: %s", options.InputDir)
	}

	// 如果是覆盖模式，检查是否设置了force标志
	if options.Mode == utils.ModeOverwrite && !options.Force {
		return fmt.Errorf("覆盖模式需要设置--force标志")
	}

	// 设置默认值
	if options.OutputDir == "" && options.Mode == utils.ModeSafe {
		options.OutputDir = utils.DefaultOutputDir
	}

	return nil
}

// Execute 添加所有子命令到根命令并设置标志
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 初始化标志
	rootCmd.PersistentFlags().StringVarP(&inputDir, "input", "i", ".", "输入目录，包含YAML文件")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", utils.DefaultOutputDir, "输出目录")
	rootCmd.PersistentFlags().StringVarP(&mode, "mode", "m", utils.ModeSafe, "处理模式: safe（安全）, overwrite（覆盖）, dry-run（演示）")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "强制覆盖模式，谨慎使用")
	rootCmd.PersistentFlags().BoolVarP(&precheck, "precheck", "p", false, "执行预检查")
}
