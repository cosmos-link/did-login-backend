package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var jwtKey = []byte("your_secret_key_2026") // 实际生产请用环境变量

// 存储WebAuthn挑战的临时map (生产环境应该使用Redis)
var challenges = make(map[string]string)

// JWT 载荷
type Claims struct {
	DID      string `json:"did"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

// 生成随机挑战 - 返回base64url格式（无padding）
func generateChallenge() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	// base64url编码（移除padding）
	encoded := base64.URLEncoding.EncodeToString(bytes)
	return strings.TrimRight(encoded, "=")
}

// 验证WebAuthn注册
func verifyRegistration(clientDataJSON, attestationObject string, expectedChallenge string) error {
	// 解析clientDataJSON - 处理base64url格式（可能缺少padding）
	// 添加padding以确保能正确解码
	decodedStr := clientDataJSON
	switch len(clientDataJSON) % 4 {
	case 2:
		decodedStr += "=="
	case 3:
		decodedStr += "="
	}

	clientDataBytes, err := base64.URLEncoding.DecodeString(decodedStr)
	if err != nil {
		// 如果URLEncoding失败，尝试StdEncoding
		clientDataBytes, err = base64.StdEncoding.DecodeString(decodedStr)
		if err != nil {
			return fmt.Errorf("解析clientDataJSON失败: %v", err)
		}
	}

	var clientData struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Origin    string `json:"origin"`
	}

	if err := json.Unmarshal(clientDataBytes, &clientData); err != nil {
		return fmt.Errorf("解析clientData失败: %v", err)
	}

	// 验证类型
	if clientData.Type != "webauthn.create" {
		return fmt.Errorf("无效的类型: %s", clientData.Type)
	}

	// 验证挑战
	if clientData.Challenge != expectedChallenge {
		fmt.Printf("【注册】Challenge不匹配! 从clientData获得: %s (长度%d), 预期: %s (长度%d)\n", clientData.Challenge, len(clientData.Challenge), expectedChallenge, len(expectedChallenge))
		return fmt.Errorf("挑战不匹配")
	}
	fmt.Printf("【注册】Challenge验证成功 ✓\n")

	// 验证来源 - 支持localhost和生产环境
	if clientData.Origin == "" {
		return fmt.Errorf("来源不能为空")
	}
	// 基本的来源格式验证，允许http/https协议
	if !strings.HasPrefix(clientData.Origin, "http://") && !strings.HasPrefix(clientData.Origin, "https://") {
		return fmt.Errorf("无效的来源协议: %s", clientData.Origin)
	}

	return nil
}

// 验证WebAuthn认证
func verifyAuthentication(clientDataJSON, authenticatorData, signature string, expectedChallenge string) error {
	// 解析clientDataJSON - 处理base64url格式（可能缺少padding）
	decodedStr := clientDataJSON
	switch len(clientDataJSON) % 4 {
	case 2:
		decodedStr += "=="
	case 3:
		decodedStr += "="
	}

	clientDataBytes, err := base64.URLEncoding.DecodeString(decodedStr)
	if err != nil {
		// 如果URLEncoding失败，尝试StdEncoding
		clientDataBytes, err = base64.StdEncoding.DecodeString(decodedStr)
		if err != nil {
			return fmt.Errorf("解析clientDataJSON失败: %v", err)
		}
	}

	var clientData struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Origin    string `json:"origin"`
	}

	if err := json.Unmarshal(clientDataBytes, &clientData); err != nil {
		return fmt.Errorf("解析clientData失败: %v", err)
	}

	// 验证类型
	if clientData.Type != "webauthn.get" {
		return fmt.Errorf("无效的类型: %s", clientData.Type)
	}

	// 验证挑战
	if clientData.Challenge != expectedChallenge {
		return fmt.Errorf("挑战不匹配")
	}

	// 验证来源 - 支持localhost和生产环境
	if clientData.Origin == "" {
		return fmt.Errorf("来源不能为空")
	}
	// 基本的来源格式验证，允许http/https协议
	if !strings.HasPrefix(clientData.Origin, "http://") && !strings.HasPrefix(clientData.Origin, "https://") {
		return fmt.Errorf("无效的来源协议: %s", clientData.Origin)
	}

	// 验证authenticatorData
	authDataBytes, err := base64.URLEncoding.DecodeString(authenticatorData)
	if err != nil {
		return fmt.Errorf("解析authenticatorData失败: %v", err)
	}

	if len(authDataBytes) < 37 {
		return fmt.Errorf("authenticatorData太短")
	}

	// 动态验证RP ID Hash - 支持localhost和生产环境
	rpIdHash := authDataBytes[0:32]

	// 从Origin提取RP ID
	rpId := extractRpIdFromOrigin(clientData.Origin)
	expectedRpIdHash := sha256.Sum256([]byte(rpId))

	// 比较RP ID Hash
	for i := 0; i < 32; i++ {
		if rpIdHash[i] != expectedRpIdHash[i] {
			return fmt.Errorf("RP ID Hash不匹配: 期望=%s, 实际Origin=%s", rpId, clientData.Origin)
		}
	}

	// 检查用户存在标志位
	flags := authDataBytes[32]
	if flags&0x01 == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// 从Origin URL中提取RP ID
func extractRpIdFromOrigin(origin string) string {
	// 移除协议前缀
	if strings.HasPrefix(origin, "https://") {
		origin = origin[8:]
	} else if strings.HasPrefix(origin, "http://") {
		origin = origin[7:]
	}

	// 移除端口号
	if strings.Contains(origin, ":") {
		origin = strings.Split(origin, ":")[0]
	}

	// 移除路径
	if strings.Contains(origin, "/") {
		origin = strings.Split(origin, "/")[0]
	}

	return origin
}

// safeMigrate 安全的数据库迁移函数
func safeMigrate(db *gorm.DB) error {
	// 要迁移的模型列表
	models := []interface{}{&User{}, &Application{}, &AppPermission{}}

	for _, model := range models {
		// 获取表名
		typeName := fmt.Sprintf("%T", model)
		var tableName string
		switch typeName {
		case "*main.User":
			tableName = "users"
		case "*main.Application":
			tableName = "applications"
		case "*main.AppPermission":
			tableName = "app_permissions"
		default:
			tableName = "unknown"
		}

		fmt.Printf("Migrating table: %s\n", tableName)

		// 检查表是否存在
		var count int64
		err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count).Error

		if err != nil {
			// 如果查询出错（可能是权限问题），直接尝试迁移
			fmt.Printf("Warning: Failed to check table existence for %s: %v, attempting migration anyway\n", tableName, err)
			if err := db.AutoMigrate(model); err != nil {
				return fmt.Errorf("failed to migrate %s: %v", tableName, err)
			}
			continue
		}

		if count > 0 {
			// 表已存在，检查是否有主键冲突
			fmt.Printf("Table %s already exists, checking for primary key conflicts\n", tableName)

			// 对于已存在的表，尝试删除可能存在的主键约束（避免 Multiple primary key defined 错误）
			// 注意：这仅在开发环境使用，生产环境需要更谨慎的处理
			if tableName == "applications" {
				// 检查是否存在主键约束
				var pkCount int64
				err := db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = ? AND constraint_type = 'PRIMARY KEY'", tableName).Scan(&pkCount).Error
				if err == nil && pkCount > 0 {
					fmt.Printf("Found existing primary key on %s, skipping AutoMigrate to avoid conflict\n", tableName)
					continue
				}
			}
		}

		// 执行迁移
		if err := db.AutoMigrate(model); err != nil {
			// 如果是主键冲突错误，记录警告但继续
			if strings.Contains(err.Error(), "Multiple primary key") || strings.Contains(err.Error(), "1068") {
				fmt.Printf("Warning: Primary key conflict on %s (table may already have correct structure): %v\n", tableName, err)
				continue
			}
			return fmt.Errorf("failed to migrate %s: %v", tableName, err)
		}

		fmt.Printf("✅ Successfully migrated %s\n", tableName)
	}

	return nil
}

func initDB() {
	// 从环境变量读取数据库配置
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "47.84.96.59" // 默认阿里云MySQL地址
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3308" // 默认端口
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "ykt123456" // 默认密码
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ykt_db"
	}

	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	fmt.Printf("Connecting to database: %s@%s:%s/%s\n", dbUser, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// 安全的数据库迁移
	fmt.Println("Starting database migration...")
	if err := safeMigrate(db); err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	DB = db
	fmt.Println("✅ Database migration completed")

	// 初始化示例数据
	initSeedData(db)
}

// 初始化示例数据
func initSeedData(db *gorm.DB) {
	// 检查 applications 表是否存在
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", "applications").Scan(&count).Error
	if err != nil || count == 0 {
		fmt.Println("Applications table not ready, skipping seed data initialization")
		return
	}

	// 检查是否已经有数据
	var appCount int64
	db.Model(&Application{}).Count(&appCount)
	if appCount > 0 {
		fmt.Println("数据库已有数据，跳过初始化")
		return
	}

	// 等待服务器启动完成
	go func() {
		time.Sleep(3 * time.Second) // 等待3秒确保服务器完全启动

		// 调用API创建示例应用
		appsData := []map[string]interface{}{
			{
				"name":           "用户管理系统",
				"container_name": "user-management",
				"port":           3002,
				"base_url":       "http://localhost",
				"description":    "管理企业和个人用户",
				"user_types":     []string{"企业", "个人", "社区", "政府"},
			},
			{
				"name":           "数据分析平台",
				"container_name": "data-analytics",
				"port":           3003,
				"base_url":       "http://localhost",
				"description":    "数据可视化和分析工具",
				"user_types":     []string{"企业", "机构", "政府"},
			},
			{
				"name":           "文档管理中心",
				"container_name": "doc-center",
				"port":           3004,
				"base_url":       "http://localhost",
				"description":    "企业文档存储和共享",
				"user_types":     []string{"企业", "机构", "政府"},
			},
			{
				"name":           "社区论坛",
				"container_name": "community-forum",
				"port":           3005,
				"base_url":       "http://localhost",
				"description":    "社区成员交流平台",
				"user_types":     []string{"个人", "社区", "政府"},
			},
			{
				"name":           "政务服务大厅",
				"container_name": "gov-services",
				"port":           3006,
				"base_url":       "http://localhost",
				"description":    "政府服务在线办理",
				"user_types":     []string{"政府"},
			},
			{
				"name":           "机构认证中心",
				"container_name": "org-auth",
				"port":           3007,
				"base_url":       "http://localhost",
				"description":    "机构资质认证",
				"user_types":     []string{"机构", "政府"},
			},
		}

		createdCount := 0
		for _, appData := range appsData {
			if createAppViaAPI(appData) {
				createdCount++
			}
		}

		fmt.Printf("✓ 通过API创建了 %d 个示例应用\n", createdCount)
	}()
}

// 通过API创建应用
func createAppViaAPI(appData map[string]interface{}) bool {
	jsonData, _ := json.Marshal(appData)

	resp, err := http.Post("http://localhost:60208/api/apps", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		fmt.Printf("创建应用失败 %s: %v\n", appData["name"], err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✓ 成功创建应用: %s\n", appData["name"])
		return true
	} else {
		fmt.Printf("创建应用失败 %s: HTTP %d\n", appData["name"], resp.StatusCode)
		return false
	}
}

// 生成 JWT (有效期 7 天)
func generateToken(did string, userType string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		DID:      did,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func main() {
	initDB()
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 1. 注册接口
	r.POST("/api/register", func(c *gin.Context) {
		var input struct {
			DID      string `json:"did"`
			Email    string `json:"email"`
			Password string `json:"password"`
			UserType string `json:"user_type"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
		user := User{
			DID:          input.DID,
			Email:        input.Email,
			PasswordHash: string(hash),
			UserType:     input.UserType,
		}

		if err := DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
	})

	// 1.5. WebAuthn注册选项生成
	r.POST("/api/webauthn/register/begin", func(c *gin.Context) {
		var input struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		challenge := generateChallenge()
		challenges[input.Email] = challenge
		fmt.Printf("【注册Begin】为用户 %s 生成challenge: %s (长度%d)\n", input.Email, challenge, len(challenge))

		// 动态获取RP ID，支持localhost和生产环境
		rpId := c.Request.Host
		// 如果是端口地址，只取域名部分
		if strings.Contains(rpId, ":") {
			rpId = strings.Split(rpId, ":")[0]
		}

		options := gin.H{
			"challenge": challenge,
			"rp": gin.H{
				"name": "DID Portal",
				"id":   rpId,
			},
			"user": gin.H{
				"id":          base64.URLEncoding.EncodeToString([]byte(input.Email)),
				"name":        input.Email,
				"displayName": input.Email,
			},
			"pubKeyCredParams": []gin.H{
				{"type": "public-key", "alg": -7}, // ES256
			},
			"authenticatorSelection": gin.H{
				"authenticatorAttachment": "platform",
				"userVerification":        "required",
			},
			"timeout": 60000,
		}

		c.JSON(http.StatusOK, options)
	})

	// 1.6. WebAuthn注册完成
	r.POST("/api/webauthn/register/finish", func(c *gin.Context) {
		var input struct {
			Email      string `json:"email"`
			Credential gin.H  `json:"credential"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 验证挑战
		expectedChallenge, exists := challenges[input.Email]
		if !exists {
			// 打印当前所有存储的challenge keys
			var keys []string
			for k := range challenges {
				keys = append(keys, k)
			}
			fmt.Printf("【注册Finish】错误: 没有为邮箱 %s 找到challenge。当前存储的keys: %v\n", input.Email, keys)
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的挑战"})
			return
		}
		fmt.Printf("【注册Finish】为邮箱 %s 找到存储的challenge: %s (长度%d)\n", input.Email, expectedChallenge, len(expectedChallenge))
		delete(challenges, input.Email)

		// 真正的WebAuthn验证
		clientDataJSON := input.Credential["response"].(map[string]interface{})["clientDataJSON"].(string)
		attestationObject := input.Credential["response"].(map[string]interface{})["attestationObject"].(string)

		if err := verifyRegistration(clientDataJSON, attestationObject, expectedChallenge); err != nil {
			fmt.Printf("WebAuthn注册验证失败: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证失败: " + err.Error()})
			return
		}

		fmt.Printf("WebAuthn注册验证成功\n")

		credentialId := input.Credential["id"].(string)

		// 更新用户的WebAuthn信息
		var user User
		if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}

		// 存储凭证ID (简化处理)
		credIdBytes, _ := base64.URLEncoding.DecodeString(credentialId)
		user.CredentialID = credIdBytes
		DB.Save(&user)

		c.JSON(http.StatusOK, gin.H{"verified": true})
	})

	// 2. 登录接口 (第一阶段：Email+密码)
	r.POST("/api/login/basic", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 添加调试日志
		fmt.Printf("接收到的登录请求: Email=%s, Password长度=%d\n", input.Email, len(input.Password))

		var user User
		if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			fmt.Printf("用户查找失败: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
			return
		}

		// 生成JWT令牌
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": user.Email,
			"did":   user.DID,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, err := token.SignedString([]byte("your-secret-key"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
			return
		}

		// 返回完整的用户信息，可用于备用认证
		c.JSON(http.StatusOK, gin.H{
			"message":   "基础验证通过",
			"token":     tokenString,
			"did":       user.DID,
			"user_type": user.UserType,
			"email":     user.Email,
		})
	})

	// 2.5. WebAuthn登录选项生成
	r.POST("/api/webauthn/login/begin", func(c *gin.Context) {
		var input struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 获取用户的凭证信息
		var user User
		if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}

		if len(user.CredentialID) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户未注册指纹"})
			return
		}

		challenge := generateChallenge()
		challenges[input.Email] = challenge

		// 动态获取RP ID，支持localhost和生产环境
		rpId := c.Request.Host
		// 如果是端口地址，只取域名部分
		if strings.Contains(rpId, ":") {
			rpId = strings.Split(rpId, ":")[0]
		}

		options := gin.H{
			"challenge": challenge,
			"timeout":   60000,
			"rpId":      rpId,
			"allowCredentials": []gin.H{
				{
					"type": "public-key",
					"id":   base64.URLEncoding.EncodeToString(user.CredentialID),
				},
			},
			"userVerification": "required",
		}

		c.JSON(http.StatusOK, options)
	})

	// 3. WebAuthn 验证完成并下发 7 天 JWT
	r.POST("/api/login/verify-webauthn", func(c *gin.Context) {
		var input struct {
			Email      string `json:"email"`
			Credential gin.H  `json:"credential"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Printf("WebAuthn验证请求: Email=%s\n", input.Email)

		// 验证挑战
		expectedChallenge, exists := challenges[input.Email]
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的挑战"})
			return
		}
		delete(challenges, input.Email)

		// 真正的WebAuthn验证
		clientDataJSON := input.Credential["response"].(map[string]interface{})["clientDataJSON"].(string)
		authenticatorData := input.Credential["response"].(map[string]interface{})["authenticatorData"].(string)
		signature := input.Credential["response"].(map[string]interface{})["signature"].(string)

		if err := verifyAuthentication(clientDataJSON, authenticatorData, signature, expectedChallenge); err != nil {
			fmt.Printf("WebAuthn认证验证失败: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "验证失败: " + err.Error()})
			return
		}

		fmt.Printf("WebAuthn认证验证成功\n")

		var user User
		if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			return
		}

		token, _ := generateToken(user.DID, user.UserType)
		c.JSON(http.StatusOK, gin.H{
			"token":     token,
			"user_type": user.UserType,
			"did":       user.DID,
		})
	})

	// 4. 获取 App 列表
	r.GET("/api/apps", func(c *gin.Context) {
		// 从查询参数获取 userType
		userType := c.Query("user_type")
		if userType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_type 参数缺失"})
			return
		}

		var apps []Application
		err := DB.Table("applications").
			Select("applications.*").
			Joins("JOIN app_permissions ON app_permissions.app_id = applications.app_id").
			Where("app_permissions.user_type = ?", userType).
			Find(&apps).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}

		c.JSON(http.StatusOK, apps)
	})

	// 4.1. 添加 App
	r.POST("/api/apps", func(c *gin.Context) {
		var input struct {
			Name          string   `json:"name"`
			ContainerName string   `json:"container_name"`
			Port          int      `json:"port"`
			BaseURL       string   `json:"base_url"`
			Description   string   `json:"description"`
			UserTypes     []string `json:"user_types"` // 允许访问的用户类型列表
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 创建应用
		app := Application{
			Name:          input.Name,
			ContainerName: input.ContainerName,
			Port:          input.Port,
			BaseURL:       input.BaseURL,
			Description:   input.Description,
		}

		if err := DB.Create(&app).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建应用失败"})
			return
		}

		// 创建权限映射
		for _, userType := range input.UserTypes {
			permission := AppPermission{
				UserType: userType,
				AppID:    app.AppID,
			}
			if err := DB.Create(&permission).Error; err != nil {
				fmt.Printf("创建权限失败: %v\n", err)
				// 继续创建其他权限，不中断流程
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "应用创建成功",
			"app_id":  app.AppID,
			"app":     app,
		})
	})

	// 4.2. 删除 App
	r.DELETE("/api/apps/:id", func(c *gin.Context) {
		appID := c.Param("id")

		// 先删除相关权限
		if err := DB.Where("app_id = ?", appID).Delete(&AppPermission{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除应用权限失败"})
			return
		}

		// 删除应用
		if err := DB.Where("app_id = ?", appID).Delete(&Application{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除应用失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "应用删除成功"})
	})

	// 5. DID 验证接口（用于助记词恢复）
	r.POST("/api/verify-did", func(c *gin.Context) {
		var input struct {
			DID string `json:"did"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user User
		if err := DB.Where("did = ?", input.DID).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "DID 不存在", "exists": false})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"exists":    true,
			"email":     user.Email,
			"user_type": user.UserType,
		})
	})

	// 6. 密码重置接口（通过 DID）
	r.POST("/api/reset-password", func(c *gin.Context) {
		var input struct {
			DID         string `json:"did"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user User
		if err := DB.Where("did = ?", input.DID).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "DID 不存在"})
			return
		}

		// 生成新密码哈希
		hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 14)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}

		// 更新密码
		user.PasswordHash = string(hash)
		if err := DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码更新失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "密码重置成功",
			"email":   user.Email,
			"did":     user.DID,
		})
	})

	r.Run(":60208")
}
