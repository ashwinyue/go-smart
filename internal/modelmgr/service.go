package modelmgr

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go-smart/internal/config"
	modelpkg "go-smart/pkg/model"
)

// Service 模型服务
type Service struct {
	chatModel model.BaseChatModel
	provider  string
}

// NewService 创建新的模型服务
func NewService(cfg *config.AIConfig) (*Service, error) {
	var chatModel model.BaseChatModel
	var err error

	switch cfg.Provider {
	case "openai":
		modelConfig := modelpkg.ModelConfig{
			APIKey:      cfg.OpenAI.APIKey,
			ModelName:   cfg.OpenAI.Model,
			Temperature: cfg.OpenAI.Temperature,
			APIBase:     cfg.OpenAI.BaseURL,
		}
		chatModel, err = modelpkg.NewOpenAIModel(modelConfig)
	case "mock":
		chatModel = modelpkg.NewMockModel()
	default:
		// 默认使用OpenAI模型
		cfg.Provider = "openai"
		modelConfig := modelpkg.ModelConfig{
			APIKey:      cfg.OpenAI.APIKey,
			ModelName:   cfg.OpenAI.Model,
			Temperature: cfg.OpenAI.Temperature,
			APIBase:     cfg.OpenAI.BaseURL,
		}
		chatModel, err = modelpkg.NewOpenAIModel(modelConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("创建模型失败: %w", err)
	}

	return &Service{
		chatModel: chatModel,
		provider:  cfg.Provider,
	}, nil
}

// GetChatModel 获取聊天模型
func (s *Service) GetChatModel() model.BaseChatModel {
	return s.chatModel
}

// GetProvider 获取提供商
func (s *Service) GetProvider() string {
	return s.provider
}

// Generate 生成回复
func (s *Service) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return s.chatModel.Generate(ctx, messages)
}