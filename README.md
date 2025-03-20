# Kubernetes 配置处理工具

一个用于自动化处理Kubernetes配置文件中环境变量的工具，能够从ConfigMap和Secret中查找并填充缺少值的环境变量。同时支持从.env文件快速生成ConfigMap和Secret资源。

## 功能特点

1. **环境变量处理**
   - 递归扫描指定目录下的所有YAML文件（支持 *.yaml和*.yml扩展名）
   - 自动识别包含Deployment、StatefulSet、DaemonSet等工作负载类型的资源文件
   - 遍历spec.template.spec.containers[*].env数组
   - 处理满足以下条件的env项：
     - 具有name字段
     - 未设置value字段
     - 未使用valueFrom引用机制

2. **值源查找优先级**
   - 先检查同命名空间的ConfigMap
     - metadata.name等于环境变量名的小写形式（示例：JWT_SECRET → jwt-secret）
     - 取data字段中同名key的值
   - 若未找到，检查同命名空间的Secret
     - metadata.name等于环境变量名的小写形式
     - 取stringData字段中同名key的值

3. **.env文件转换**
   - 支持从.env文件一键生成Kubernetes ConfigMap或Secret资源
   - 自动处理键值对并生成标准YAML格式
   - 为Secret资源添加Opaque类型
   - 自动添加source-file标签记录来源

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

### 环境变量处理

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

### .env文件转换

```bash
# 基本用法（将.env文件转换为ConfigMap，自动生成资源名称）
./k8sconfig-processor converter .env

# 生成Secret资源
./k8sconfig-processor converter .env -t secret

# 指定资源名称
./k8sconfig-processor converter .env -n my-config

# 完整用法
./k8sconfig-processor converter .env --type=secret --name=app-secret
```

## 示例

### 环境变量处理

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

### .env文件转换

.env文件内容：

```
# 数据库配置
DB_HOST=db.example.com
DB_PASSWORD=s3cr3t
API_TIMEOUT=30
```

生成的Secret资源：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
  labels:
    source-file: .env
type: Opaque
stringData:
  DB_HOST: db.example.com    
  DB_PASSWORD: s3cr3t
  API_TIMEOUT: "30"
```

## 许可证

MIT 