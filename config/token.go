package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/trae2api/pkg/logger"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type TokenConfig struct {
	ClientID     string `json:"ClientID"`
	RefreshToken string `json:"RefreshToken"`
	ClientSecret string `json:"ClientSecret"`
	UserID       string `json:"UserID"`
}

type TokenResponse struct {
	Result struct {
		Token           string `json:"Token"`
		TokenExpireAt   int64  `json:"TokenExpireAt"`
		RefreshToken    string `json:"RefreshToken"`
		RefreshExpireAt int64  `json:"RefreshExpireAt"`
	} `json:"Result"`
}

var (
	tokenMutex      sync.RWMutex
	currentToken    string
	tokenExpireAt   int64
	refreshExpireAt int64
)

func RefreshIDEToken() error {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	now := time.Now().Unix() * 1000

	// 如果已经有了 refreshExpireAt，先检查 RefreshToken 是否过期
	if refreshExpireAt > 0 && now >= refreshExpireAt {
		logger.Log.Error("RefreshToken 已过期，请更新环境变量中的 REFRESH_TOKEN\n" +
			"----------------------------------------\n" +
			"当前时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
			"过期时间: " + time.Unix(refreshExpireAt/1000, 0).Format("2006-01-02 15:04:05") + "\n" +
			"----------------------------------------")
		return nil
	}

	// 检查是否需要刷新 Token（提前5分钟刷新）
	if now < tokenExpireAt-300000 {
		return nil
	}

	tokenConfig := TokenConfig{
		ClientID:     os.Getenv("CLIENT_ID"),
		RefreshToken: os.Getenv("REFRESH_TOKEN"),
		ClientSecret: "-",
		UserID:       os.Getenv("USER_ID"),
	}

	jsonData, err := json.Marshal(tokenConfig)
	if err != nil {
		return fmt.Errorf("marshal token config failed: %v", err)
	}

	// 打印请求参数
	logger.Log.Info("开始执行Token获取......")
	//logger.Log.Info(fmt.Sprintf("请求参数:\n%s", string(jsonData)))

	resp, err := http.Post(
		"https://api-sg-central.trae.ai/cloudide/api/v3/trae/oauth/ExchangeToken",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("refresh token request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %v", err)
	}

	// 只在响应码不是 200 时打印响应内容
	if resp.StatusCode != http.StatusOK {
		logger.Log.Error(fmt.Sprintf("请求失败\n状态码: %d\n响应内容:\n%s", resp.StatusCode, string(respBody)))
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// 重新创建一个新的 Reader 给 json.Decoder 使用
	var tokenResp TokenResponse
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decode token response failed: %v", err)
	}
	logger.Log.Info("请求远端成功,即将更新Token")

	currentToken = tokenResp.Result.Token
	tokenExpireAt = tokenResp.Result.TokenExpireAt
	refreshExpireAt = tokenResp.Result.RefreshExpireAt

	logger.Log.Info("Token 获取成功\n" +
		"----------------------------------------\n" +
		"当前时间: " + time.Now().Format("2006-01-02 15:04:05") + "\n" +
		"Token 有效期至: " + time.UnixMilli(tokenExpireAt).Format("2006-01-02 15:04:05") + "\n" +
		"RefreshToken 有效期至: " + time.UnixMilli(refreshExpireAt).Format("2006-01-02 15:04:05") + "\n" +
		"----------------------------------------")

	return nil
}

func GetCurrentToken() string {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()
	return currentToken
}

func IsRefreshTokenExpired() bool {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()
	return refreshExpireAt > 0 && time.Now().Unix()*1000 >= refreshExpireAt
}
