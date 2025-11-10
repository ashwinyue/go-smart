package tools

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
)

// ToolFunction 定义工具函数的通用接口
type ToolFunction interface {
	Call(args map[string]interface{}) (map[string]interface{}, error)
	GetDescription() string
	GetName() string
	GetParameters() map[string]interface{}
}

// ToolRegistry 工具注册表，用于管理所有可用的工具
type ToolRegistry struct {
	tools map[string]ToolFunction
}

// NewToolRegistry 创建新的工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolFunction),
	}
}

// RegisterTool 注册工具到注册表
func (r *ToolRegistry) RegisterTool(tool ToolFunction) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}
	
	name := tool.GetName()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool with name '%s' already registered", name)
	}
	
	r.tools[name] = tool
	return nil
}

// GetTool 根据名称获取工具
func (r *ToolRegistry) GetTool(name string) (ToolFunction, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// GetAllTools 获取所有已注册的工具
func (r *ToolRegistry) GetAllTools() map[string]ToolFunction {
	// 创建副本以避免外部修改
	tools := make(map[string]ToolFunction)
	for name, tool := range r.tools {
		tools[name] = tool
	}
	return tools
}

// GetToolsSchema 获取所有工具的JSON Schema格式描述，用于大模型调用
func (r *ToolRegistry) GetToolsSchema() []map[string]interface{} {
	schemas := make([]map[string]interface{}, 0, len(r.tools))
	
	for _, tool := range r.tools {
		schema := map[string]interface{}{
			"name":        tool.GetName(),
			"description": tool.GetDescription(),
			"parameters":  tool.GetParameters(),
		}
		schemas = append(schemas, schema)
	}
	
	return schemas
}

// CallTool 调用指定工具
func (r *ToolRegistry) CallTool(name string, args map[string]interface{}) (map[string]interface{}, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}
	
	return tool.Call(args)
}

// UnregisterTool 从注册表中移除工具
func (r *ToolRegistry) UnregisterTool(name string) error {
	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool '%s' not found", name)
	}
	
	delete(r.tools, name)
	return nil
}

// Clear 清空所有已注册的工具
func (r *ToolRegistry) Clear() {
	r.tools = make(map[string]ToolFunction)
}

// Count 返回已注册工具的数量
func (r *ToolRegistry) Count() int {
	return len(r.tools)
}

// ToolCallRequest 表示大模型返回的工具调用请求
type ToolCallRequest struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResponse 表示工具调用的响应
type ToolCallResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// BaseTool 基础工具实现，提供通用功能
type BaseTool struct {
	name        string
	description string
	parameters  map[string]interface{}
}

// NewBaseTool 创建基础工具实例
func NewBaseTool(name, description string, parameters map[string]interface{}) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		parameters:  parameters,
	}
}

// GetName 获取工具名称
func (t *BaseTool) GetName() string {
	return t.name
}

// GetDescription 获取工具描述
func (t *BaseTool) GetDescription() string {
	return t.description
}

// GetParameters 获取工具参数
func (t *BaseTool) GetParameters() map[string]interface{} {
	return t.parameters
}

// ValidateArgs 验证参数是否符合工具的参数模式
func (t *BaseTool) ValidateArgs(args map[string]interface{}) error {
	// 简单验证：检查必需参数是否存在
	if requiredParams, ok := t.parameters["required"].([]interface{}); ok {
		for _, param := range requiredParams {
			paramName, ok := param.(string)
			if !ok {
				continue
			}
			
			if _, exists := args[paramName]; !exists {
				return fmt.Errorf("missing required parameter: %s", paramName)
			}
		}
	}
	
	return nil
}

// ConvertArgs 将参数转换为指定的类型
func (t *BaseTool) ConvertArgs(args map[string]interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal args: %w", err)
	}
	
	err = json.Unmarshal(jsonBytes, target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal args to target type: %w", err)
	}
	
	return nil
}

// GetFunctionName 从函数获取名称
func GetFunctionName(fn interface{}) string {
	v := reflect.ValueOf(fn)
	if v.Kind() == reflect.Func {
		return runtime.FuncForPC(v.Pointer()).Name()
	}
	return ""
}