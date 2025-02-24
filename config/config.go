package config

import (
	"fmt"
	"os"
	"time"

	"github.com/trae2api/pkg/logger"
)

type Config struct {
	AppID           string
	ClientID        string
	RefreshToken    string
	UserID          string
	BaseURL         string
	RefreshTokenURL string
	GetFileIDURL    string
	UploadFileURL   string
	AuthToken       string
	IDEVersion      string
}

var AppConfig Config

func InitConfig() error {
	// 从环境变量读取配置
	AppConfig = Config{
		// 必要配置
		AppID:        getEnv("APP_ID", ""),
		ClientID:     getEnv("CLIENT_ID", ""),
		RefreshToken: getEnv("REFRESH_TOKEN", ""),
		UserID:       getEnv("USER_ID", ""),
		// 可选配置
		BaseURL:         getEnv("BASE_URL", "https://a0ai-api-sg.byteintlapi.com"),
		RefreshTokenURL: getEnv("REFRESH_TOKEN_URL", "https://api-sg-central.trae.ai"),
		GetFileIDURL:    getEnv("GET_FILE_ID_URL", "https://imagex-ap-singapore-1.bytevcloudapi.com"),
		UploadFileURL:   getEnv("UPLOAD_FILE_URL", "https://tos-sg16-share.vodupload.com"),
		// 非必填配置
		AuthToken:  getEnv("AUTH_TOKEN", ""),
		IDEVersion: getEnv("IDE_VERSION", "1.0.4"),
	}

	// 初始化获取 Token
	if err := RefreshIDEToken(AppConfig.RefreshTokenURL); err != nil {
		return fmt.Errorf("initial token refresh failed: %v", err)
	}

	// 启动定期刷新 Token 的 goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			if err := RefreshIDEToken(AppConfig.RefreshTokenURL); err != nil {
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
		"RefreshTokenURL: " + AppConfig.RefreshTokenURL + "\n" +
		"GetFileIDURL: " + AppConfig.GetFileIDURL + "\n" +
		"UploadFileURL: " + AppConfig.UploadFileURL + "\n" +
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
