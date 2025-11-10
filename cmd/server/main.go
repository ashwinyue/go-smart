package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go-smart/internal/config"
	"go-smart/internal/handler"
	"go-smart/internal/logger"
	"go-smart/internal/modelmgr"
	"go-smart/internal/server"
	"go-smart/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		panic("加载配置失败: " + err.Error())
	}

	// 创建日志器
	log, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		panic("创建日志器失败: " + err.Error())
	}
	defer func() {
		// 刷新日志缓冲区
		if syncErr := log.Sync(); syncErr != nil {
			log.Error("同步日志失败", map[string]interface{}{
				"error": syncErr.Error(),
			})
		}
	}()

	log.Info("应用程序启动", nil)

	// 创建模型服务
	modelService, err := modelmgr.NewService(&cfg.AI)
	if err != nil {
		log.Error("创建模型服务失败", map[string]interface{}{
			"error": err.Error(),
		})
		panic("创建模型服务失败: " + err.Error())
	}

	log.Info("模型服务创建成功", map[string]interface{}{
		"provider": modelService.GetProvider(),
	})

	// 创建对话服务
	conversationService, err := service.NewConversationService(
		context.Background(),
		modelService.GetChatModel(),
		log,
		cfg,
	)
	if err != nil {
		log.Error("创建对话服务失败", map[string]interface{}{
			"error": err.Error(),
		})
		panic("创建对话服务失败: " + err.Error())
	}

	// 创建工作流服务
	workflowService, err := service.NewWorkflowService(cfg, log)
	if err != nil {
		log.Error("创建工作流服务失败", map[string]interface{}{
			"error": err.Error(),
		})
		panic("创建工作流服务失败: " + err.Error())
	}

	// 创建聊天处理器
	chatHandler := handler.NewChatHandler(conversationService, workflowService, log)

	// 创建HTTP服务器
	httpServer := server.NewServer(&cfg.Server, log)
	httpServer.SetupRoutes(chatHandler)

	// 启动HTTP服务器
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Error("HTTP服务器启动失败", map[string]interface{}{
				"error": err.Error(),
			})
			panic("HTTP服务器启动失败: " + err.Error())
		}
	}()

	log.Info("应用程序启动完成", nil)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭应用程序", nil)

	// 创建关闭上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 优雅关闭HTTP服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP服务器关闭失败", map[string]interface{}{
			"error": err.Error(),
		})
	}

	log.Info("应用程序已关闭", nil)
}