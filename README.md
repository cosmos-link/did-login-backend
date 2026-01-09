# 区块链 DID 安全门户

基于以太坊区块链和 WebAuthn 技术的去中心化身份认证系统，支持 DID 身份管理和生物识别登录。

## 📋 项目简介

本项目是一个完整的去中心化身份（DID）认证系统，集成了：
- 🔗 以太坊区块链 DID 生成（基于助记词）
- 📱 WebAuthn 指纹/面部识别（Touch ID/Face ID）
- 🔐 JWT 令牌认证（7天有效期）
- 🗄️ MySQL 数据库存储
- 🐳 Docker 容器化部署
- 🔑 助记词恢复和密码重置

## 🛠️ 技术栈

- **前端**: HTML5 + JavaScript + Ethers.js 5.7.2 + Tailwind CSS
- **后端**: Go 1.24 + Gin 框架 + GORM + JWT + bcrypt
- **数据库**: MySQL 8.0
- **容器化**: Docker + Docker Compose
- **身份认证**: WebAuthn (FIDO2) 生物识别 + 密码双重验证

## 🚀 快速启动

### 1. 环境要求

- Docker 和 Docker Compose
- 支持 WebAuthn 的现代浏览器（Chrome 67+, Firefox 60+, Safari 14+）
- 具备指纹识别或面部识别功能的设备

### 2. 启动项目

```bash
# 克隆项目
cd /Users/chilly/go/src/github.com/cosmos-link/did-login

# 使用 Docker Compose 启动所有服务
docker-compose up --build -d

# 查看容器状态
docker-compose ps

# 查看后端日志
docker logs portal_backend --tail 20
```

### 3. 访问应用

- **前端界面**: http://localhost:3001
- **后端 API**: http://localhost:8080
- **数据库**: localhost:3306 (root/password123)

### 4. 停止服务

```bash
# 停止所有容器
docker-compose down

# 停止并删除数据卷（重置数据库）
docker-compose down -v
```

## 📋 API 接口使用指南

### 应用管理接口

#### 1. 添加新应用
```bash
curl -X POST "http://localhost:8080/api/apps" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "应用名称",
    "container_name": "container-name", 
    "port": 3008,
    "base_url": "http://localhost",
    "description": "应用描述",
    "user_types": ["企业", "机构", "个人", "社区", "政府"]
  }'
```

#### 实际示例 - 添加钱包应用：
```bash
curl -X POST "http://localhost:8080/api/apps" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的钱包",
    "container_name": "my-wallet",
    "port": 3008,
    "base_url": "http://localhost", 
    "description": "管理个人钱包",
    "user_types": ["企业", "机构", "个人", "社区", "政府"]
  }'
```

#### 2. 获取应用列表
```bash
# 获取个人用户可访问的应用
curl -X GET "http://localhost:8080/api/apps?user_type=个人"

# 获取企业用户可访问的应用  
curl -X GET "http://localhost:8080/api/apps?user_type=企业"
```

#### 3. 删除应用
```bash
# 删除ID为7的应用
curl -X DELETE "http://localhost:8080/api/apps/7"
```

### 用户认证接口

#### 1. 用户注册
```bash
curl -X POST "http://localhost:8080/api/register" \
  -H "Content-Type: application/json" \
  -d '{
    "did": "0x1234567890abcdef...",
    "email": "user@example.com",
    "password": "password123",
    "user_type": "个人"
  }'
```

#### 2. 用户登录
```bash
curl -X POST "http://localhost:8080/api/login/basic" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "chilly.zhong32@gmail.com",
    "password": "123456"
  }'
```

#### 3. DID验证
```bash
curl -X POST "http://localhost:8080/api/verify-did" \
  -H "Content-Type: application/json" \
  -d '{
    "did": "0x1234567890abcdef..."
  }'
```

## 📱 使用指南

### 注册新账户

1. 访问 http://localhost:3001
2. 点击"新用户注册"
3. 填写邮箱、选择用户类型、设置密码
4. 点击"生成 DID 并注册"
5. 系统会使用 ethers.js 生成：
   - **DID**: 以太坊地址（如 0x742d35Cc...）
   - **助记词**: 12位助记词用于恢复身份
6. **重要**: 请妥善备份显示的助记词
7. 可选择设置 Touch ID/Face ID 增强安全性

### 登录系统

**方式一：邮箱密码登录（推荐）**
1. 输入注册的邮箱和密码
2. 点击"登录并验证指纹"
3. 基础认证成功后：
   - 如已设置生物识别，弹出 Touch ID/Face ID 验证
   - 如未设置，系统询问是否现在设置
4. 验证成功后获得 **7天有效期** 的 JWT Token

**方式二：助记词恢复登录**
1. 点击"助记词找回"
2. 输入 12位助记词（空格分隔）
3. 系统验证 DID 并显示关联邮箱
4. 设置新密码（至少6位）
5. 系统自动跳转登录页并填充凭据
6. 自动完成登录流程

### 应用访问

登录成功后：
1. 自动跳转到应用列表页
2. 根据用户类型显示不同的应用
3. 用户信息显示在顶部（邮箱、DID、类型）
4. 点击应用卡片可访问对应服务

### 指纹认证设置

首次登录时，系统会询问是否设置指纹认证：
1. 点击"确定"同意设置
2. 浏览器弹出生物识别界面
3. 使用 Touch ID 或 Face ID 完成录入
4. 下次登录可使用指纹快速验证

## 🔧 开发调试

### 查看日志

```bash
# 后端日志
docker logs portal_backend -f

# 数据库日志
docker logs portal_db -f

# 前端日志
docker logs portal_frontend -f
```

### 进入容器调试

```bash
# 进入后端容器
docker exec -it portal_backend /bin/bash

# 进入数据库容器
docker exec -it portal_db mysql -u portal_user -p
```

### 单独启动服务

```bash
# 只启动数据库
docker-compose up -d db

# 只启动后端
docker-compose up --build -d backend

# 只启动前端
docker-compose up -d frontend
```

## 📁 项目结构

```
did-login/
├── docker-compose.yaml          # Docker编排配置
├── README.md                    # 项目说明文档
├── API_DOCUMENTATION.md         # API接口详细文档
├── dev-sync.sh                  # 开发同步脚本
├── backend/                     # Go后端服务
│   ├── Dockerfile              # 后端容器配置
│   ├── go.mod                  # Go依赖管理
│   ├── go.sum                  # 依赖版本锁定
│   ├── main.go                 # 主服务程序（路由+业务逻辑）
│   └── models.go               # 数据模型定义（User/App/Permission）
└── frontend/                    # 前端静态文件
    ├── index.html              # 主页面（登录/注册/恢复界面）
    └── app.js                  # JavaScript逻辑（ethers.js + WebAuthn）
```

## 🗄️ 数据模型

### 用户表 (users)
- `did` (主键): 以太坊地址作为唯一标识
- `email`: 唯一邮箱
- `password_hash`: bcrypt 加密密码
- `user_type`: 用户类型（企业/个人/社区/机构/政府）
- `credential_id`: WebAuthn 凭证ID
- `public_key`: WebAuthn 公钥
- `sign_count`: 防重放计数器

### 应用表 (applications)
- `app_id` (主键): 应用唯一ID
- `name`: 应用名称
- `container_name`: Docker 容器名
- `port`: 访问端口
- `base_url`: 基础URL
- `description`: 应用描述

### 权限表 (app_permissions)
- `id` (主键): 权限记录ID
- `user_type`: 用户类型
- `app_id`: 关联的应用ID

## 👥 用户权限体系

### 企业用户
- 用户管理系统 (port: 3002)
- 数据分析平台 (port: 3003)
- 文档管理中心 (port: 3004)

### 个人用户
- 用户管理系统 (port: 3002)
- 社区论坛 (port: 3005)

### 社区用户
- 社区论坛 (port: 3005)
- 用户管理系统 (port: 3002)

### 机构用户
- 机构认证中心 (port: 3007)
- 数据分析平台 (port: 3003)
- 文档管理中心 (port: 3004)

### 政府用户
- **所有应用完整权限** (ports: 3002-3007)

## 🛡️ 安全特性

- **区块链 DID**: 基于以太坊助记词生成不可篡改的去中心化身份
- **WebAuthn 认证**: 使用 FIDO2 标准的生物识别技术
- **密码加密**: bcrypt 算法（成本因子 14）加密存储用户密码
- **JWT 令牌**: 7天有效期的安全会话管理机制
- **CORS 保护**: 跨域请求安全控制
- **RP ID 验证**: WebAuthn 域名验证防止钓鱼攻击
- **助记词本地化**: 助记词仅在前端生成和存储，后端不保存

## 🔐 WebAuthn 认证流程

本系统实现了完整的 WebAuthn (FIDO2) 认证流程：

### 注册流程
1. 后端生成随机 challenge 和用户凭证选项
2. 前端调用 `navigator.credentials.create()`
3. 用户完成 Touch ID/Face ID 生物识别录入
4. 后端验证 attestationObject 和 clientDataJSON
5. 存储凭证ID和公钥到数据库

### 认证流程  
1. 后端生成认证 challenge 和 allowCredentials
2. 前端调用 `navigator.credentials.get()`
3. 用户完成 Touch ID/Face ID 生物识别验证
4. 后端验证签名、authenticatorData 和 RP ID Hash
5. 返回 7 天有效期 JWT 令牌完成登录

## 🐛 故障排除

### 常见问题

**问题1**: 容器启动失败
```bash
# 解决方案：清理Docker缓存重新构建
docker-compose down -v
docker system prune -f
docker-compose up --build -d
```

**问题2**: 数据库连接失败
```bash
# 检查数据库容器状态
docker logs portal_db
# 重启数据库容器
docker-compose restart db
```

**问题3**: 指纹认证不工作
- 确保使用 HTTPS 或 localhost 环境
- 检查浏览器 WebAuthn 支持（Chrome DevTools > Application > WebAuthn）
- 验证设备具有生物识别功能
- 查看浏览器控制台错误信息

**问题4**: 端口占用冲突
```bash
# 查看端口占用
lsof -i :3001
lsof -i :8080
lsof -i :3306
# 修改 docker-compose.yaml 中的端口映射
```

**问题5**: 助记词恢复失败
- 确保助记词正确（12个单词，空格分隔）
- 检查助记词对应的 DID 是否已注册
- 查看浏览器控制台的 ethers.js 错误信息

## 📄 API 接口

详细的 API 文档请参考 [API_DOCUMENTATION.md](API_DOCUMENTATION.md)

### 核心接口概览

#### 用户管理
- `POST /api/register` - 用户注册（DID + 邮箱 + 密码）
- `POST /api/login/basic` - 基础登录认证（邮箱 + 密码）
- `POST /api/verify-did` - 验证 DID 是否存在（助记词恢复用）
- `POST /api/reset-password` - 通过 DID 重置密码

#### WebAuthn 认证
- `POST /api/webauthn/register/begin` - 开始指纹注册
- `POST /api/webauthn/register/finish` - 完成指纹注册  
- `POST /api/webauthn/login/begin` - 开始指纹认证
- `POST /api/login/verify-webauthn` - 验证指纹并颁发 JWT

#### 应用管理
- `GET /api/apps?user_type=企业` - 根据用户类型获取应用列表

## 🎯 功能特性

### ✅ 已实现
- [x] 基于 ethers.js 的 DID 生成和助记词管理
- [x] 邮箱 + 密码注册和登录
- [x] WebAuthn 生物识别注册和认证
- [x] JWT 7天免登录持久化
- [x] 助记词恢复 DID 和密码重置
- [x] 基于用户类型的权限控制
- [x] 应用列表动态展示
- [x] Docker 容器化部署
- [x] 自动初始化示例数据

### 🚧 待开发
- [ ] 管理员后台（应用管理、用户管理）
- [ ] 多因素认证（MFA）
- [ ] 审计日志记录
- [ ] 密钥轮换机制
- [ ] Redis 缓存 WebAuthn challenge
- [ ] 前端 UI 优化和国际化
- [ ] 单元测试和集成测试
- [ ] CI/CD 流程

## 🤝 贡献指南

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 📜 许可证

本项目采用 MIT 许可证 - 详见 LICENSE 文件

## 📞 联系方式

项目维护者: Cosmos Link Team

项目链接: [https://github.com/cosmos-link/did-login](https://github.com/cosmos-link/did-login)

## 🙏 致谢

- [Ethers.js](https://docs.ethers.io/) - 以太坊 JavaScript 库
- [Gin](https://gin-gonic.com/) - Go Web 框架
- [GORM](https://gorm.io/) - Go ORM 库
- [WebAuthn](https://webauthn.io/) - Web 认证标准
- [Tailwind CSS](https://tailwindcss.com/) - CSS 框架

---

**最后更新时间**: 2026-01-04  
**版本**: v1.0.0  
**状态**: ✅ 生产就绪
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

## 📜 开源协议

本项目采用 MIT 协议，详见 [LICENSE](LICENSE) 文件。

## 🔗 相关链接

- [WebAuthn规范](https://www.w3.org/TR/webauthn-2/)
- [以太坊DID标准](https://github.com/decentralized-identity/ethr-did-resolver)
- [Go Gin框架](https://github.com/gin-gonic/gin)
- [Ethers.js文档](https://docs.ethers.io/v5/)