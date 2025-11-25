package config

import (
	"log"

	"github.com/spf13/viper"
)

// 将配置文件赋值给自定义的结构体
type Config struct {
	App struct {
		Name string `yaml:"name"`
		Port string `yaml:"port"`
	} `yaml:"app"`
	Database struct {
		Dsn          string `yaml:"dsn"`
		MaxIdleCONNS int    `yaml:"MaxIdleCONNS"`
		MaxOpenCONNS int    `yaml:"MaxOpenCONNS"`
	}
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

var AppConfig *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error Reading Config file: %v", err)
	}

	AppConfig = &Config{}

	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	InitDB()
	InitRedis()
}
