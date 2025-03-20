package utils

// 支持的Kubernetes资源类型
const (
	// 工作负载类型
	DeploymentKind  = "Deployment"
	StatefulSetKind = "StatefulSet"
	DaemonSetKind   = "DaemonSet"

	// 配置资源类型
	ConfigMapKind = "ConfigMap"
	SecretKind    = "Secret"

	// YAML文件扩展名
	YamlExt = ".yaml"
	YmlExt  = ".yml"

	// 默认命名空间
	DefaultNamespace = "default"

	// 处理模式
	ModeSafe      = "safe"      // 保留原文件，生成新版本
	ModeOverwrite = "overwrite" // 原地更新
	ModeDryRun    = "dry-run"   // 只输出差异

	// 输出目录
	DefaultOutputDir = "./processed"
)
