package models

import "gorm.io/gorm"

// 点赞记录，确保每用户对每文章唯一
type ArticleLike struct {
	gorm.Model
	UserID    uint `gorm:"index;not null"`
	ArticleID uint `gorm:"index;not null"`
}

// 在迁移或 DB 初始化后，为 (user_id, article_id) 添加唯一索引：
// db.Model(&models.ArticleLike{}).AddUniqueIndex("idx_user_article", "user_id", "article_id")
