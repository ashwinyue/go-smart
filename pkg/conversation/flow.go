package conversation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// ConversationStep 对话步骤接口
type ConversationStep interface {
	// Execute 执行步骤
	Execute(ctx context.Context, state *ConversationState, input string) (string, error)
	// CanTransition 判断是否可以转换到此步骤
	CanTransition(state *ConversationState, input string) bool
	// GetName 获取步骤名称
	GetName() string
}

// StepRegistry 步骤注册表
type StepRegistry struct {
	steps map[string]ConversationStep
}

// NewStepRegistry 创建新的步骤注册表
func NewStepRegistry() *StepRegistry {
	return &StepRegistry{
		steps: make(map[string]ConversationStep),
	}
}

// Register 注册步骤
func (r *StepRegistry) Register(step ConversationStep) {
	r.steps[step.GetName()] = step
}

// Get 获取步骤
func (r *StepRegistry) Get(name string) (ConversationStep, bool) {
	step, exists := r.steps[name]
	return step, exists
}

// GetAll 获取所有步骤
func (r *StepRegistry) GetAll() map[string]ConversationStep {
	return r.steps
}

// ConversationFlow 对话流程管理器
type ConversationFlow struct {
	registry *StepRegistry
}

// NewConversationFlow 创建新的对话流程管理器
func NewConversationFlow() *ConversationFlow {
	flow := &ConversationFlow{
		registry: NewStepRegistry(),
	}
	
	// 注册默认步骤
	flow.registerDefaultSteps()
	
	return flow
}

// ProcessInput 处理用户输入
func (f *ConversationFlow) ProcessInput(ctx context.Context, state *ConversationState, input string) (string, error) {
	// 获取当前步骤
	currentStep, exists := f.registry.Get(state.CurrentStep)
	if !exists {
		return "", fmt.Errorf("未找到当前步骤: %s", state.CurrentStep)
	}
	
	// 执行当前步骤
	response, err := currentStep.Execute(ctx, state, input)
	if err != nil {
		return "", fmt.Errorf("执行步骤失败: %w", err)
	}
	
	// 检查是否需要转换到下一步
	for _, step := range f.registry.GetAll() {
		if step.GetName() != state.CurrentStep && step.CanTransition(state, input) {
			state.CurrentStep = step.GetName()
			break
		}
	}
	
	return response, nil
}

// RegisterStep 注册步骤
func (f *ConversationFlow) RegisterStep(step ConversationStep) {
	f.registry.Register(step)
}

// registerDefaultSteps 注册默认步骤
func (f *ConversationFlow) registerDefaultSteps() {
	// 注册问候步骤
	f.RegisterStep(&GreetingStep{})
	
	// 注册订单查询步骤
	f.RegisterStep(&OrderQueryStep{})
	
	// 注册退款申请步骤
	f.RegisterStep(&RefundRequestStep{})
}

// GreetingStep 问候步骤
type GreetingStep struct{}

func (s *GreetingStep) GetName() string {
	return "greeting"
}

func (s *GreetingStep) Execute(ctx context.Context, state *ConversationState, input string) (string, error) {
	// 检查用户意图
	if strings.Contains(input, "查订单") || strings.Contains(input, "订单") {
		state.CurrentStep = "order_query"
		return "请提供您的订单号，以便我为您查询订单信息。", nil
	}
	
	if strings.Contains(input, "退款") || strings.Contains(input, "退单") {
		state.CurrentStep = "refund_request"
		return "请提供您的订单号和退款原因，我将为您处理退款申请。", nil
	}
	
	return "您好！我是智能客服助手，可以帮您查询订单、处理退款等。请问有什么可以帮助您的？", nil
}

func (s *GreetingStep) CanTransition(state *ConversationState, input string) bool {
	// 问候步骤是初始步骤，不需要转换条件
	return false
}

// OrderQueryStep 订单查询步骤
type OrderQueryStep struct{}

func (s *OrderQueryStep) GetName() string {
	return "order_query"
}

func (s *OrderQueryStep) Execute(ctx context.Context, state *ConversationState, input string) (string, error) {
	// 尝试提取订单号
	orderID := extractOrderID(input)
	
	if orderID == "" {
		return "抱歉，我没有找到有效的订单号。请提供您的订单号，格式通常为'ORD'开头的字符串。", nil
	}
	
	// 保存订单号到上下文
	state.Context["order_id"] = orderID
	
	// 模拟查询订单
	orderStatus := queryOrderStatus(orderID)
	
	// 重置步骤为问候
	state.CurrentStep = "greeting"
	
	return fmt.Sprintf("订单查询结果：\n订单号：%s\n%s", orderID, orderStatus), nil
}

func (s *OrderQueryStep) CanTransition(state *ConversationState, input string) bool {
	// 如果当前不是订单查询步骤，且输入包含订单相关关键词，则可以转换
	return state.CurrentStep != "order_query" && 
		(strings.Contains(input, "查订单") || strings.Contains(input, "订单"))
}

// RefundRequestStep 退款申请步骤
type RefundRequestStep struct{}

func (s *RefundRequestStep) GetName() string {
	return "refund_request"
}

func (s *RefundRequestStep) Execute(ctx context.Context, state *ConversationState, input string) (string, error) {
	// 尝试提取订单号
	orderID := extractOrderID(input)
	
	if orderID == "" {
		return "抱歉，我没有找到有效的订单号。请提供您的订单号，格式通常为'ORD'开头的字符串。", nil
	}
	
	// 保存订单号到上下文
	state.Context["order_id"] = orderID
	
	// 尝试提取退款原因
	refundReason := extractRefundReason(input)
	if refundReason == "" {
		state.Context["awaiting_refund_reason"] = true
		return "请说明您的退款原因，例如：商品质量问题、不想要了、发错货等。", nil
	}
	
	// 保存退款原因
	state.Context["refund_reason"] = refundReason
	delete(state.Context, "awaiting_refund_reason")
	
	// 模拟处理退款申请
	refundResult := processRefundRequest(orderID, refundReason)
	
	// 重置步骤为问候
	state.CurrentStep = "greeting"
	
	return fmt.Sprintf("退款申请结果：\n订单号：%s\n%s", orderID, refundResult), nil
}

func (s *RefundRequestStep) CanTransition(state *ConversationState, input string) bool {
	// 如果当前不是退款申请步骤，且输入包含退款相关关键词，则可以转换
	return state.CurrentStep != "refund_request" && 
		(strings.Contains(input, "退款") || strings.Contains(input, "退单"))
}

// extractOrderID 从文本中提取订单号
func extractOrderID(text string) string {
	// 简单的订单号匹配模式，假设订单号是ORD开头的字符串
	re := regexp.MustCompile(`ORD\w+`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// extractRefundReason 从文本中提取退款原因
func extractRefundReason(text string) string {
	// 简单的退款原因提取
	reasons := []string{"质量问题", "不想要了", "发错货", "损坏", "不符合描述"}
	
	for _, reason := range reasons {
		if strings.Contains(text, reason) {
			return reason
		}
	}
	
	return ""
}

// queryOrderStatus 模拟查询订单状态
func queryOrderStatus(orderID string) string {
	// 模拟订单状态
	return "订单状态：已发货\n物流信息：顺丰快递，单号SF123456789\n预计送达：明天"
}

// processRefundRequest 模拟处理退款申请
func processRefundRequest(orderID, reason string) string {
	// 模拟退款处理结果
	return fmt.Sprintf("退款申请已提交\n退款原因：%s\n处理状态：审核中\n预计完成时间：3-5个工作日", reason)
}