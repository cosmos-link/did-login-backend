---
name: fix-dockerfile-copy-path
overview: 修复 Dockerfile 中 go.mod 和 go.sum 文件路径错误，将 COPY 指令从 `go.mod go.sum ./` 修正为 `src/go.mod src/go.sum ./`，解决 Docker 构建时找不到依赖文件的问题。
todos:
  - id: locate-dockerfile
    content: 定位并读取项目 Dockerfile 内容
    status: completed
  - id: fix-copy-path
    content: 修改 COPY 指令路径为 src/go.mod src/go.sum ./
    status: completed
    dependencies:
      - locate-dockerfile
  - id: verify-docker-build
    content: 执行 Docker 构建验证修复效果
    status: completed
    dependencies:
      - fix-copy-path
---

## 产品概述

修复 Dockerfile 中 go.mod 和 go.sum 文件路径错误，解决 Docker 构建时找不到依赖文件的问题。

## 核心功能

- 定位项目中的 Dockerfile 文件
- 修正 COPY 指令路径从 `go.mod go.sum ./` 到 `src/go.mod src/go.sum ./`
- 验证 Docker 构建流程是否成功

## 技术栈

- Docker

## 实现细节

### 核心目录结构

仅显示需要修改的 Dockerfile 文件：

```
project-root/
├── src/
│   ├── go.mod
│   └── go.sum
└── Dockerfile          # 需要修改的文件
```

### 关键代码结构

**Dockerfile 修改内容**：

需要修改的 COPY 指令位置：

```
# 修改前
COPY go.mod go.sum ./

# 修改后
COPY src/go.mod src/go.sum ./
```

**修改原因分析**：

- 构建上下文根目录下不存在 go.mod 和 go.sum 文件
- 这两个文件实际存放在 src/ 子目录中
- Docker 构建时从构建上下文查找文件，原路径错误导致构建失败并提示找不到 /go.sum