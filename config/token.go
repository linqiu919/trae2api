package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/trae2api/pkg/logger"
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
	refreshToken    string
)

func RefreshIDEToken(baseURL string) error {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	now := time.Now().Unix() * 1000

	// 判断 RefreshToken 是否过期
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

	// 使用内存中的refreshToken（如果存在），否则使用环境变量中的refreshToken
	currentRefreshToken := refreshToken
	if currentRefreshToken == "" {
		currentRefreshToken = os.Getenv("REFRESH_TOKEN")
	}

	// 请求新的 Refresh Token
	refreshConfig := TokenConfig{
		ClientID:     os.Getenv("CLIENT_ID"),
		RefreshToken: currentRefreshToken,
		ClientSecret: "-",
		UserID:       os.Getenv("USER_ID"),
	}

	jsonData, err := json.Marshal(refreshConfig)
	if err != nil {
		return fmt.Errorf("marshal refresh config failed: %v", err)
	}

	logger.Log.Info("开始执行RefreshToken获取......")

	resp, err := http.Post(baseURL+"/cloudide/api/v3/trae/oauth/ExchangeToken",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("refresh token request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error(fmt.Sprintf("请求失败\n状态码: %d\n响应内容:\n%s", resp.StatusCode, string(respBody)))
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var refreshResp TokenResponse
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&refreshResp); err != nil {
		return fmt.Errorf("decode refresh response failed: %v", err)
	}

	// 将新的refreshToken保存到内存中
	refreshToken = refreshResp.Result.RefreshToken

	logger.Log.Info("获取到新的 RefreshToken: " + refreshToken + "\n")

	// 使用新的 RefreshToken 刷新 Token
	tokenConfig := TokenConfig{
		ClientID:     os.Getenv("CLIENT_ID"),
		RefreshToken: refreshToken,
		ClientSecret: "-",
		UserID:       os.Getenv("USER_ID"),
	}

	jsonData, err = json.Marshal(tokenConfig)
	if err != nil {
		return fmt.Errorf("marshal token config failed: %v", err)
	}

	logger.Log.Info("开始执行Token获取......")

	resp, err = http.Post(
		"https://api-sg-central.trae.ai/cloudide/api/v3/trae/oauth/ExchangeToken",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("refresh token request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error(fmt.Sprintf("请求失败\n状态码: %d\n响应内容:\n%s", resp.StatusCode, string(respBody)))
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

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
		"Token: " + currentToken + "\n" +
		"Token 有效期至: " + time.UnixMilli(tokenExpireAt).Format("2006-01-02 15:04:05") + "\n" +
		"RefreshToken: " + refreshToken + "\n" +
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
