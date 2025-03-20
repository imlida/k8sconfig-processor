# Kubernetes 配置处理工具

一个用于自动化处理Kubernetes配置文件中环境变量的工具，能够从ConfigMap和Secret中查找并填充缺少值的环境变量。

## 功能特点

1. **输入处理**
   - 递归扫描指定目录下的所有YAML文件（支持 *.yaml和*.yml扩展名）
   - 自动识别包含Deployment、StatefulSet、DaemonSet等工作负载类型的资源文件

2. **环境变量处理**
   - 遍历spec.template.spec.containers[*].env数组
   - 处理满足以下条件的env项：
     - 具有name字段
     - 未设置value字段
     - 未使用valueFrom引用机制

3. **值源查找优先级**
   - 先检查同命名空间的ConfigMap
     - metadata.name等于环境变量名的小写形式（示例：JWT_SECRET → jwt-secret）
     - 取data字段中同名key的值
   - 若未找到，检查同命名空间的Secret
     - metadata.name等于环境变量名的小写形式
     - 取stringData字段中同名key的值

4. **异常处理**
   - 未找到对应配置时保留原结构，添加警告
   - 处理失败时输出详细错误日志

5. **输出策略**
   - 安全模式：保留原文件，生成带注释的版本（默认）
   - 覆盖模式：原地更新（需添加-force参数）
   - 差异对比：输出git-style diff（-dry-run模式）

## 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/k8sconfig-processor.git
cd k8sconfig-processor

# 构建项目
go build -o k8sconfig-processor
```

## 使用方法

```bash
# 基本用法（处理当前目录下的所有YAML文件，输出到./processed/）
./k8sconfig-processor

# 指定输入目录
./k8sconfig-processor -i ./my-k8s-configs/

# 指定输出目录
./k8sconfig-processor -o ./processed-configs/

# 覆盖模式（需要强制标志）
./k8sconfig-processor -m overwrite -f

# 差异对比模式
./k8sconfig-processor -m dry-run

# 执行预检查
./k8sconfig-processor -p
```

## 示例

原始配置：

```yaml
env:
  - name: JWT_SECRET
  - name: PGDATABASE
```

处理后：

```yaml
env:
  - name: JWT_SECRET
    valueFrom:
      secretKeyRef:
        name: jwt-secret
        key: JWT_SECRET
  - name: PGDATABASE
    valueFrom:
      configMapKeyRef:
        name: pgdatabase
        key: PGDATABASE
```

## 许可证

MIT 