package models

import (
	"time"

	"gorm.io/gorm"
)

// 收藏记录
type Favorite struct {
	ID        uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID    uint    `gorm:"index;not null"`
	ArticleID uint    `gorm:"index;not null"`
	Article   Article `gorm:"foreignKey:ArticleID;references:ID" json:"Article"`
}
