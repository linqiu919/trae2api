package main

import (
	"github.com/gin-gonic/gin"
	"github.com/trae2api/api"
	"github.com/trae2api/pkg/logger"
	"net/http"
	"time"
)

func main() {

	// 设置全局时区为东八区（CST）
	time.Local = time.FixedZone("CST", 8*3600)

	// 设置 Gin 为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化日志
	logger.Init()

	// 初始化配置
	//if err := config.InitConfig(); err != nil {
	//	logger.Log.Fatalf("Trae2API Config Init Failed: %v", err)
	//}

	r := gin.Default()

	// 加载HTML模板
	r.LoadHTMLGlob("templates/*")

	// 添加鉴权中间件，但排除 /index、/login 和 /verify 路径
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/index" || path == "/login" || path == "/verify" {
			c.Next()
			return
		}
		api.AuthMiddleware()(c)
	})

	// 添加根路径重定向到登录页面
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	// 添加登录相关路由
	r.GET("/login", api.LoginPageHandler)
	r.POST("/verify", api.VerifyPasswordHandler)

	// 添加索引页面路由
	r.GET("/index", api.IndexHandler)

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
