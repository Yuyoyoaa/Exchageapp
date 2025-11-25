package models

import "gorm.io/gorm"

type User struct {
	gorm.Model        // 嵌入gorm预定义结构体
	Username   string `gorm:"unique"` // 标签用户名不重复
	Password   string
}
