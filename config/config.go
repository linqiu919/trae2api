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
}

var RefreshTokenCacheEnabled = getEnv("REFRESH_TOKEN_CACHE_ENABLED", "false")
var AutoContinueEnabled = getEnv("AUTO_CONTINUE_ENABLED", "false")

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
		BaseURL:         getEnv("BASE_URL", "https://trae-api-sg.mchost.guru"),
		RefreshTokenURL: getEnv("REFRESH_TOKEN_URL", "https://api-sg-central.trae.ai"),
		GetFileIDURL:    getEnv("GET_FILE_ID_URL", "https://imagex-ap-singapore-1.bytevcloudapi.com"),
		UploadFileURL:   getEnv("UPLOAD_FILE_URL", "https://tos-sg16-share.vodupload.com"),
		// 非必填配置
		AuthToken: getEnv("AUTH_TOKEN", ""),
	}

	// redis连接字符串 示例: redis://default:pwd@localhost:6379
	if RefreshTokenCacheEnabled == "true" {
		if RedisConnString == "" {
			logger.Log.Fatalln("未配置环境变量 REDIS_CONN_STRING")
		}
	}

	// 打印是否开启claude3.7自动发起继续对话
	logger.Log.Info("当前是否开启claude3.7自动继续请求: " + AutoContinueEnabled)

	// 是否为开发调试模式
	codingMode := os.Getenv("CODING_MODE") == "true"
	codingToken := os.Getenv("CODING_TOKEN")

	refreshTokenUrl := AppConfig.RefreshTokenURL
	// 初始化获取 Token
	if err := RefreshIDEToken(refreshTokenUrl, codingMode, codingToken); err != nil {
		return fmt.Errorf("initial token refresh failed: %v", err)
	}

	// 启动定期刷新 Token 的 goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			if err := RefreshIDEToken(refreshTokenUrl, codingMode, codingToken); err != nil {
				logger.Log.Errorf("自动刷新 Token 失败: %v", err)
			}
		}
	}()

	logger.Log.Info("Trae2Api配置加载完成:\n" +
		"----------------------------------------\n" +
		"AppID:        " + AppConfig.AppID + "\n" +
		"ClientID:     " + AppConfig.ClientID + "\n" +
		"UserID:       " + AppConfig.UserID + "\n" +
		"RefreshToken: " + AppConfig.RefreshToken + "\n" +
		"BaseURL:      " + AppConfig.BaseURL + "\n" +
		"RefreshTokenURL: " + AppConfig.RefreshTokenURL + "\n" +
		"GetFileIDURL: " + AppConfig.GetFileIDURL + "\n" +
		"UploadFileURL: " + AppConfig.UploadFileURL + "\n" +
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
