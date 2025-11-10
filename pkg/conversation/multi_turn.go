package conversation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go-smart/pkg/tools"
)

// MultiTurnConversation 多轮对话处理器
type MultiTurnConversation struct {
	manager       *Manager
	orderTool     *tools.QueryOrder
	refundTool    *tools.RefundTool
	chatModel     model.BaseChatModel
}

// NewMultiTurnConversation 创建多轮对话处理器
func NewMultiTurnConversation(
	manager *Manager,
	orderTool *tools.QueryOrder,
	refundTool *tools.RefundTool,
	chatModel model.BaseChatModel,
) *MultiTurnConversation {
	return &MultiTurnConversation{
		manager:    manager,
		orderTool:  orderTool,
		refundTool: refundTool,
		chatModel:  chatModel,
	}
}

// ProcessMessage 处理用户消息
func (m *MultiTurnConversation) ProcessMessage(ctx context.Context, sessionID, userMessage string) (string, error) {
	// 获取或创建会话状态
	_ = m.manager.GetOrCreateState(sessionID, "default_user")
	
	// 添加用户消息到历史
	err := m.manager.stateManager.AddMessage(sessionID, "user", userMessage, nil)
	if err != nil {
		return "", fmt.Errorf("添加用户消息失败: %w", err)
	}
	
	// 获取当前对话步骤
	currentStep, err := m.manager.GetCurrentStep(sessionID)
	if err != nil {
		currentStep = "greeting" // 默认步骤
	}
	
	// 根据当前步骤处理消息
	var response string
	
	switch currentStep {
	case "greeting":
		response, err = m.handleGreetingStep(ctx, sessionID, userMessage)
	case "order_query":
		response, err = m.handleOrderQueryStep(ctx, sessionID, userMessage)
	case "refund_request":
		response, err = m.handleRefundRequestStep(ctx, sessionID, userMessage)
	default:
		// 检测用户意图
		intent := m.detectIntent(userMessage)
		
		switch intent {
		case "order_query":
			response, err = m.startOrderQuery(ctx, sessionID)
		case "refund_request":
			response, err = m.startRefundRequest(ctx, sessionID)
		default:
			// 使用基础对话链处理
			response, err = m.handleGeneralChat(ctx, sessionID, userMessage)
		}
	}
	
	if err != nil {
		return "", err
	}
	
	// 添加助手响应到历史
	err = m.manager.stateManager.AddMessage(sessionID, "assistant", response, nil)
	if err != nil {
		return "", fmt.Errorf("添加助手消息失败: %w", err)
	}
	
	return response, nil
}

// detectIntent 检测用户意图
func (m *MultiTurnConversation) detectIntent(message string) string {
	// 转换为小写以便匹配
	lowerMessage := strings.ToLower(message)
	
	// 订单查询关键词
	orderKeywords := []string{"查订单", "查询订单", "订单状态", "我的订单", "查一下订单", "订单信息"}
	for _, keyword := range orderKeywords {
		if strings.Contains(lowerMessage, keyword) {
			return "order_query"
		}
	}
	
	// 退款申请关键词
	refundKeywords := []string{"退款", "退货", "申请退款", "我要退款", "怎么退款", "退款申请"}
	for _, keyword := range refundKeywords {
		if strings.Contains(lowerMessage, keyword) {
			return "refund_request"
		}
	}
	
	return "general"
}

// handleGreetingStep 处理问候步骤
func (m *MultiTurnConversation) handleGreetingStep(ctx context.Context, sessionID, message string) (string, error) {
	// 检测用户意图
	intent := m.detectIntent(message)
	
	switch intent {
	case "order_query":
		return m.startOrderQuery(ctx, sessionID)
	case "refund_request":
		return m.startRefundRequest(ctx, sessionID)
	default:
		// 通用问候响应
		response := "您好！我是智能客服助手，可以帮助您查询订单信息、处理退款申请等。请问有什么可以帮助您的吗？"
		return response, nil
	}
}

// startOrderQuery 开始订单查询流程
func (m *MultiTurnConversation) startOrderQuery(ctx context.Context, sessionID string) (string, error) {
	// 设置当前步骤为订单查询
	err := m.manager.stateManager.SetCurrentStep(sessionID, "order_query")
	if err != nil {
		return "", fmt.Errorf("设置当前步骤失败: %w", err)
	}
	
	// 检查用户是否已经提供了订单号
	state, exists := m.manager.stateManager.GetState(sessionID)
	if !exists {
		return "", fmt.Errorf("获取状态失败: %w", ErrStateNotFound)
	}
	
	if orderID, exists := state.Context["order_id"]; exists {
		// 如果已有订单号，直接查询
		return m.processOrderQuery(ctx, sessionID, orderID.(string))
	}
	
	// 否则询问订单号
	response := "好的，我可以帮您查询订单信息。请提供您的订单号，通常以'ORD'开头。"
	return response, nil
}

// handleOrderQueryStep 处理订单查询步骤
func (m *MultiTurnConversation) handleOrderQueryStep(ctx context.Context, sessionID, message string) (string, error) {
	// 尝试从消息中提取订单号
	orderID := m.extractOrderID(message)
	
	if orderID == "" {
		// 没有找到订单号，继续询问
		response := "抱歉，我没有找到有效的订单号。请提供您的订单号，通常以'ORD'开头。"
		return response, nil
	}
	
	// 保存订单号到上下文
	err := m.manager.stateManager.SetContext(sessionID, "order_id", orderID)
	if err != nil {
		return "", fmt.Errorf("保存订单号失败: %w", err)
	}
	
	// 处理订单查询
	return m.processOrderQuery(ctx, sessionID, orderID)
}

// processOrderQuery 处理订单查询
func (m *MultiTurnConversation) processOrderQuery(ctx context.Context, sessionID, orderID string) (string, error) {
	// 调用订单查询工具
	orderInfo, err := m.orderTool.Query(ctx, orderID)
	if err != nil {
		// 查询失败
		response := fmt.Sprintf("查询订单失败: %s", err.Error())
		return response, nil
	}
	
	// 格式化订单信息
	formattedInfo := m.orderTool.FormatOrderInfo(orderInfo)
	
	// 重置对话步骤
	err = m.manager.stateManager.SetCurrentStep(sessionID, "greeting")
	if err != nil {
		return "", fmt.Errorf("重置对话步骤失败: %w", err)
	}
	
	// 返回订单信息
	response := fmt.Sprintf("查询成功！以下是您的订单信息：\n\n%s\n\n还有其他可以帮助您的吗？", formattedInfo)
	return response, nil
}

// extractOrderID 从消息中提取订单号
func (m *MultiTurnConversation) extractOrderID(message string) string {
	// 订单号通常以ORD开头，后跟数字
	re := regexp.MustCompile(`[A-Za-z]*\d{6,}`)
	matches := re.FindAllString(message, -1)
	
	for _, match := range matches {
		// 检查是否包含ORD
		if strings.Contains(strings.ToUpper(match), "ORD") {
			return strings.ToUpper(match)
		}
	}
	
	// 如果没有找到ORD开头的，返回第一个匹配项
	if len(matches) > 0 {
		return strings.ToUpper(matches[0])
	}
	
	return ""
}

// startRefundRequest 开始退款申请流程
func (m *MultiTurnConversation) startRefundRequest(ctx context.Context, sessionID string) (string, error) {
	// 设置当前步骤为退款申请
	err := m.manager.stateManager.SetCurrentStep(sessionID, "refund_request")
	if err != nil {
		return "", fmt.Errorf("设置当前步骤失败: %w", err)
	}
	
	// 检查用户是否已经提供了订单号
	state, exists := m.manager.stateManager.GetState(sessionID)
	if !exists {
		return "", fmt.Errorf("获取状态失败: %w", ErrStateNotFound)
	}
	
	if orderID, exists := state.Context["order_id"]; exists {
		// 如果已有订单号，继续下一步
		if reason, exists := state.Context["refund_reason"]; exists {
			// 如果已有退款原因，直接处理退款申请
			return m.processRefundRequest(ctx, sessionID, orderID.(string), reason.(string))
		}
		// 询问退款原因
		response := "好的，您要为订单 " + orderID.(string) + " 申请退款。请告诉我退款原因，例如：商品质量问题、不想要了等。"
		return response, nil
	}
	
	// 否则询问订单号
	response := "好的，我可以帮您处理退款申请。请提供您的订单号，通常以'ORD'开头。"
	return response, nil
}

// handleRefundRequestStep 处理退款申请步骤
func (m *MultiTurnConversation) handleRefundRequestStep(ctx context.Context, sessionID, message string) (string, error) {
	// 获取当前状态
	state, exists := m.manager.stateManager.GetState(sessionID)
	if !exists {
		return "", fmt.Errorf("获取状态失败: %w", ErrStateNotFound)
	}
	
	// 检查是否已有订单号
	if _, exists := state.Context["order_id"]; !exists {
		// 尝试从消息中提取订单号
		orderID := m.extractOrderID(message)
		if orderID == "" {
			// 没有找到订单号，继续询问
			response := "抱歉，我没有找到有效的订单号。请提供您的订单号，通常以'ORD'开头。"
			return response, nil
		}
		
		// 保存订单号到上下文
		err := m.manager.stateManager.SetContext(sessionID, "order_id", orderID)
		if err != nil {
			return "", fmt.Errorf("保存订单号失败: %w", err)
		}
		
		// 询问退款原因
		response := "好的，您要为订单 " + orderID + " 申请退款。请告诉我退款原因，例如：商品质量问题、不想要了等。"
		return response, nil
	}
	
	// 已有订单号，检查是否有退款原因
	if _, exists := state.Context["refund_reason"]; !exists {
		// 保存退款原因
		err := m.manager.stateManager.SetContext(sessionID, "refund_reason", message)
		if err != nil {
			return "", fmt.Errorf("保存退款原因失败: %w", err)
		}
		
		// 处理退款申请
		orderID := state.Context["order_id"].(string)
		return m.processRefundRequest(ctx, sessionID, orderID, message)
	}
	
	// 已有订单号和退款原因，可能是用户在补充信息
	response := "您已经提交了退款申请，正在处理中。请稍等片刻，或者您可以提供新的订单号来申请其他订单的退款。"
	return response, nil
}

// processRefundRequest 处理退款申请
func (m *MultiTurnConversation) processRefundRequest(ctx context.Context, sessionID, orderID, reason string) (string, error) {
	// 调用退款工具
	refundInfo, err := m.refundTool.SubmitRefund(ctx, orderID, reason)
	if err != nil {
		// 退款申请失败
		response := fmt.Sprintf("退款申请失败: %s", err.Error())
		return response, nil
	}
	
	// 格式化退款信息
	formattedInfo := m.refundTool.FormatRefundInfo(refundInfo)
	
	// 重置对话步骤
	err = m.manager.stateManager.SetCurrentStep(sessionID, "greeting")
	if err != nil {
		return "", fmt.Errorf("重置对话步骤失败: %w", err)
	}
	
	// 返回退款申请信息
	response := fmt.Sprintf("退款申请已提交！以下是您的申请信息：\n\n%s\n\n还有其他可以帮助您的吗？", formattedInfo)
	return response, nil
}

// handleGeneralChat 处理通用聊天
func (m *MultiTurnConversation) handleGeneralChat(ctx context.Context, sessionID, message string) (string, error) {
	// 创建对话模板
	chatTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个智能客服助手，可以帮助用户查询订单信息、处理退款申请等。请友好、专业地回答用户的问题。"),
		schema.UserMessage("{query}"),
	)
	
	// 构建对话链
	chain := compose.NewChain[map[string]any, map[string]any]()
	
	// 添加模板
	chain.AppendChatTemplate(chatTemplate)
	
	// 添加聊天模型
	chain.AppendChatModel(m.chatModel)
	
	// 添加输出解析器
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (map[string]any, error) {
		if msg != nil {
			content := msg.Content
			return map[string]any{"response": content}, nil
		}
		return map[string]any{"response": "抱歉，我无法理解您的问题。请尝试重新表述或询问其他问题。"}, nil
	}))
	
	// 编译链
	compiledChain, err := chain.Compile(ctx)
	if err != nil {
		return "", err
	}
	
	// 执行对话链
	result, err := compiledChain.Invoke(ctx, map[string]any{
		"query": message,
	})
	if err != nil {
		return "", err
	}
	
	// 提取响应
	response, ok := result["response"].(string)
	if !ok {
		return "抱歉，我无法理解您的问题。请尝试重新表述或询问其他问题。", nil
	}
	
	return response, nil
}