package processor

import (
	"strings"

	"github.com/k8sconfig-processor/pkg/utils"
)

// 构建ConfigMap缓存
func BuildConfigCache(resources []utils.KubeResource, cache *utils.ConfigCache) {
	for _, resource := range resources {
		namespace := resource.Metadata.Namespace
		name := resource.Metadata.Name

		// 处理ConfigMap
		if resource.Kind == utils.ConfigMapKind && resource.Data != nil {
			// 确保命名空间映射存在
			if _, exists := cache.ConfigMaps[namespace]; !exists {
				cache.ConfigMaps[namespace] = make(map[string]map[string]string)
			}

			// 添加或更新ConfigMap数据
			cache.ConfigMaps[namespace][name] = resource.Data
		}

		// 处理Secret
		if resource.Kind == utils.SecretKind {
			// 确保命名空间映射存在
			if _, exists := cache.Secrets[namespace]; !exists {
				cache.Secrets[namespace] = make(map[string]map[string]string)
			}

			// 处理stringData字段
			if resource.StringData != nil {
				cache.Secrets[namespace][name] = resource.StringData
			}

			// 处理data字段
			if resource.Data != nil {
				// 如果该命名空间下还没有该名称的Secret，则初始化
				if _, exists := cache.Secrets[namespace][name]; !exists {
					cache.Secrets[namespace][name] = make(map[string]string)
				}

				// 将data中的数据复制到缓存中
				for key, value := range resource.Data {
					cache.Secrets[namespace][name][key] = value
				}
			}
		}
	}
}

// 从缓存中查找配置
func FindConfigValue(envName string, namespace string, cache *utils.ConfigCache) (string, string, string, bool) {
	// 生成配置对象名称（小写并替换下划线为中划线）
	configName := strings.ToLower(strings.ReplaceAll(envName, "_", "-"))

	// 首先检查ConfigMap
	if namespaceConfigs, exists := cache.ConfigMaps[namespace]; exists {
		if data, exists := namespaceConfigs[configName]; exists {
			if value, exists := data[envName]; exists {
				return value, configName, utils.ConfigMapKind, true
			}
		}
	}

	// 然后检查Secret
	if namespaceSecrets, exists := cache.Secrets[namespace]; exists {
		if data, exists := namespaceSecrets[configName]; exists {
			if value, exists := data[envName]; exists {
				return value, configName, utils.SecretKind, true
			}
		}
	}

	// 未找到配置
	return "", "", "", false
}
