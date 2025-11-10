package model

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"go-smart/internal/config"
	"go-smart/internal/logger"
)

// ModelManager 模型管理器
type ModelManager struct {
	currentModel     model.BaseChatModel
	currentModelName string
	currentAPIKey    string
	currentAPIBase   string
	mu               sync.RWMutex
	logger           *logger.Logger
	config           *config.Config
}

// NewModelManager 创建模型管理器
func NewModelManager(cfg *config.Config, log *logger.Logger) *ModelManager {
	mm := &ModelManager{
		currentModelName: cfg.AI.OpenAI.Model,
		currentAPIKey:    cfg.AI.OpenAI.APIKey,
		currentAPIBase:   cfg.AI.OpenAI.BaseURL,
		logger:           log,
		config:           cfg,
	}
	
	// 初始化模型
	if err := mm.initModel(); err != nil {
		log.Error("初始化模型失败", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	return mm
}

// initModel 初始化模型
func (mm *ModelManager) initModel() error {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	var err error
	var modelInstance model.BaseChatModel
	
	// 根据配置创建模型
	switch mm.config.AI.Provider {
	case "openai":
		modelInstance, err = mm.createOpenAIModel()
	case "mock":
		modelInstance, err = mm.createMockModel()
	default:
		// 默认使用OpenAI模型
		mm.config.AI.Provider = "openai"
		modelInstance, err = mm.createOpenAIModel()
	}
	
	if err != nil {
		return fmt.Errorf("创建模型失败: %w", err)
	}
	
	mm.currentModel = modelInstance
	mm.logger.Info("模型初始化成功", map[string]interface{}{
		"provider": mm.config.AI.Provider,
		"model":    mm.currentModelName,
	})
	
	return nil
}

// createOpenAIModel 创建OpenAI模型
func (mm *ModelManager) createOpenAIModel() (model.BaseChatModel, error) {
	cfg := &openai.ChatModelConfig{
		Model:   mm.currentModelName,
		APIKey:  mm.currentAPIKey,
		BaseURL: mm.currentAPIBase,
	}
	
	return openai.NewChatModel(context.Background(), cfg)
}

// createMockModel 创建Mock模型
func (mm *ModelManager) createMockModel() (model.BaseChatModel, error) {
	// 这里应该实现一个Mock模型，暂时返回错误
	return nil, fmt.Errorf("Mock模型暂未实现")
}

// GetCurrentModel 获取当前模型
func (mm *ModelManager) GetCurrentModel() model.BaseChatModel {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.currentModel
}

// GetCurrentModelInfo 获取当前模型信息
func (mm *ModelManager) GetCurrentModelInfo() map[string]string {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	return map[string]string{
		"provider":   mm.config.AI.Provider,
		"model_name": mm.currentModelName,
		"api_base":   mm.currentAPIBase,
		"status":     "active",
	}
}

// UpdateModel 更新模型配置
func (mm *ModelManager) UpdateModel(provider, modelName, apiKey, apiBase string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	// 保存旧配置以便回滚
	oldProvider := mm.config.AI.Provider
	oldModelName := mm.currentModelName
	oldAPIKey := mm.currentAPIKey
	oldAPIBase := mm.currentAPIBase
	
	// 更新配置
	if provider != "" {
		mm.config.AI.Provider = provider
	}
	if modelName != "" {
		mm.currentModelName = modelName
		mm.config.AI.OpenAI.Model = modelName
	}
	if apiKey != "" {
		mm.currentAPIKey = apiKey
		mm.config.AI.OpenAI.APIKey = apiKey
	}
	if apiBase != "" {
		mm.currentAPIBase = apiBase
		mm.config.AI.OpenAI.BaseURL = apiBase
	}
	
	// 初始化新模型
	err := mm.initModel()
	if err != nil {
		// 回滚到旧配置
		mm.config.AI.Provider = oldProvider
		mm.currentModelName = oldModelName
		mm.currentAPIKey = oldAPIKey
		mm.currentAPIBase = oldAPIBase
		
		mm.logger.Error("模型更新失败，已回滚到旧配置", map[string]interface{}{
			"error": err.Error(),
		})
		
		return fmt.Errorf("模型更新失败: %w", err)
	}
	
	mm.logger.Info("模型更新成功", map[string]interface{}{
		"provider":   mm.config.AI.Provider,
		"model_name": mm.currentModelName,
	})
	
	return nil
}

// ReloadFromEnv 从环境变量重新加载模型配置
func (mm *ModelManager) ReloadFromEnv() error {
	// 从环境变量读取新配置
	newProvider := os.Getenv("AI_PROVIDER")
	newModelName := os.Getenv("AI_MODEL")
	newAPIKey := os.Getenv("AI_API_KEY")
	newAPIBase := os.Getenv("AI_API_BASE")
	
	// 如果环境变量为空，使用当前配置
	if newProvider == "" {
		newProvider = mm.config.AI.Provider
	}
	if newModelName == "" {
		newModelName = mm.currentModelName
	}
	if newAPIKey == "" {
		newAPIKey = mm.currentAPIKey
	}
	if newAPIBase == "" {
		newAPIBase = mm.currentAPIBase
	}
	
	// 检查是否有变化
	if (newProvider == mm.config.AI.Provider && 
		newModelName == mm.currentModelName && 
		newAPIKey == mm.currentAPIKey && 
		newAPIBase == mm.currentAPIBase) {
		mm.logger.Info("模型配置无变化，无需更新", nil)
		return nil
	}
	
	// 更新模型
	return mm.UpdateModel(newProvider, newModelName, newAPIKey, newAPIBase)
}

// GetAvailableProviders 获取可用的模型提供商列表
func (mm *ModelManager) GetAvailableProviders() []string {
	return []string{"openai", "mock"}
}

// GetAvailableModels 获取指定提供商的可用模型列表
func (mm *ModelManager) GetAvailableModels(provider string) []string {
	switch provider {
	case "openai":
		return []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
	case "mock":
		return []string{"mock-model"}
	default:
		return []string{}
	}
}