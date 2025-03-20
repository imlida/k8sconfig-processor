package utils

// 配置对象缓存
type ConfigCache struct {
	// 按类型存储的配置缓存: map[资源类型][namespace][name]map[key]value
	ConfigMaps map[string]map[string]map[string]string
	Secrets    map[string]map[string]map[string]string
}

// 新建配置缓存
func NewConfigCache() *ConfigCache {
	return &ConfigCache{
		ConfigMaps: make(map[string]map[string]map[string]string),
		Secrets:    make(map[string]map[string]map[string]string),
	}
}

// 处理报告
type ProcessReport struct {
	// 处理统计
	TotalFiles        int
	ProcessedFiles    int
	SuccessfulUpdates int

	// 警告和错误
	Warnings []string
	Errors   []string
}

// 新建处理报告
func NewProcessReport() *ProcessReport {
	return &ProcessReport{
		Warnings: make([]string, 0),
		Errors:   make([]string, 0),
	}
}

// 处理选项
type ProcessOptions struct {
	// 输入目录
	InputDir string

	// 输出目录
	OutputDir string

	// 处理模式
	Mode string

	// 是否强制覆盖
	Force bool

	// 是否执行预检查
	Precheck bool
}

// 解析的K8s资源
type KubeResource struct {
	// API版本
	APIVersion string `yaml:"apiVersion"`

	// 资源类型
	Kind string `yaml:"kind"`

	// 元数据
	Metadata struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace,omitempty"`
	} `yaml:"metadata"`

	// 规格(仅适用于工作负载)
	Spec map[string]interface{} `yaml:"spec,omitempty"`

	// 数据(用于ConfigMap)
	Data map[string]string `yaml:"data,omitempty"`

	// 字符串数据(用于Secret)
	StringData map[string]string `yaml:"stringData,omitempty"`
}
