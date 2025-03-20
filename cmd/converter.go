package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/k8sconfig-processor/pkg/converter"
	"github.com/spf13/cobra"
)

var (
	// 资源类型
	resourceType string
	// 资源名称
	resourceName string
)

// converterCmd 表示converter命令
var converterCmd = &cobra.Command{
	Use:   "converter [.env文件路径]",
	Short: "K8s配置转换工具",
	Long:  `将.env文件转换为Kubernetes ConfigMap或Secret资源`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 获取.env文件路径
		filePath := args[0]

		// 验证文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("错误: 文件不存在: %s\n", filePath)
			os.Exit(1)
		}

		// 验证资源类型
		if resourceType != "cm" && resourceType != "secret" {
			fmt.Println("错误: 资源类型必须是 cm 或 secret")
			os.Exit(1)
		}

		// 如果未指定资源名称，使用文件名作为默认值
		if resourceName == "" {
			baseFileName := filepath.Base(filePath)
			fileNameWithoutExt := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))

			if resourceType == "cm" {
				resourceName = fileNameWithoutExt + "-config"
			} else {
				resourceName = fileNameWithoutExt + "-secret"
			}
		}

		// 执行转换
		err := converter.GenerateFromEnvFile(filePath, resourceType, resourceName)
		if err != nil {
			fmt.Printf("转换失败: %s\n", err)
			os.Exit(1)
		}
	},
	Example: `  # 生成Secret资源
  converter .env --type=secret --name=app-secret
  
  # 生成ConfigMap资源
  converter .env --type=cm --name=app-config
  
  # 简化形式(默认名称基于文件名)
  converter .env -t cm`,
}

func init() {
	// 添加converter命令到根命令
	rootCmd.AddCommand(converterCmd)

	// 添加命令的标志
	converterCmd.Flags().StringVarP(&resourceType, "type", "t", "cm", "资源类型: cm或secret")
	converterCmd.Flags().StringVarP(&resourceName, "name", "n", "", "资源名称(默认基于文件名)")
}
