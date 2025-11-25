package models

import "gorm.io/gorm"

// 在mysql表格中都会变为小写
type Article struct {
	gorm.Model
	Title   string `binding:"required"`
	Content string `binding:"required"`
	Preview string `binding:"required"` // 文字的预览
	Likes   int    `gorm:"default:0"`
}
