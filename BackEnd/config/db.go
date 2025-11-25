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
	if err := global.Db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to migrate database, got error: %v", err)
	}

	fmt.Println("Database initialized and migrated successfully")
}
