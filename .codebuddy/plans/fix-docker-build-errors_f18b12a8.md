---
name: fix-docker-build-errors
overview: 修复 Docker 构建错误：移除未使用的 database/sql 包导入，优化 Dockerfile 以减少敏感数据警告。
todos:
  - id: explore-project
    content: 使用 [subagent:code-explorer] 搜索并分析 main.go 和 Dockerfile 文件结构
    status: completed
  - id: fix-go-import
    content: 移除 main.go 中未使用的 `database/sql` 包导入语句
    status: completed
    dependencies:
      - explore-project
  - id: analyze-docker-warnings
    content: 检查 Dockerfile 中 ENV 设置的敏感数据，识别警告来源
    status: completed
    dependencies:
      - explore-project
  - id: optimize-dockerfile
    content: 将敏感 ENV 改为 ARG 构建参数或运行时注入方式
    status: completed
    dependencies:
      - analyze-docker-warnings
  - id: verify-build
    content: 执行 Docker 构建验证，确认错误和警告已解决
    status: completed
    dependencies:
      - fix-go-import
      - optimize-dockerfile
---

## 产品概述

修复基于 Go 语言的后端服务项目 Docker 构建错误，解决编译失败和敏感数据警告问题，确保容器化部署流程正常运行。

## 核心功能

- 移除 main.go 文件中未使用的 `database/sql` 包导入
- 优化 Dockerfile 配置以减少敏感数据在 ENV 中的警告
- 验证 Docker 镜像构建成功并通过安全扫描

## 技术栈

- **语言**：Go
- **容器化**：Docker
- **构建工具**：Docker Build

## 技术架构

### 系统架构

- **架构模式**：单层命令行应用架构（main 函数直接执行）
- **组件结构**：main.go（入口文件）→ 内部逻辑 → Dockerfile（容器配置）
- **构建流程**：Docker Build → Go 编译 → 镜像生成

### 模块划分

- **主程序模块**：main.go 包含应用入口和核心逻辑
- **容器配置模块**：Dockerfile 定义构建和运行环境
- **依赖管理**：Go modules 管理第三方包

### 数据流

代码修改 → Docker 构建触发 → Go 编译检查 → 镜像层构建 → 安全扫描 → 构建完成

## 实施细节

### 核心目录结构

```
did-login-backend/
├── main.go               # 需要修复未使用导入的主文件
├── Dockerfile            # 需要优化敏感数据配置
├── go.mod                # Go 依赖管理
└── go.sum                # 依赖校验
```

### 关键代码结构

**待修复的导入问题**：

```
// main.go 第 6 行未使用的导入
import (
    "database/sql"  // ← 需要移除，未在代码中使用
    // 其他导入...
)
```

**Dockerfile 敏感数据警告点**：

```
# 可能包含以下导致警告的配置
ENV DATABASE_PASSWORD=secret123  # 敏感数据直接通过 ENV 设置
# 需要改为构建参数或运行时注入
```

## 代理扩展

### SubAgent

- **code-explorer**
- 用途：探索项目结构，定位 main.go 和 Dockerfile 文件，分析导入依赖关系
- 预期结果：获取完整的文件路径和代码上下文，确认 database/sql 导入位置及使用情况

### MCP

- **CloudBase MCP**
- 用途：查询云开发环境配置（如项目使用 CloudBase 部署），检查相关构建配置
- 预期结果：确认是否需要额外的云构建配置调整