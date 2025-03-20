package processor

import (
	"fmt"

	"github.com/k8sconfig-processor/pkg/utils"
)

// 工作负载处理器
type WorkloadProcessor struct {
	// 配置缓存
	ConfigCache *utils.ConfigCache
	// 处理报告
	Report *utils.ProcessReport
}

// 创建新的工作负载处理器
func NewWorkloadProcessor(configCache *utils.ConfigCache, report *utils.ProcessReport) *WorkloadProcessor {
	return &WorkloadProcessor{
		ConfigCache: configCache,
		Report:      report,
	}
}

// 判断是否为工作负载资源
func IsWorkloadResource(kind string) bool {
	return kind == utils.DeploymentKind ||
		kind == utils.StatefulSetKind ||
		kind == utils.DaemonSetKind
}

// 处理工作负载资源
func (p *WorkloadProcessor) ProcessWorkload(resource *utils.KubeResource) (bool, error) {
	// 验证资源类型
	if !IsWorkloadResource(resource.Kind) {
		return false, nil
	}

	modified := false
	namespace := resource.Metadata.Namespace
	resourceName := resource.Metadata.Name

	// 获取pod模板
	template, exists := resource.Spec["template"]
	if !exists {
		p.Report.Warnings = append(p.Report.Warnings,
			fmt.Sprintf("资源 %s/%s 没有template字段", namespace, resourceName))
		return false, nil
	}

	// 获取pod规格
	templateMap, ok := template.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("资源 %s/%s 的template字段格式无效", namespace, resourceName)
	}

	spec, exists := templateMap["spec"]
	if !exists {
		p.Report.Warnings = append(p.Report.Warnings,
			fmt.Sprintf("资源 %s/%s 没有spec字段", namespace, resourceName))
		return false, nil
	}

	// 获取容器列表
	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("资源 %s/%s 的spec字段格式无效", namespace, resourceName)
	}

	containers, exists := specMap["containers"]
	if !exists {
		p.Report.Warnings = append(p.Report.Warnings,
			fmt.Sprintf("资源 %s/%s 没有containers字段", namespace, resourceName))
		return false, nil
	}

	// 遍历容器
	containersList, ok := containers.([]interface{})
	if !ok {
		return false, fmt.Errorf("资源 %s/%s 的containers字段格式无效", namespace, resourceName)
	}

	for i, container := range containersList {
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			continue
		}

		// 处理环境变量
		envModified, err := p.processContainerEnv(containerMap, namespace, resourceName)
		if err != nil {
			return false, err
		}

		if envModified {
			modified = true
			// 更新容器
			containersList[i] = containerMap
		}
	}

	// 如果有修改，更新容器列表
	if modified {
		specMap["containers"] = containersList
		templateMap["spec"] = specMap
		resource.Spec["template"] = templateMap
		p.Report.SuccessfulUpdates++
	}

	return modified, nil
}

// 处理容器环境变量
func (p *WorkloadProcessor) processContainerEnv(container map[string]interface{}, namespace, resourceName string) (bool, error) {
	modified := false

	// 获取环境变量列表
	env, exists := container["env"]
	if !exists {
		// 没有环境变量，不需要处理
		return false, nil
	}

	envList, ok := env.([]interface{})
	if !ok {
		return false, fmt.Errorf("资源 %s/%s 的env字段格式无效", namespace, resourceName)
	}

	// 遍历环境变量
	for i, envVar := range envList {
		envVarMap, ok := envVar.(map[string]interface{})
		if !ok {
			continue
		}

		// 检查是否满足处理条件
		name, hasName := envVarMap["name"]
		_, hasValue := envVarMap["value"]
		_, hasValueFrom := envVarMap["valueFrom"]

		// 如果有name字段，但没有value和valueFrom，则需要处理
		if hasName && !hasValue && !hasValueFrom {
			envName, ok := name.(string)
			if !ok {
				continue
			}

			// 从缓存中查找配置值
			_, configName, configKind, found := FindConfigValue(envName, namespace, p.ConfigCache)
			if found {
				// 根据类型创建valueFrom引用
				valueFrom := make(map[string]interface{})

				if configKind == utils.ConfigMapKind {
					configMapKeyRef := map[string]interface{}{
						"name": configName,
						"key":  envName,
					}
					valueFrom["configMapKeyRef"] = configMapKeyRef
				} else if configKind == utils.SecretKind {
					secretKeyRef := map[string]interface{}{
						"name": configName,
						"key":  envName,
					}
					valueFrom["secretKeyRef"] = secretKeyRef
				}

				// 更新环境变量
				envVarMap["valueFrom"] = valueFrom
				envList[i] = envVarMap
				modified = true
			} else {
				// 未找到配置，添加警告注释
				p.Report.Warnings = append(p.Report.Warnings,
					fmt.Sprintf("未找到环境变量 %s 的配置 (资源: %s/%s)",
						envName, namespace, resourceName))
			}
		}
	}

	// 如果有修改，更新环境变量列表
	if modified {
		container["env"] = envList
	}

	return modified, nil
}
