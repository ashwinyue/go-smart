package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"go-smart/pkg/llm"
	"go-smart/pkg/tools"
	"strings"
)

// State 状态图状态
type State struct {
	Messages     []Message `json:"messages"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	ToolResults  []ToolResult `json:"tool_results,omitempty"`
	NextAction   string `json:"next_action"`
	IsComplete   bool `json:"is_complete"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"` // user, assistant, system
	Content string `json:"content"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Args     map[string]interface{} `json:"args"`
}

// ToolResult 工具结果
type ToolResult struct {
	ToolCallID string                 `json:"tool_call_id"`
	Result     map[string]interface{} `json:"result"`
	Error      string                 `json:"error,omitempty"`
}

// Workflow 工作流
type Workflow struct {
	llmClient     llm.LLMClient
	toolManager   *tools.ToolManager
	state         State
}

// NewWorkflow 创建工作流
func NewWorkflow(llmClient llm.LLMClient, toolManager *tools.ToolManager) *Workflow {
	return &Workflow{
		llmClient:   llmClient,
		toolManager: toolManager,
		state: State{
			Messages:   []Message{},
			IsComplete: false,
		},
	}
}

// ProcessMessage 处理消息
func (w *Workflow) ProcessMessage(ctx context.Context, userMessage string) (string, error) {
	// 添加用户消息到状态
	w.state.Messages = append(w.state.Messages, Message{
		Role:    "user",
		Content: userMessage,
	})
	
	// 循环处理直到完成
	for !w.state.IsComplete {
		// 调用大模型
		response, err := w.callModel(ctx)
		if err != nil {
			return "", fmt.Errorf("调用模型失败: %v", err)
		}
		
		// 添加助手回复到状态
		w.state.Messages = append(w.state.Messages, Message{
			Role:    "assistant",
			Content: response.Content,
		})
		
		// 检查是否有工具调用
		if len(response.ToolCalls) > 0 {
			// 保存工具调用
			w.state.ToolCalls = append(w.state.ToolCalls, response.ToolCalls...)
			
			// 执行工具调用
			for _, toolCall := range response.ToolCalls {
				result, err := w.executeTool(ctx, toolCall)
				if err != nil {
					toolResult := ToolResult{
						ToolCallID: toolCall.ID,
						Error:      err.Error(),
					}
					w.state.ToolResults = append(w.state.ToolResults, toolResult)
				} else {
					toolResult := ToolResult{
						ToolCallID: toolCall.ID,
						Result:     result,
					}
					w.state.ToolResults = append(w.state.ToolResults, toolResult)
				}
			}
			
			// 继续循环，让模型处理工具结果
			continue
		} else {
			// 没有工具调用，完成处理
			w.state.IsComplete = true
		}
	}
	
	// 获取最后的助手回复
	if len(w.state.Messages) > 0 {
		lastMessage := w.state.Messages[len(w.state.Messages)-1]
		if lastMessage.Role == "assistant" {
			return lastMessage.Content, nil
		}
	}
	
	return "", fmt.Errorf("未找到有效的回复")
}

// ModelResponse 模型响应
type ModelResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// callModel 调用模型
func (w *Workflow) callModel(ctx context.Context) (*ModelResponse, error) {
	// 构建消息
	messages := make([]map[string]interface{}, 0, len(w.state.Messages)+len(w.state.ToolResults))
	
	// 添加系统提示
	systemPrompt := w.buildSystemPrompt()
	messages = append(messages, map[string]interface{}{
		"role":    "system",
		"content": systemPrompt,
	})
	
	// 添加历史消息
	for _, msg := range w.state.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	
	// 添加工具结果
	for _, result := range w.state.ToolResults {
		content := ""
		if result.Error != "" {
			content = fmt.Sprintf("工具执行出错: %s", result.Error)
		} else {
			resultJSON, _ := json.Marshal(result.Result)
			content = string(resultJSON)
		}
		
		messages = append(messages, map[string]interface{}{
			"role":    "tool",
			"content": content,
		})
	}
	
	// 获取工具定义
	availableTools := w.toolManager.GetAllTools()
	toolDefinitions := make([]map[string]interface{}, 0, len(availableTools))
	
	for _, tool := range availableTools {
		toolDefinitions = append(toolDefinitions, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.GetName(),
				"description": tool.GetDescription(),
				"parameters":  tool.GetParameters(),
			},
		})
	}
	
	// 调用模型
	response, err := w.llmClient.Chat(ctx, messages, toolDefinitions)
	if err != nil {
		return nil, err
	}
	
	// 解析响应
	modelResponse := &ModelResponse{
		Content: response.Content,
	}
	
	// 解析工具调用
	if response.ToolCalls != nil {
		for _, toolCall := range response.ToolCalls {
			modelResponse.ToolCalls = append(modelResponse.ToolCalls, ToolCall{
				ID:   toolCall.ID,
				Name: toolCall.Function.Name,
				Args: toolCall.Function.Arguments,
			})
		}
	}
	
	return modelResponse, nil
}

// buildSystemPrompt 构建系统提示
func (w *Workflow) buildSystemPrompt() string {
	tools := w.toolManager.GetAllTools()
	var toolDescriptions []string
	
	for _, tool := range tools {
		toolDescriptions = append(toolDescriptions, fmt.Sprintf("- %s: %s", tool.GetName(), tool.GetDescription()))
	}
	
	return fmt.Sprintf(`你是一个智能助手，可以帮助用户处理订单查询、退款申请和发票相关的问题。

可用工具:
%s

使用工具的规则:
1. 当用户需要查询订单信息时，使用order_query_tool
2. 当用户需要申请退款时，使用refund_request_tool
3. 当用户需要创建或查询发票时，使用invoice_tool

请根据用户的问题，选择合适的工具来帮助用户。如果不需要使用工具，可以直接回答用户的问题。`, strings.Join(toolDescriptions, "\n"))
}

// executeTool 执行工具
func (w *Workflow) executeTool(ctx context.Context, toolCall ToolCall) (map[string]interface{}, error) {
	tool, exists := w.toolManager.GetTool(toolCall.Name)
	if !exists {
		return nil, fmt.Errorf("工具不存在: %s", toolCall.Name)
	}
	
	return tool.Call(toolCall.Args)
}

// Reset 重置工作流状态
func (w *Workflow) Reset() {
	w.state = State{
		Messages:   []Message{},
		IsComplete: false,
	}
}