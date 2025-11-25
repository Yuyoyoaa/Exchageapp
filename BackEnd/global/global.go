package global

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 全局变量
var (
	Db      *gorm.DB
	RedisDB *redis.Client
)
