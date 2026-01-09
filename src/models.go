package main

import (
	"time"
)

// User 用户表 - DID 作为主键
type User struct {
	DID          string    `gorm:"primaryKey;column:did;size:100"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	UserType     string    `gorm:"type:varchar(20);not null;index"` // 企业, 个人, 社区, 机构, 政府
	
	// WebAuthn 指纹相关字段
	CredentialID []byte    `gorm:"type:blob"`              // 凭证ID（base64编码后的数据）
	PublicKey    []byte    `gorm:"type:blob"`              // 指纹公钥
	SignCount    uint32    `gorm:"default:0"`               // 签名计数器(防重放攻击)
	
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Application App信息表
type Application struct {
	AppID         uint      `gorm:"primaryKey;autoIncrement"`
	Name          string    `gorm:"type:varchar(100);not null"`
	ContainerName string    `gorm:"type:varchar(100);not null"`
	Port          int       `gorm:"not null"`
	BaseURL       string    `gorm:"type:varchar(255)"`
	Description   string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Application) TableName() string {
	return "applications"
}

// AppPermission 权限映射表
type AppPermission struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserType  string    `gorm:"type:varchar(20);not null;index"` // 关联 User.UserType
	AppID     uint      `gorm:"not null;index"`                  // 关联 Application.AppID
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (AppPermission) TableName() string {
	return "app_permissions"
}
