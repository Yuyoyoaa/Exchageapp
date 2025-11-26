package config

import (
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() {
	dsn := AppConfig.Database.Dsn
	fmt.Println("DSN:", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize database, got error: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to configure database, got error: %v", err)
	}

	sqlDB.SetMaxIdleConns(AppConfig.Database.MaxIdleCONNS)
	sqlDB.SetMaxOpenConns(AppConfig.Database.MaxOpenCONNS)
	sqlDB.SetConnMaxLifetime(time.Hour)

	global.Db = db

	// 迁移模型
	if err := global.Db.AutoMigrate(
		&models.User{},
		&models.Article{},
		&models.Category{},
		&models.ArticleLike{},
		&models.Favorite{},
		&models.Comment{},
		&models.ExchangeRate{},
	); err != nil {
		log.Fatalf("Failed to migrate database, got error: %v", err)
	}

	// 为 likes 添加唯一索引（兼容不同 gorm 版本的写法，若报错可删）：
	_ = global.Db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_article ON article_likes (user_id, article_id);")

	fmt.Println("Database initialized and migrated successfully")
}
