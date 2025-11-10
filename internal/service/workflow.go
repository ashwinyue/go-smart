package service

import (
	"context"
	"fmt"
	"go-smart/internal/config"
	"go-smart/internal/logger"
	"go-smart/pkg/graph"
	"go-smart/pkg/llm"
	"go-smart/pkg/model"
	"go-smart/pkg/tools"
)

// WorkflowService 工作流服务
type WorkflowService struct {
	workflow    *graph.Workflow
	llmClient   llm.LLMClient
	toolManager *tools.ToolManager
	logger      *logger.Logger
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(cfg *config.Config, log *logger.Logger) (*WorkflowService, error) {
	// 创建模型管理器
	modelManager := model.NewModelManager(cfg, log)
	
	// 创建LLM客户端
	llmClient := llm.NewEinoLLMClient(modelManager)
	
	// 创建工具管理器
	toolManager := tools.NewToolManager()
	
	// 创建工作流
	workflow := graph.NewWorkflow(llmClient, toolManager)
	
	return &WorkflowService{
		workflow:    workflow,
		llmClient:   llmClient,
		toolManager: toolManager,
		logger:      log,
	}, nil
}

// ProcessMessage 处理消息
func (s *WorkflowService) ProcessMessage(ctx context.Context, message string) (map[string]interface{}, error) {
	s.logger.Info("工作流处理消息", map[string]interface{}{
		"message": message,
	})
	
	// 重置工作流状态
	s.workflow.Reset()
	
	// 处理消息
	response, err := s.workflow.ProcessMessage(ctx, message)
	if err != nil {
		s.logger.Error("工作流处理消息失败", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("工作流处理消息失败: %w", err)
	}
	
	s.logger.Info("工作流处理消息成功", map[string]interface{}{
		"response": response,
	})
	
	return map[string]interface{}{
		"response": response,
	}, nil
}

// ProcessMultiTurnMessage 处理多轮对话消息
func (s *WorkflowService) ProcessMultiTurnMessage(ctx context.Context, sessionID, message string) (string, error) {
	s.logger.Info("工作流处理多轮对话消息", map[string]interface{}{
		"session_id": sessionID,
		"message":    message,
	})
	
	// TODO: 实现会话状态管理
	// 目前简单处理，每次都重置工作流
	s.workflow.Reset()
	
	// 处理消息
	response, err := s.workflow.ProcessMessage(ctx, message)
	if err != nil {
		s.logger.Error("工作流处理多轮对话消息失败", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("工作流处理多轮对话消息失败: %w", err)
	}
	
	s.logger.Info("工作流处理多轮对话消息成功", map[string]interface{}{
		"session_id": sessionID,
		"response":   response,
	})
	
	return response, nil
}

// GetModelInfo 获取模型信息
func (s *WorkflowService) GetModelInfo() map[string]string {
	return s.llmClient.GetModelInfo()
}

// GetTools 获取所有工具信息
func (s *WorkflowService) GetTools() map[string]interface{} {
	tools := s.toolManager.GetAllTools()
	toolInfos := make(map[string]interface{})
	
	for name, tool := range tools {
		toolInfos[name] = map[string]interface{}{
			"name":        tool.GetName(),
			"description": tool.GetDescription(),
			"parameters":  tool.GetParameters(),
		}
	}
	
	return toolInfos
}