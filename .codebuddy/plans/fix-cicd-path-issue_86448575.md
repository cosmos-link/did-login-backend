---
name: fix-cicd-path-issue
overview: 修复CI/CD工作流中Go源码路径检查错误，将错误的路径 `src/Go-WebServer-V1.1/main.go` 修正为正确的 `src/main.go`
todos:
  - id: search-incorrect-path
    content: 使用 [subagent:code-explorer] 搜索所有引用错误路径的文件
    status: completed
  - id: update-workflow-files
    content: 批量修正 GitHub Actions 工作流文件中的路径配置
    status: completed
    dependencies:
      - search-incorrect-path
  - id: verify-path-corrections
    content: 验证所有路径已正确更新为 src/main.go
    status: completed
    dependencies:
      - update-workflow-files
---

## 产品概述

修复 GitHub CI/CD 工作流中 Go 源码路径检查错误，将错误路径 `src/Go-WebServer-V1.1/main.go` 修正为正确路径 `src/main.go`，确保 CI/CD 流水线能够正确识别和构建 Go 应用程序。

## 核心功能

- 定位所有引用错误路径的 CI/CD 配置文件
- 批量修正路径配置
- 验证修复结果确保无遗漏

## 技术栈

- CI/CD 平台：GitHub Actions
- 项目类型：Go Web 服务器
- 目标路径：`src/main.go`

## 实施细节

### 核心目录结构

项目 CI/CD 配置位于：

```
project-root/
├── .github/
│   └── workflows/
│       └── *.yml          # GitHub Actions 工作流文件
└── src/
    └── main.go            # 正确的 Go 入口文件
```

### 关键代码结构

需要全局替换的路径模式：

- 错误路径：`src/Go-WebServer-V1.1/main.go`
- 正确路径：`src/main.go`

## 代理扩展

### SubAgent

- **code-explorer**
- 用途：在项目全范围内搜索包含错误路径字符串的所有文件
- 预期结果：返回所有需要修改的 CI/CD 配置文件列表及具体位置