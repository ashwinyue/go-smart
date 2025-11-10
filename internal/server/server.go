package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-smart/internal/config"
	"go-smart/internal/handler"
	"go-smart/internal/logger"
)

// Server HTTP服务器
type Server struct {
	config *config.ServerConfig
	logger *logger.Logger
	router *gin.Engine
	server *http.Server
}

// NewServer 创建新的HTTP服务器
func NewServer(cfg *config.ServerConfig, log *logger.Logger) *Server {
	// 设置Gin模式
	gin.SetMode(cfg.Mode)

	// 创建Gin路由器
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 创建HTTP服务器
	readTimeout, err := time.ParseDuration(fmt.Sprintf("%vs", cfg.ReadTimeout))
	if err != nil {
		log.Error("解析读取超时时间失败", map[string]interface{}{
			"error": err.Error(),
			"value": cfg.ReadTimeout,
		})
		readTimeout = 30 * time.Second // 默认30秒
	}
	
	writeTimeout, err := time.ParseDuration(fmt.Sprintf("%vs", cfg.WriteTimeout))
	if err != nil {
		log.Error("解析写入超时时间失败", map[string]interface{}{
			"error": err.Error(),
			"value": cfg.WriteTimeout,
		})
		writeTimeout = 30 * time.Second // 默认30秒
	}
	
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return &Server{
		config: cfg,
		logger: log,
		router: router,
		server: server,
	}
}

// GetRouter 获取Gin路由器
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	s.logger.Info("启动HTTP服务器", map[string]interface{}{
		"port": s.config.Port,
		"mode": s.config.Mode,
	})

	return s.server.ListenAndServe()
}

// Shutdown 优雅关闭HTTP服务器
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("正在关闭HTTP服务器", nil)
	return s.server.Shutdown(ctx)
}

// SetupRoutes 设置路由
func (s *Server) SetupRoutes(chatHandler *handler.ChatHandler) {
	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API路由组
	api := s.router.Group("/api/v1")
	{
		// 聊天接口
		api.POST("/chat", chatHandler.Chat)
		
		// 订单查询接口
		api.POST("/order/query", chatHandler.OrderQuery)
		
		// 发票相关接口
		api.POST("/invoice/create", chatHandler.CreateInvoice)
		api.POST("/invoice/query", chatHandler.QueryInvoice)
		
		// 模型管理接口
		api.GET("/model/current", chatHandler.GetCurrentModel)
		api.PUT("/model/update", chatHandler.UpdateModel)
		
		// 插件管理接口
		api.GET("/plugins", chatHandler.GetPlugins)
		api.POST("/plugins/:name/reload", chatHandler.ReloadPlugin)
		api.POST("/plugins/:name/unload", chatHandler.UnloadPlugin)
		api.POST("/plugin/execute", chatHandler.ExecutePluginFunction)
		
		// 对话历史接口
		api.POST("/conversation/history", chatHandler.History)
		
		// 清除对话历史接口
		api.POST("/conversation/clear", chatHandler.Clear)
		
		// 测试接口
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}
}