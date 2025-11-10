package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ModelConfig 模型配置
type ModelConfig struct {
	APIKey      string
	ModelName   string
	Temperature float64
	APIBase     string // 新增API基础URL，用于自定义OpenAI兼容API
	ModelType   string // 新增模型类型，用于区分不同模型提供商
}

// OpenAIModel OpenAI 模型适配器
type OpenAIModel struct {
	apiKey      string
	modelName   string
	temperature float64
	apiBase     string
	client      *http.Client
}

// NewOpenAIModel 创建 OpenAI 模型实例
func NewOpenAIModel(config ModelConfig) (model.BaseChatModel, error) {
	if config.APIKey == "" {
		config.APIKey = os.Getenv("OPENAI_API_KEY")
		if config.APIKey == "" {
			return nil, fmt.Errorf("未设置 OpenAI API Key，请设置环境变量 OPENAI_API_KEY 或在配置中提供")
		}
	}
	
	if config.ModelName == "" {
		config.ModelName = "gpt-3.5-turbo" // 默认使用 gpt-3.5-turbo
	}
	
	if config.APIBase == "" {
		config.APIBase = os.Getenv("OPENAI_API_BASE")
		if config.APIBase == "" {
			config.APIBase = "https://api.openai.com/v1" // 默认使用官方API
		}
	}
	
	return &OpenAIModel{
		apiKey:      config.APIKey,
		modelName:   config.ModelName,
		temperature: config.Temperature,
		apiBase:     config.APIBase,
		client:      &http.Client{},
	}, nil
}

// OpenAIRequest OpenAI API 请求结构
type OpenAIRequest struct {
	Model       string         `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64        `json:"temperature,omitempty"`
}

// OpenAIMessage OpenAI 消息结构
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse OpenAI API 响应结构
type OpenAIResponse struct {
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// BindTools 绑定工具（暂不支持）
func (m *OpenAIModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}

// Generate 生成回复
func (m *OpenAIModel) Generate(ctx context.Context, messages []*schema.Message, options ...model.Option) (*schema.Message, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("没有提供消息")
	}
	
	// 转换消息格式
	openaiMessages := make([]OpenAIMessage, 0, len(messages))
	for _, msg := range messages {
		role := "user"
		switch msg.Role {
		case schema.Assistant:
			role = "assistant"
		case schema.System:
			role = "system"
		}
		
		openaiMessages = append(openaiMessages, OpenAIMessage{
			Role:    role,
			Content: msg.Content,
		})
	}
	
	// 创建请求
	request := OpenAIRequest{
		Model:       m.modelName,
		Messages:    openaiMessages,
		Temperature: m.temperature,
	}
	
	// 序列化请求
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	
	// 创建 HTTP 请求
	apiURL := m.apiBase
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	apiURL += "chat/completions"
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	
	// 发送请求
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	
	// 解析响应
	var openaiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	// 检查错误
	if openaiResp.Error != nil {
		return nil, fmt.Errorf("API 错误: %s", openaiResp.Error.Message)
	}
	
	// 检查响应
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("没有收到响应")
	}
	
	// 返回结果
	return schema.AssistantMessage(openaiResp.Choices[0].Message.Content, nil), nil
}

// Stream 流式生成回复（暂不支持）
func (m *OpenAIModel) Stream(ctx context.Context, messages []*schema.Message, options ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, fmt.Errorf("暂不支持流式回复")
}

// GetType 获取模型类型
func (m *OpenAIModel) GetType() string {
	return "openai"
}

// GetTokenCount 获取 token 数量（暂不支持）
func (m *OpenAIModel) GetTokenCount(ctx context.Context, messages []*schema.Message) (int, error) {
	return 0, nil
}

// MockModel 用于测试的模拟模型
type MockModel struct {
	responses map[string]string
}

// NewMockModel 创建模拟模型实例
func NewMockModel() *MockModel {
	return &MockModel{
		responses: map[string]string{
			"我昨天下的单": "您昨天下的订单已经发货，预计明天送达。订单号：ORD20240114001。",
			"查订单":      "请提供您的订单号，我将为您查询订单状态。",
			"退款":       "请提供您需要退款的订单号，我将为您处理退款申请。",
		},
	}
}

// BindTools 绑定工具（模拟模型不需要）
func (m *MockModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}

// Generate 生成回复
func (m *MockModel) Generate(ctx context.Context, messages []*schema.Message, options ...model.Option) (*schema.Message, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("没有提供消息")
	}
	
	lastMessage := messages[len(messages)-1]
	content := lastMessage.Content
	
	// 查找预设回复
	if response, ok := m.responses[content]; ok {
		return schema.AssistantMessage(response, nil), nil
	}
	
	// 默认回复
	return schema.AssistantMessage("感谢您的咨询，我会尽力为您提供帮助。", nil), nil
}

// Stream 流式生成回复（模拟模型不支持）
func (m *MockModel) Stream(ctx context.Context, messages []*schema.Message, options ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, fmt.Errorf("模拟模型不支持流式回复")
}

// GetType 获取模型类型
func (m *MockModel) GetType() string {
	return "mock"
}

// BindTools 绑定工具（模拟模型不需要）
func (m *MockModel) GetTokenCount(ctx context.Context, messages []*schema.Message) (int, error) {
	return 0, nil
}