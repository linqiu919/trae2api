package config

import (
	"fmt"
	"github.com/trae2api/pkg/logger"
	"os"
	"time"
)

type Config struct {
	AppID        string
	ClientID     string
	RefreshToken string
	UserID       string
	BaseURL      string
	IDEVersion   string
	AuthToken    string
}

var AppConfig Config

func InitConfig() error {
	// 从环境变量读取配置
	AppConfig = Config{
		AppID:        getEnv("APP_ID", ""),
		ClientID:     getEnv("CLIENT_ID", ""),
		RefreshToken: getEnv("REFRESH_TOKEN", ""),
		UserID:       getEnv("USER_ID", ""),
		BaseURL:      getEnv("BASE_URL", "https://a0ai-api-sg.byteintlapi.com"),
		IDEVersion:   getEnv("IDE_VERSION", "1.0.4"),
		AuthToken:    getEnv("AUTH_TOKEN", ""),
	}

	// 初始化获取 Token
	if err := RefreshIDEToken(); err != nil {
		return fmt.Errorf("initial token refresh failed: %v", err)
	}

	// 启动定期刷新 Token 的 goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			if err := RefreshIDEToken(); err != nil {
				logger.Log.Errorf("自动刷新 Token 失败: %v", err)
			}
		}
	}()

	logger.Log.Info("配置加载完成\n" +
		"----------------------------------------\n" +
		"AppID:        " + AppConfig.AppID + "\n" +
		"ClientID:     " + AppConfig.ClientID + "\n" +
		"UserID:       " + AppConfig.UserID + "\n" +
		"RefreshToken: " + AppConfig.RefreshToken + "\n" +
		"BaseURL:      " + AppConfig.BaseURL + "\n" +
		"IDEVersion:   " + AppConfig.IDEVersion + "\n" +
		"AuthToken:    " + AppConfig.AuthToken + "\n" +
		"----------------------------------------")

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
