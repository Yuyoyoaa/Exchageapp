package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	ArticleID uint   `json:"articleId" gorm:"index;not null"`
	UserID    uint   `json:"userId" gorm:"index;not null"`
	UserName  string `json:"userName"`
	Content   string `json:"content" binding:"required"`
	ParentID  *uint  `json:"parentId,omitempty"` // 支持二级回复
}
