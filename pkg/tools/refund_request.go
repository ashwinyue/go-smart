package tools

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// RefundRequest 退款申请
type RefundRequest struct {
	OrderID      string    `json:"order_id"`
	Reason       string    `json:"reason"`
	Amount       float64   `json:"amount"`
	RequestTime  time.Time `json:"request_time"`
	Status       string    `json:"status"`
	RequestID    string    `json:"request_id"`
	ProcessTime  time.Time `json:"process_time"`
	Response     string    `json:"response"`
}

// RefundTool 退款工具
type RefundTool struct {
	// 模拟数据库
	refunds map[string]RefundRequest
	// 订单查询工具
	orderTool *QueryOrder
}

// NewRefundTool 创建退款工具
func NewRefundTool(orderTool *QueryOrder) *RefundTool {
	r := &RefundTool{
		refunds:   make(map[string]RefundRequest),
		orderTool: orderTool,
	}
	
	// 初始化模拟数据
	r.initMockData()
	
	return r
}

// initMockData 初始化模拟数据
func (r *RefundTool) initMockData() {
	now := time.Now()
	
	// 创建一些模拟退款申请
	refunds := []RefundRequest{
		{
			OrderID:     "ORD123456",
			Reason:      "商品质量问题",
			Amount:      1299.00,
			RequestTime: now.Add(-48 * time.Hour),
			Status:      "已批准",
			RequestID:   "REF001",
			ProcessTime: now.Add(-24 * time.Hour),
			Response:    "退款已批准，将在3-5个工作日内原路退回您的支付账户",
		},
		{
			OrderID:     "ORD789012",
			Reason:      "不想要了",
			Amount:      399.00,
			RequestTime: now.Add(-12 * time.Hour),
			Status:      "处理中",
			RequestID:   "REF002",
			ProcessTime: time.Time{},
			Response:    "",
		},
	}
	
	for _, refund := range refunds {
		r.refunds[refund.RequestID] = refund
	}
}

// CheckRefundEligibility 检查退款资格
func (r *RefundTool) CheckRefundEligibility(ctx context.Context, orderID string) (bool, string, error) {
	// 查询订单信息
	order, err := r.orderTool.Query(ctx, orderID)
	if err != nil {
		return false, "", fmt.Errorf("查询订单失败: %v", err)
	}
	
	// 检查订单状态
	switch order.Status {
	case "已送达":
		// 已送达的订单，检查是否在7天内
		if time.Since(order.EstDelivery) > 7*24*time.Hour {
			return false, "订单已超过7天退货期", nil
		}
		return true, "订单在退货期内，可以申请退款", nil
	case "已发货":
		return true, "订单已发货但未送达，可以申请退款", nil
	case "待发货":
		return true, "订单未发货，可以直接取消订单退款", nil
	case "已取消":
		return false, "订单已取消，无法再次退款", nil
	default:
		return false, "当前订单状态不支持退款", nil
	}
}

// SubmitRefund 提交退款申请
func (r *RefundTool) SubmitRefund(ctx context.Context, orderID, reason string) (*RefundRequest, error) {
	// 检查退款资格
	eligible, message, err := r.CheckRefundEligibility(ctx, orderID)
	if err != nil {
		return nil, err
	}
	
	if !eligible {
		return nil, fmt.Errorf("不符合退款条件: %s", message)
	}
	
	// 查询订单信息
	order, err := r.orderTool.Query(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %v", err)
	}
	
	// 生成退款ID
	refundID := fmt.Sprintf("REF%d", rand.Intn(100000))
	
	// 创建退款申请
	refund := RefundRequest{
		OrderID:     orderID,
		Reason:      reason,
		Amount:      order.TotalAmount,
		RequestTime: time.Now(),
		Status:      "处理中",
		RequestID:   refundID,
		ProcessTime: time.Time{},
		Response:    "",
	}
	
	// 保存到模拟数据库
	r.refunds[refundID] = refund
	
	// 模拟处理延迟
	time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(200)))
	
	return &refund, nil
}

// ProcessRefund 处理退款申请
func (r *RefundTool) ProcessRefund(ctx context.Context, refundID string) (*RefundRequest, error) {
	refund, exists := r.refunds[refundID]
	if !exists {
		return nil, fmt.Errorf("退款申请不存在: %s", refundID)
	}
	
	if refund.Status != "处理中" {
		return nil, fmt.Errorf("退款申请已处理，当前状态: %s", refund.Status)
	}
	
	// 模拟处理过程
	time.Sleep(time.Millisecond * time.Duration(200+rand.Intn(300)))
	
	// 更新状态
	rand.Seed(time.Now().UnixNano())
	approved := rand.Intn(10) > 2 // 80%概率批准
	
	if approved {
		refund.Status = "已批准"
		refund.Response = "退款已批准，将在3-5个工作日内原路退回您的支付账户"
	} else {
		refund.Status = "已拒绝"
		refund.Response = "抱歉，根据退款政策，您的申请不符合退款条件"
	}
	
	refund.ProcessTime = time.Now()
	r.refunds[refundID] = refund
	
	return &refund, nil
}

// QueryRefund 查询退款状态
func (r *RefundTool) QueryRefund(ctx context.Context, refundID string) (*RefundRequest, error) {
	refund, exists := r.refunds[refundID]
	if !exists {
		return nil, fmt.Errorf("退款申请不存在: %s", refundID)
	}
	
	return &refund, nil
}

// FormatRefundInfo 格式化退款信息
func (r *RefundTool) FormatRefundInfo(refund *RefundRequest) string {
	var result string
	
	result += fmt.Sprintf("退款申请号: %s\n", refund.RequestID)
	result += fmt.Sprintf("关联订单号: %s\n", refund.OrderID)
	result += fmt.Sprintf("退款金额: %.2f\n", refund.Amount)
	result += fmt.Sprintf("申请原因: %s\n", refund.Reason)
	result += fmt.Sprintf("申请时间: %s\n", refund.RequestTime.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("处理状态: %s\n", refund.Status)
	
	if !refund.ProcessTime.IsZero() {
		result += fmt.Sprintf("处理时间: %s\n", refund.ProcessTime.Format("2006-01-02 15:04:05"))
	}
	
	if refund.Response != "" {
		result += fmt.Sprintf("处理结果: %s\n", refund.Response)
	}
	
	return result
}

// GetName 获取工具名称
func (r *RefundTool) GetName() string {
	return "refund_request"
}

// GetDescription 获取工具描述
func (r *RefundTool) GetDescription() string {
	return "处理退款申请，包括提交退款申请、查询退款状态等"
}

// GetParameters 获取工具参数
func (r *RefundTool) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "操作类型：submit（提交退款申请）或 query（查询退款状态）",
				"enum":        []string{"submit", "query"},
			},
			"order_id": map[string]interface{}{
				"type":        "string",
				"description": "订单号，提交退款申请时必需",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "退款原因，提交退款申请时必需",
			},
			"refund_id": map[string]interface{}{
				"type":        "string",
				"description": "退款申请号，查询退款状态时必需",
			},
		},
		"required": []string{"action"},
	}
}

// Call 实现工具调用接口
func (r *RefundTool) Call(args map[string]interface{}) (map[string]interface{}, error) {
	// 获取action参数
	action, ok := args["action"].(string)
	if !ok {
		return map[string]interface{}{
			"success": false,
			"error":   "缺少action参数",
		}, fmt.Errorf("缺少action参数")
	}
	
	ctx := context.Background()
	
	switch action {
	case "submit":
		// 获取提交退款申请所需参数
		orderID, _ := args["order_id"].(string)
		reason, _ := args["reason"].(string)
		
		// 提交退款申请
		refund, err := r.SubmitRefund(ctx, orderID, reason)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, err
		}
		
		formattedInfo := r.FormatRefundInfo(refund)
		
		return map[string]interface{}{
			"success":        true,
			"refund":         refund,
			"formatted_info": formattedInfo,
		}, nil
		
	case "query":
		// 获取查询退款状态所需参数
		refundID, _ := args["refund_id"].(string)
		
		// 查询退款状态
		refund, err := r.QueryRefund(ctx, refundID)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, err
		}
		
		formattedInfo := r.FormatRefundInfo(refund)
		
		return map[string]interface{}{
			"success":        true,
			"refund":         refund,
			"formatted_info": formattedInfo,
		}, nil
		
	default:
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("不支持的操作: %s", action),
		}, fmt.Errorf("不支持的操作: %s", action)
	}
}