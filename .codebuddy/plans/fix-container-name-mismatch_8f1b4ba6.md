---
name: fix-container-name-mismatch
overview: 修复 Docker 部署中容器和镜像文件名不匹配问题。CI/CD 使用 go-webserver 命名，但实际部署和验证使用 did-new，导致容器找不到和镜像文件缺失错误。
todos:
  - id: search-naming-issues
    content: 使用 [subagent:code-explorer] 搜索所有命名不一致的文件
    status: completed
  - id: decide-naming-standard
    content: 确定统一的容器和镜像命名标准
    status: completed
    dependencies:
      - search-naming-issues
  - id: update-github-actions
    content: 修改 GitHub Actions 工作流中的容器和镜像名称
    status: completed
    dependencies:
      - decide-naming-standard
  - id: update-deploy-scripts
    content: 修改部署脚本中的容器和镜像引用
    status: completed
    dependencies:
      - decide-naming-standard
  - id: update-config-ini
    content: 更新 config.ini 中的应用名称配置
    status: completed
    dependencies:
      - decide-naming-standard
  - id: test-deployment
    content: 在测试环境验证完整部署流程
    status: completed
    dependencies:
      - update-github-actions
      - update-deploy-scripts
      - update-config-ini
---

## 问题概述

修复 Docker 部署流程中容器名称和镜像文件名不匹配的问题。CI/CD 配置使用 `go-webserver` 命名，而实际部署脚本使用 `did-new` 命名，导致容器无法找到和镜像文件缺失错误。

## 核心修复内容

- 统一容器命名规范：将 `go-webserver-container` 与 `did-new-container` 对齐
- 统一镜像文件名：将 `go-webserver` 相关镜像名与 `did-new.tar` 对齐
- 更新 config.ini 中的应用程序名称配置，确保与部署脚本一致
- 修复 GitHub Actions 工作流中的容器验证步骤

## 技术栈

- 容器化：Docker
- CI/CD：GitHub Actions
- 配置管理：INI 文件格式

## 技术架构

### 系统架构

这是一个配置修复任务，无需复杂的系统架构变更。主要涉及三个配置层：

- **CI/CD 层**：GitHub Actions 工作流文件（`.github/workflows/`）
- **部署脚本层**：Shell/Python 部署脚本
- **应用配置层**：`config.ini`

### 模块划分

- **GitHub Actions 模块**：负责构建和推送 Docker 镜像
- **部署脚本模块**：负责加载镜像和启动容器
- **配置管理模块**：定义应用基础名称

### 数据流

GitHub Actions 构建 → 生成镜像 tar 包 → 上传产物 → 部署脚本下载 → 加载镜像 → 启动容器 → 健康检查

## 实现细节

### 核心目录结构

```
did-login-backend/
├── .github/
│   └── workflows/
│       └── deploy.yml          # 需要修改：容器名称和镜像文件名
├── scripts/
│   └── deploy.sh               # 需要修改：容器和镜像命名
└── config.ini                  # 需要修改：app_name 配置
```

### 关键配置项

**config.ini 示例**：

```
[app]
name=go-webserver  # 需要统一为 did-new 或保持 go-webserver
```

**GitHub Actions 片段**：

```
# 需要统一命名
container_name: go-webserver-container
image_file: did-new.tar
```

### 技术实现计划

1. **问题定位**：搜索所有使用 `go-webserver` 和 `did-new` 的文件位置
2. **命名决策**：确定统一使用 `go-webserver` 还是 `did-new` 作为标准名称
3. **配置更新**：批量替换所有不一致的命名引用
4. **验证测试**：在测试环境执行完整部署流程验证修复效果

### 集成点

- Docker 镜像仓库或 GitHub Artifacts
- 目标部署服务器的 Docker 守护进程
- GitHub Actions Runner 与部署服务器的通信

## Agent Extensions

### SubAgent

- **code-explorer**
- 用途：搜索代码库中所有包含 `go-webserver` 和 `did-new` 命名的文件
- 预期结果：生成包含所有不匹配位置的文件清单和具体行号