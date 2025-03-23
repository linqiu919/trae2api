package main

import (
	"github.com/gin-gonic/gin"
	"github.com/trae2api/api"
	"github.com/trae2api/config"
	"github.com/trae2api/middleware"
	"github.com/trae2api/pkg/logger"
	"time"
)

func main() {

	// 设置全局时区为东八区（CST）
	time.Local = time.FixedZone("CST", 8*3600)

	// 设置 Gin 为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化日志
	logger.Init()

	// Initialize Redis
	err := config.InitRedisClient()
	if err != nil {
		logger.Log.Fatalln("failed to initialize Redis: " + err.Error())
	}

	// 初始化配置
	if err := config.InitConfig(); err != nil {
		logger.Log.Fatalf("Trae2API Config Init Failed: %v", err)
	}

	r := gin.Default()

	// 跨域
	r.Use(middleware.CORS())

	// 添加鉴权中间件
	r.Use(func(c *gin.Context) {
		api.AuthMiddleware()(c)
	})

	// OpenAI 格式的 API 路由
	r.GET("/v1/models", api.GetModels)
	r.POST("/v1/chat/completions", api.CreateChatCompletion)

	logger.Log.WithFields(map[string]interface{}{
		"port": 17080,
		"mode": gin.Mode(),
	}).Info("API 服务启动成功")

	// 启动服务器
	if err := r.Run(":17080"); err != nil {
		logger.Log.Fatalf("启动服务失败: %v", err)
	}
}
