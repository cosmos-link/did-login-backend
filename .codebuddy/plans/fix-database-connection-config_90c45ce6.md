---
name: fix-database-connection-config
overview: 修复 main.go 中的硬编码数据库连接配置，改为从环境变量读取数据库连接信息（DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD），解决 Docker 容器无法连接阿里云 MySQL 数据库的问题。
todos:
  - id: explore-codebase
    content: 使用 [subagent:code-explorer] 定位 main.go 中的硬编码数据库配置
    status: completed
  - id: update-initdb-function
    content: 修改 initDB() 函数，从环境变量读取数据库连接信息
    status: completed
    dependencies:
      - explore-codebase
  - id: add-env-validation
    content: 添加环境变量验证和错误处理逻辑
    status: completed
    dependencies:
      - update-initdb-function
  - id: update-dockerfile
    content: 更新 Dockerfile，添加数据库环境变量定义
    status: completed
    dependencies:
      - update-initdb-function
  - id: update-docker-compose
    content: 更新 docker-compose.yml，传递数据库连接配置
    status: completed
    dependencies:
      - update-dockerfile
  - id: test-local-connection
    content: 在本地测试数据库连接功能
    status: completed
    dependencies:
      - add-env-validation
  - id: deploy-verify
    content: 部署 Docker 容器并验证容器运行正常
    status: completed
    dependencies:
      - update-docker-compose
      - test-local-connection
---

## 产品概述

修复 did-login-backend 项目中 main.go 文件的数据库连接配置问题。当前配置采用硬编码方式，导致 Docker 容器无法连接到阿里云 MySQL 数据库，容器处于重启循环状态。

## 核心功能

- 将硬编码的数据库连接信息改为从环境变量读取（DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD）
- 修复 Docker 容器的数据库连接问题
- 增加环境变量缺失的错误处理和验证

## 技术栈

- **语言**: Go
- **数据库**: MySQL (阿里云)
- **部署**: Docker

## 架构设计

### 系统架构

本项目为现有 Go 应用的小幅修改，不涉及架构变更。仅需修改 `initDB()` 函数的数据库连接初始化逻辑。

### 数据流

环境变量 → Go 应用启动 → `initDB()` 读取配置 → 建立数据库连接 → 应用正常运行

## 实现细节

### 核心目录结构

仅显示需要修改的文件：

```
did-login-backend/
├── main.go              # 修改：initDB() 函数改为从环境变量读取配置
├── Dockerfile           # 修改：添加 ENV 指令定义环境变量
└── docker-compose.yml   # 修改：在 environment 部分传递数据库连接信息
```

### 关键代码结构

**当前硬编码配置（问题）**：

```
// 当前问题代码示例
func initDB() *sql.DB {
    dsn := "user:password@tcp(host:3306)/dbname"
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        panic("failed to connect database")
    }
    return db
}
```

**修改后的环境变量配置**：

```
// 修改后的解决方案
import (
    "database/sql"
    "fmt"
    "os"
    _ "github.com/go-sql-driver/mysql"
)

func initDB() *sql.DB {
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    
    // 验证必要的环境变量
    if host == "" || port == "" || user == "" || password == "" || dbname == "" {
        panic("Missing required database environment variables")
    }
    
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        user, password, host, port, dbname)
    
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(fmt.Sprintf("Failed to connect database: %v", err))
    }
    
    // 测试连接
    if err = db.Ping(); err != nil {
        panic(fmt.Sprintf("Database ping failed: %v", err))
    }
    
    return db
}
```

### 技术实现计划

1. **问题陈述**: main.go 中的数据库连接字符串硬编码，导致 Docker 环境无法连接阿里云 MySQL
2. **解决方案**: 使用 Go 标准库 `os.Getenv()` 读取环境变量，动态构建 DSN
3. **关键技术**: Go `os` 包、`fmt.Sprintf()`、环境变量验证
4. **实现步骤**:

- 在 `initDB()` 函数中添加环境变量读取逻辑
- 添加环境变量存在性验证
- 使用 `fmt.Sprintf()` 动态构建连接字符串
- 添加连接测试（db.Ping()）
- 更新 Dockerfile 和 docker-compose.yml 配置

5. **测试策略**: 本地设置环境变量测试连接，Docker 构建后验证容器不再重启

## Agent Extensions

### SubAgent

- **code-explorer**
- 目的：探索代码库，定位 main.go 中的硬编码数据库配置
- 预期结果：准确定位 initDB() 函数及硬编码的 DSN 字符串位置