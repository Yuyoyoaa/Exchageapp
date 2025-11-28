package models

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID        uint `gorm:"primarykey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Preview    string `json:"preview" binding:"required"`
	Cover      string `json:"cover,omitempty"`
	LikesCount int64  `gorm:"default:0" json:"likesCount"`
	ViewsCount int64  `gorm:"default:0" json:"viewsCount"`
	AuthorID   uint   `json:"authorId"`             // 创建者（通常为 admin）
	CategoryID uint   `json:"categoryId,omitempty"` // 分类
	Status     string `gorm:"default:'published'" json:"status"`
}
