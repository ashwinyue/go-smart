package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"go-smart/pkg/model"
	
	"github.com/cloudwego/eino/schema"
)

// LLMClient 大语言模型客户端接口
type LLMClient interface {
	// Chat 对话
	Chat(ctx context.Context, messages []map[string]interface{}, tools []map[string]interface{}) (*ChatResponse, error)
	// GetModelInfo 获取模型信息
	GetModelInfo() map[string]string
}

// ChatResponse 对话响应
type ChatResponse struct {
	Content   string      `json:"content"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string                 `json:"id"`
	Function ToolCallFunction       `json:"function"`
}

// ToolCallFunction 工具调用函数
type ToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// EinoLLMClient 基于Eino的LLM客户端实现
type EinoLLMClient struct {
	modelManager *model.ModelManager
}

// NewEinoLLMClient 创建Eino LLM客户端
func NewEinoLLMClient(modelManager *model.ModelManager) *EinoLLMClient {
	return &EinoLLMClient{
		modelManager: modelManager,
	}
}

// Chat 实现对话
func (c *EinoLLMClient) Chat(ctx context.Context, messages []map[string]interface{}, tools []map[string]interface{}) (*ChatResponse, error) {
	// 获取当前模型
	model := c.modelManager.GetCurrentModel()
	if model == nil {
		return nil, fmt.Errorf("模型未初始化")
	}
	
	// 转换消息格式
	einoMessages, err := c.convertMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("消息格式转换失败: %v", err)
	}
	
	// 转换为schema.Message类型
	schemaMessages := make([]*schema.Message, 0, len(einoMessages))
	for _, msg := range einoMessages {
		if msgMap, ok := msg.(map[string]interface{}); ok {
			role, _ := msgMap["role"].(string)
			content, _ := msgMap["content"].(string)
			
			var message *schema.Message
			switch role {
			case "system":
				message = schema.SystemMessage(content)
			case "assistant":
				// AssistantMessage需要额外的ToolCall参数
				message = schema.AssistantMessage(content, nil)
			case "user":
				message = schema.UserMessage(content)
			default:
				message = schema.UserMessage(content)
			}
			
			schemaMessages = append(schemaMessages, message)
		}
	}
	
	// 调用模型
	result, err := model.Generate(ctx, schemaMessages)
	if err != nil {
		return nil, fmt.Errorf("模型调用失败: %v", err)
	}
	
	// 转换响应格式
	response := &ChatResponse{
		Content: result.Content,
	}
	
	// 如果有工具调用，转换工具调用格式
	if result.ToolCalls != nil {
		for _, toolCall := range result.ToolCalls {
			// 解析工具调用参数
			var args map[string]interface{}
			if toolCall.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					// 如果解析失败，使用原始字符串
					args = map[string]interface{}{"raw": toolCall.Function.Arguments}
				}
			}
			
			response.ToolCalls = append(response.ToolCalls, ToolCall{
				ID: toolCall.ID,
				Function: ToolCallFunction{
					Name:      toolCall.Function.Name,
					Arguments: args,
				},
			})
		}
	}
	
	return response, nil
}

// GetModelInfo 获取模型信息
func (c *EinoLLMClient) GetModelInfo() map[string]string {
	return c.modelManager.GetCurrentModelInfo()
}

// convertMessages 转换消息格式
func (c *EinoLLMClient) convertMessages(messages []map[string]interface{}) ([]interface{}, error) {
	// 这里需要根据Eino的实际消息格式进行转换
	// 以下是伪代码，需要根据Eino的实际API实现
	var einoMessages []interface{}
	
	for _, msg := range messages {
		role, ok := msg["role"].(string)
		if !ok {
			return nil, fmt.Errorf("消息角色缺失")
		}
		
		content, ok := msg["content"].(string)
		if !ok {
			return nil, fmt.Errorf("消息内容缺失")
		}
		
		// 根据Eino的实际API创建消息对象
		einoMsg := map[string]interface{}{
			"role":    role,
			"content": content,
		}
		
		einoMessages = append(einoMessages, einoMsg)
	}
	
	return einoMessages, nil
}

// convertTools 转换工具格式
func (c *EinoLLMClient) convertTools(tools []map[string]interface{}) ([]interface{}, error) {
	// 这里需要根据Eino的实际工具格式进行转换
	// 以下是伪代码，需要根据Eino的实际API实现
	var einoTools []interface{}
	
	for _, tool := range tools {
		toolType, ok := tool["type"].(string)
		if !ok {
			return nil, fmt.Errorf("工具类型缺失")
		}
		
		if toolType != "function" {
			return nil, fmt.Errorf("不支持的工具类型: %s", toolType)
		}
		
		function, ok := tool["function"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("工具函数定义缺失")
		}
		
		name, ok := function["name"].(string)
		if !ok {
			return nil, fmt.Errorf("工具名称缺失")
		}
		
		description, ok := function["description"].(string)
		if !ok {
			return nil, fmt.Errorf("工具描述缺失")
		}
		
		parameters, ok := function["parameters"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("工具参数定义缺失")
		}
		
		// 根据Eino的实际API创建工具对象
		einoTool := map[string]interface{}{
			"type":        toolType,
			"function": map[string]interface{}{
				"name":        name,
				"description": description,
				"parameters":  parameters,
			},
		}
		
		einoTools = append(einoTools, einoTool)
	}
	
	return einoTools, nil
}