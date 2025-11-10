package handler

import (
	"net/http"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"go-smart/internal/logger"
	"go-smart/internal/service"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	conversationService *service.ConversationService
	workflowService    *service.WorkflowService
	logger             *logger.Logger
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(conversationService *service.ConversationService, workflowService *service.WorkflowService, log *logger.Logger) *ChatHandler {
	return &ChatHandler{
		conversationService: conversationService,
		workflowService:    workflowService,
		logger:             log,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message   string `json:"message" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
	UseWorkflow bool `json:"use_workflow,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Response string `json:"response"`
	Date     string `json:"date,omitempty"`
}

// Chat 处理聊天请求
func (h *ChatHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的聊天请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("收到聊天请求", map[string]interface{}{
		"message":     req.Message,
		"session_id":  req.SessionID,
		"use_workflow": req.UseWorkflow,
	})

	var response string
	var err error

	// 根据请求决定使用哪种处理方式
	if req.UseWorkflow {
		// 使用新的工作流处理
		if req.SessionID != "" {
			response, err = h.workflowService.ProcessMultiTurnMessage(c.Request.Context(), req.SessionID, req.Message)
		} else {
			result, procErr := h.workflowService.ProcessMessage(c.Request.Context(), req.Message)
			if procErr == nil {
				response = result["response"].(string)
			}
			if procErr != nil {
				err = procErr
			}
		}
	} else {
		// 使用原有的对话服务处理
		if req.SessionID != "" {
			response, err = h.conversationService.ProcessMultiTurnMessage(c.Request.Context(), req.SessionID, req.Message)
		} else {
			result, procErr := h.conversationService.ProcessMessage(c.Request.Context(), req.Message)
			if procErr == nil {
				response = result["response"].(string)
			}
			if procErr != nil {
				err = procErr
			}
		}
	}

	if err != nil {
		h.logger.Error("处理聊天消息失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "处理消息失败",
		})
		return
	}

	// 返回响应
	chatResponse := ChatResponse{
		Response: response,
	}

	c.JSON(http.StatusOK, chatResponse)
}

// OrderQueryRequest 订单查询请求
type OrderQueryRequest struct {
	Query string `json:"query" binding:"required"`
}

// OrderQueryResponse 订单查询响应
type OrderQueryResponse struct {
	Result string `json:"result"`
}

// OrderQuery 处理订单查询请求
func (h *ChatHandler) OrderQuery(c *gin.Context) {
	var req OrderQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的订单查询请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("收到订单查询请求", map[string]interface{}{
		"query": req.Query,
	})

	// 处理订单查询
	result, err := h.conversationService.ProcessOrderQuery(c.Request.Context(), req.Query)
	if err != nil {
		h.logger.Error("处理订单查询失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "处理订单查询失败",
		})
		return
	}

	// 返回响应
	response := OrderQueryResponse{
		Result: result,
	}

	c.JSON(http.StatusOK, response)
}

// HistoryRequest 获取对话历史请求
type HistoryRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// HistoryResponse 获取对话历史响应
type HistoryResponse struct {
	SessionID string      `json:"session_id"`
	History   []Message   `json:"history"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// History 获取对话历史
func (h *ChatHandler) History(c *gin.Context) {
	var req HistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的对话历史请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("获取对话历史", map[string]interface{}{
		"session_id": req.SessionID,
	})

	// 获取对话历史
	messages, err := h.conversationService.GetConversationHistory(req.SessionID)
	if err != nil {
		h.logger.Error("获取对话历史失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取对话历史失败",
		})
		return
	}

	// 转换消息格式
	history := make([]Message, 0, len(messages))
	for _, msg := range messages {
		role := "user"
		if msg.Role == schema.Assistant {
			role = "assistant"
		}
		history = append(history, Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 返回响应
	response := HistoryResponse{
		SessionID: req.SessionID,
		History:   history,
	}

	c.JSON(http.StatusOK, response)
}

// ClearRequest 清除对话历史请求
type ClearRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// ClearResponse 清除对话历史响应
type ClearResponse struct {
	SessionID string `json:"session_id"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}

// Clear 清除对话历史
func (h *ChatHandler) Clear(c *gin.Context) {
	var req ClearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的清除对话历史请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("清除对话历史", map[string]interface{}{
		"session_id": req.SessionID,
	})

	// 清除对话历史
	h.conversationService.ClearConversation(req.SessionID)

	// 返回响应
	response := ClearResponse{
		SessionID: req.SessionID,
		Success:   true,
		Message:   "对话历史已清除",
	}

	c.JSON(http.StatusOK, response)
}

// InvoiceRequest 发票请求
type InvoiceRequest struct {
	CustomerName string                 `json:"customer_name" binding:"required"`
	CustomerEmail string                `json:"customer_email" binding:"required"`
	Items        []InvoiceItemRequest   `json:"items" binding:"required,min=1"`
}

// InvoiceItemRequest 发票项目请求
type InvoiceItemRequest struct {
	Name     string  `json:"name" binding:"required"`
	Quantity int     `json:"quantity" binding:"required,min=1"`
	Price    float64 `json:"price" binding:"required,min=0"`
}

// InvoiceResponse 发票响应
type InvoiceResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Invoice map[string]interface{} `json:"invoice,omitempty"`
}

// CreateInvoice 创建发票
func (h *ChatHandler) CreateInvoice(c *gin.Context) {
	var req InvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的发票请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("收到创建发票请求", map[string]interface{}{
		"customer_name": req.CustomerName,
		"customer_email": req.CustomerEmail,
		"items_count":   len(req.Items),
	})

	// 准备参数
	params := map[string]interface{}{
		"customer_name":  req.CustomerName,
		"customer_email": req.CustomerEmail,
		"items":          req.Items,
	}

	// 处理发票请求
	result, err := h.conversationService.ProcessInvoiceRequest(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("处理发票请求失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "处理发票请求失败",
		})
		return
	}

	// 返回响应
	response := InvoiceResponse{
		Success: true,
		Message: "发票创建成功",
		Invoice: result,
	}

	c.JSON(http.StatusOK, response)
}

// InvoiceQueryRequest 发票查询请求
type InvoiceQueryRequest struct {
	InvoiceID string `json:"invoice_id" binding:"required"`
}

// InvoiceQueryResponse 发票查询响应
type InvoiceQueryResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Invoice map[string]interface{} `json:"invoice,omitempty"`
}

// QueryInvoice 查询发票
func (h *ChatHandler) QueryInvoice(c *gin.Context) {
	var req InvoiceQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的发票查询请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("收到发票查询请求", map[string]interface{}{
		"invoice_id": req.InvoiceID,
	})

	// 处理发票查询
	result, err := h.conversationService.ProcessInvoiceQuery(c.Request.Context(), req.InvoiceID)
	if err != nil {
		h.logger.Error("处理发票查询失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "处理发票查询失败",
		})
		return
	}

	// 返回响应
	response := InvoiceQueryResponse{
		Success: true,
		Message: "发票查询成功",
		Invoice: result,
	}

	c.JSON(http.StatusOK, response)
}

// ModelListResponse 模型列表响应
type ModelListResponse struct {
	Models []string `json:"models"`
}

// GetModels 获取可用模型列表
func (h *ChatHandler) GetModels(c *gin.Context) {
	h.logger.Info("获取模型列表", nil)

	// 获取模型管理器
	modelManager := h.conversationService.GetModelManager()
	
	// 获取当前提供商
	provider := "openai" // 默认使用openai
	
	// 获取可用模型列表
	models := modelManager.GetAvailableModels(provider)

	// 返回响应
	response := ModelListResponse{
		Models: models,
	}

	c.JSON(http.StatusOK, response)
}

// CurrentModelResponse 当前模型响应
type CurrentModelResponse struct {
	Model string `json:"model"`
}

// GetCurrentModel 获取当前模型
func (h *ChatHandler) GetCurrentModel(c *gin.Context) {
	h.logger.Info("获取当前模型", nil)

	// 获取模型管理器
	modelManager := h.conversationService.GetModelManager()
	
	// 获取当前模型信息
	modelInfo := modelManager.GetCurrentModelInfo()

	// 返回响应
	response := CurrentModelResponse{
		Model: modelInfo["model_name"],
	}

	c.JSON(http.StatusOK, response)
}

// UpdateModelRequest 更新模型请求
type UpdateModelRequest struct {
	Model string `json:"model" binding:"required"`
}

// UpdateModelResponse 更新模型响应
type UpdateModelResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Model   string `json:"model"`
}

// UpdateModel 更新当前模型
func (h *ChatHandler) UpdateModel(c *gin.Context) {
	var req UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的更新模型请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("收到更新模型请求", map[string]interface{}{
		"model": req.Model,
	})

	// 获取模型管理器
	modelManager := h.conversationService.GetModelManager()
	
	// 更新模型
	err := modelManager.UpdateModel("openai", req.Model, "", "") // 使用空字符串表示不更新API密钥和API基础URL
	if err != nil {
		h.logger.Error("更新模型失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新模型失败",
		})
		return
	}

	// 返回响应
	response := UpdateModelResponse{
		Success: true,
		Message: "模型更新成功",
		Model:   req.Model,
	}

	c.JSON(http.StatusOK, response)
}

// PluginListResponse 插件列表响应
type PluginListResponse struct {
	Plugins map[string]interface{} `json:"plugins"`
}

// GetPlugins 获取已加载的插件列表
func (h *ChatHandler) GetPlugins(c *gin.Context) {
	h.logger.Info("获取插件列表", nil)

	// 获取插件管理器
	pluginManager := h.conversationService.GetPluginManager()
	
	// 获取所有插件
	plugins := pluginManager.GetAllPlugins()

	// 转换为响应格式
	pluginList := make(map[string]interface{})
	for name, plugin := range plugins {
		tools := make([]map[string]interface{}, 0, len(plugin.Tools))
		for _, tool := range plugin.Tools {
			tools = append(tools, map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
			})
		}
		
		pluginList[name] = map[string]interface{}{
			"name":  plugin.Name,
			"path":  plugin.Path,
			"tools": tools,
		}
	}

	// 返回响应
	response := PluginListResponse{
		Plugins: pluginList,
	}

	c.JSON(http.StatusOK, response)
}

// PluginFunctionRequest 插件函数请求
type PluginFunctionRequest struct {
	FunctionName string                 `json:"function_name" binding:"required"`
	Params       map[string]interface{} `json:"params"`
}

// PluginFunctionResponse 插件函数响应
type PluginFunctionResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Result  map[string]interface{} `json:"result,omitempty"`
}

// UnloadPluginRequest 卸载插件请求
type UnloadPluginRequest struct {
	PluginName string `json:"plugin_name" binding:"required"`
}

// UnloadPluginResponse 卸载插件响应
type UnloadPluginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UnloadPlugin 卸载插件
func (h *ChatHandler) UnloadPlugin(c *gin.Context) {
	var req UnloadPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的卸载插件请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("收到卸载插件请求", map[string]interface{}{
		"plugin_name": req.PluginName,
	})

	// 获取插件管理器
	pluginManager := h.conversationService.GetPluginManager()
	
	// 卸载插件
	err := pluginManager.UnloadPlugin(req.PluginName)
	if err != nil {
		h.logger.Error("卸载插件失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "卸载插件失败",
		})
		return
	}

	// 返回响应
	response := UnloadPluginResponse{
		Success: true,
		Message: "插件卸载成功",
	}

	c.JSON(http.StatusOK, response)
}

// ExecutePluginFunction 执行插件函数
func (h *ChatHandler) ExecutePluginFunction(c *gin.Context) {
	var req PluginFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的插件函数请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("收到执行插件函数请求", map[string]interface{}{
		"function_name": req.FunctionName,
	})

	// 处理插件函数
	result, err := h.conversationService.ExecutePluginFunction(c.Request.Context(), req.FunctionName, req.Params)
	if err != nil {
		h.logger.Error("执行插件函数失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "执行插件函数失败",
		})
		return
	}

	// 返回响应
	response := PluginFunctionResponse{
		Success: true,
		Message: "插件函数执行成功",
		Result:  result,
	}

	c.JSON(http.StatusOK, response)
}

// ReloadPluginRequest 重新加载插件请求
type ReloadPluginRequest struct {
	PluginName string `json:"plugin_name" binding:"required"`
}

// ReloadPluginResponse 重新加载插件响应
type ReloadPluginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ReloadPlugin 重新加载插件
func (h *ChatHandler) ReloadPlugin(c *gin.Context) {
	var req ReloadPluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("无效的重新加载插件请求", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求格式",
		})
		return
	}

	h.logger.Info("收到重新加载插件请求", map[string]interface{}{
		"plugin_name": req.PluginName,
	})

	// 获取插件管理器
	pluginManager := h.conversationService.GetPluginManager()
	
	// 重新加载插件
	err := pluginManager.ReloadPlugin(req.PluginName)
	if err != nil {
		h.logger.Error("重新加载插件失败", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "重新加载插件失败",
		})
		return
	}

	// 返回响应
	response := ReloadPluginResponse{
		Success: true,
		Message: "插件重新加载成功",
	}

	c.JSON(http.StatusOK, response)
}