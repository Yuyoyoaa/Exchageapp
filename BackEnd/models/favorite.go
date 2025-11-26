package models

import "gorm.io/gorm"

// 收藏记录
type Favorite struct {
	gorm.Model
	UserID    uint `gorm:"index;not null"`
	ArticleID uint `gorm:"index;not null"`
}
