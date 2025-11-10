package tools

import (
	"go-smart/pkg/tools/business"
	"sync"
)

// ToolManager 工具管理器
type ToolManager struct {
	registry *ToolRegistry
	mu       sync.RWMutex
}

// NewToolManager 创建工具管理器
func NewToolManager() *ToolManager {
	tm := &ToolManager{
		registry: NewToolRegistry(),
	}
	
	// 注册默认工具
	tm.registerDefaultTools()
	
	return tm
}

// registerDefaultTools 注册默认工具
func (tm *ToolManager) registerDefaultTools() {
	// 注册订单查询工具
	orderQueryTool := business.NewOrderQueryTool()
	tm.registry.RegisterTool(orderQueryTool)
	
	// 注册退款申请工具
	refundRequestTool := business.NewRefundRequestTool(orderQueryTool)
	tm.registry.RegisterTool(refundRequestTool)
	
	// 注册发票工具
	invoiceTool := business.NewInvoiceTool()
	tm.registry.RegisterTool(invoiceTool)
}

// GetRegistry 获取工具注册表
func (m *ToolManager) GetRegistry() *ToolRegistry {
	return m.registry
}

// GetTool 获取指定工具
func (m *ToolManager) GetTool(name string) (ToolFunction, bool) {
	return m.registry.GetTool(name)
}

// GetAllTools 获取所有工具
func (m *ToolManager) GetAllTools() map[string]ToolFunction {
	return m.registry.GetAllTools()
}

// GetToolsSchema 获取所有工具的Schema
func (m *ToolManager) GetToolsSchema() []map[string]interface{} {
	return m.registry.GetToolsSchema()
}

// CallTool 调用工具
func (m *ToolManager) CallTool(name string, args map[string]interface{}) (map[string]interface{}, error) {
	return m.registry.CallTool(name, args)
}

// ReloadTools 重新加载所有工具
func (m *ToolManager) ReloadTools() error {
	// 清空当前注册表
	m.registry.Clear()
	
	// 重新注册工具
	m.registerDefaultTools()
	
	return nil
}

// RegisterTool 注册新工具
func (m *ToolManager) RegisterTool(tool ToolFunction) error {
	return m.registry.RegisterTool(tool)
}

// UnregisterTool 注销工具
func (m *ToolManager) UnregisterTool(name string) error {
	return m.registry.UnregisterTool(name)
}