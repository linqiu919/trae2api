package config

import (
	"github.com/trae2api/pkg/logger"
	"os"
)

type Config struct {
	AppID      string
	IDEToken   string
	BaseURL    string
	IDEVersion string
	AuthToken  string
}

var AppConfig Config

func InitConfig() error {
	// 从环境变量读取配置
	AppConfig = Config{
		AppID:      getEnv("APP_ID", ""),
		IDEToken:   getEnv("IDE_TOKEN", ""),
		BaseURL:    getEnv("BASE_URL", "https://a0ai-api-sg.byteintlapi.com"),
		IDEVersion: getEnv("IDE_VERSION", "1.0.2"),
		AuthToken:  getEnv("AUTH_TOKEN", ""),
	}

	logger.Log.WithFields(map[string]interface{}{
		"AppID":     AppConfig.AppID,
		"AuthToken": AppConfig.AuthToken,
	}).Info("配置加载完成")

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
