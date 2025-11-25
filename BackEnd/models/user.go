package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique" json:"username" binding:"required"`
	Password string `json:"-" binding:"required"`
	Role     string `gorm:"default:'user'" json:"role"` // user/admin
	Nickname string `json:"nickname"`                   // 昵称
	Email    string `gorm:"unique" json:"email"`        // 可为空，不强制验证
	Avatar   string `json:"avatar"`                     // 头像
}

/*Role 默认是普通用户 user，管理员为 admin

Password 不在 JSON 输出中显示（json:"-"）

Nickname, Email, Avatar 可以为空 */
