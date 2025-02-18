package main

import (
	"github.com/gin-gonic/gin"
	"github.com/trae2api/api"
	"github.com/trae2api/config"
	"github.com/trae2api/pkg/logger"
)

func main() {
	// 初始化日志
	logger.Init()

	// 初始化配置
	if err := config.InitConfig(); err != nil {
		logger.Log.Fatalf("Trae2API Config Init Failed: %v", err)
	}

	r := gin.Default()

	// 添加鉴权中间件
	r.Use(api.AuthMiddleware())

	// OpenAI 格式的 API 路由
	r.GET("/v1/models", api.GetModels)
	r.POST("/v1/chat/completions", api.CreateChatCompletion)

	logger.Log.Info("Trae2API Start at :17080")
	// 启动服务器
	if err := r.Run(":17080"); err != nil {
		logger.Log.Fatalf("Trae2API Start Failed: %v", err)
	}
}
