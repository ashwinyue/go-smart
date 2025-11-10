package service

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go-smart/internal/config"
	"go-smart/internal/logger"
	"go-smart/pkg/conversation"
	"go-smart/pkg/date"
	modelpkg "go-smart/pkg/model"
	"go-smart/pkg/plugin"
	"go-smart/pkg/tools"
)

// ConversationService 对话服务
type ConversationService struct {
	chain              compose.Runnable[map[string]any, map[string]any]
	dateParser         *date.DateProcessor
	logger             *logger.Logger
	multiTurnConv      *conversation.MultiTurnConversation
	conversationMgr    *conversation.Manager
	modelManager       *modelpkg.ModelManager
	pluginManager      *plugin.PluginManager
	invoiceTool        *tools.InvoiceTool
	orderTool          *tools.QueryOrder
	refundTool         *tools.RefundTool
}

// NewConversationService 创建新的对话服务
func NewConversationService(ctx context.Context, chatModel model.BaseChatModel, log *logger.Logger, cfg *config.Config) (*ConversationService, error) {
	// 创建日期处理器
	dateParser := date.NewDateProcessor()
	
	// 创建模型管理器
	modelManager := modelpkg.NewModelManager(cfg, log)
	
	// 创建插件管理器
	pluginManager := plugin.NewPluginManager(log, cfg.PluginsDir)
	
	// 创建对话模板
	chatTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个智能客服助手，专门帮助用户处理订单、发票和退款相关的问题。当前时间是 {current_date}。"),
		schema.UserMessage("{query}"),
	)
	
	// 创建输出解析器
	outputParser := compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (map[string]any, error) {
		content := msg.Content
		
		// 尝试从用户查询中提取日期信息
		extractedDate, dateStr, err := dateParser.ExtractDateFromText(content)
		if err == nil {
			// 如果成功提取到日期，添加到回复中
			formattedDate := dateParser.FormatDate(extractedDate, "2006年01月02日")
			content = fmt.Sprintf("%s\n\n[系统识别的日期: %s (%s)]", content, formattedDate, dateStr)
		}
		
		return map[string]any{
			"response": content,
			"date":     dateStr,
		}, nil
	})
	
	// 构建对话链: Template -> ChatModel -> OutputParser
	chain, err := compose.NewChain[map[string]any, map[string]any]().
		AppendChatTemplate(chatTemplate).
		AppendChatModel(chatModel).
		AppendLambda(outputParser).
		Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译对话链失败: %w", err)
	}
	
	// 创建对话管理器
	conversationMgr := conversation.NewManager()
	
	// 创建工具
	orderTool := tools.NewQueryOrder()
	refundTool := tools.NewRefundTool(orderTool)
	invoiceTool := tools.NewInvoiceTool()
	
	// 创建多轮对话处理器
	multiTurnConv := conversation.NewMultiTurnConversation(
		conversationMgr,
		orderTool,
		refundTool,
		chatModel,
	)
	
	return &ConversationService{
		chain:           chain,
		dateParser:      dateParser,
		logger:          log,
		multiTurnConv:   multiTurnConv,
		conversationMgr: conversationMgr,
		modelManager:    modelManager,
		pluginManager:   pluginManager,
		invoiceTool:     invoiceTool,
		orderTool:       orderTool,
		refundTool:      refundTool,
	}, nil
}

// ProcessMessage 处理用户消息
func (s *ConversationService) ProcessMessage(ctx context.Context, message string) (map[string]any, error) {
	s.logger.Info("处理用户消息", map[string]interface{}{
		"message": message,
	})

	// 准备输入参数
	input := map[string]any{
		"query":        message,
		"current_date": time.Now().Format("2006-01-02"),
	}
	
	// 执行对话链
	result, err := s.chain.Invoke(ctx, input)
	if err != nil {
		s.logger.Error("执行对话链失败", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("执行对话链失败: %w", err)
	}
	
	s.logger.Info("对话链执行成功", map[string]interface{}{
		"response": result["response"],
	})
	
	return result, nil
}

// ProcessMultiTurnMessage 处理多轮对话消息
func (s *ConversationService) ProcessMultiTurnMessage(ctx context.Context, sessionID, message string) (string, error) {
	s.logger.Info("处理多轮对话消息", map[string]interface{}{
		"session_id": sessionID,
		"message":    message,
	})

	// 使用多轮对话处理器处理消息
	response, err := s.multiTurnConv.ProcessMessage(ctx, sessionID, message)
	if err != nil {
		s.logger.Error("处理多轮对话失败", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("处理多轮对话失败: %w", err)
	}
	
	s.logger.Info("多轮对话处理成功", map[string]interface{}{
		"session_id": sessionID,
		"response":   response,
	})
	
	return response, nil
}

// GetConversationHistory 获取对话历史
func (s *ConversationService) GetConversationHistory(sessionID string) ([]schema.Message, error) {
	s.logger.Info("获取对话历史", map[string]interface{}{
		"session_id": sessionID,
	})

	// 获取对话历史
	history, err := s.conversationMgr.GetConversationHistory(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取对话历史失败: %w", err)
	}
	
	// 转换为schema.Message格式
	schemaMessages := make([]schema.Message, 0, len(history))
	for _, msg := range history {
		role := schema.User
		if msg.Role == "assistant" {
			role = schema.Assistant
		} else if msg.Role == "system" {
			role = schema.System
		}
		
		schemaMessages = append(schemaMessages, schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}
	
	return schemaMessages, nil
}

// ClearConversation 清除对话历史
func (s *ConversationService) ClearConversation(sessionID string) {
	s.logger.Info("清除对话历史", map[string]interface{}{
		"session_id": sessionID,
	})

	s.conversationMgr.RemoveConversation(sessionID)
}

// ProcessOrderQuery 处理订单查询
func (s *ConversationService) ProcessOrderQuery(ctx context.Context, query string) (string, error) {
	s.logger.Info("处理订单查询", map[string]interface{}{
		"query": query,
	})

	// 尝试从查询中提取订单号
	orderID := extractOrderID(query)
	
	// 尝试从查询中提取日期信息
	_, dateStr, err := s.dateParser.ExtractDateFromText(query)
	
	// 根据查询内容生成回复
	var response strings.Builder
	
	if strings.Contains(query, "昨天") && err == nil {
		response.WriteString(fmt.Sprintf("您查询的是昨天(%s)的订单信息。\n", dateStr))
	} else if strings.Contains(query, "前天") && err == nil {
		response.WriteString(fmt.Sprintf("您查询的是前天(%s)的订单信息。\n", dateStr))
	} else if strings.Contains(query, "今天") && err == nil {
		response.WriteString(fmt.Sprintf("您查询的是今天(%s)的订单信息。\n", dateStr))
	}
	
	if orderID != "" {
		// 调用订单查询工具获取实际订单信息
		orderInfo, err := s.orderTool.Query(ctx, orderID)
		if err != nil {
			response.WriteString(fmt.Sprintf("查询订单失败: %s\n", err.Error()))
		} else {
			// 格式化订单信息
			formattedInfo := s.orderTool.FormatOrderInfo(orderInfo)
			response.WriteString(formattedInfo)
		}
	} else {
		response.WriteString("请提供您的订单号，以便我为您查询具体的订单信息。\n")
	}
	
	result := response.String()
	s.logger.Info("订单查询处理完成", map[string]interface{}{
		"result": result,
	})
	
	return result, nil
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

// GetModelManager 获取模型管理器
func (s *ConversationService) GetModelManager() *modelpkg.ModelManager {
	return s.modelManager
}

// GetPluginManager 获取插件管理器
func (s *ConversationService) GetPluginManager() *plugin.PluginManager {
	return s.pluginManager
}

// ProcessInvoiceRequest 处理发票请求
func (s *ConversationService) ProcessInvoiceRequest(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("处理发票请求", map[string]interface{}{
		"params": params,
	})
	
	// 从参数中提取所需信息
	customerName, ok := params["customer_name"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少customer_name参数")
	}
	
	customerTaxID, ok := params["customer_tax_id"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少customer_tax_id参数")
	}
	
	// 处理商品列表
	itemsInterface, ok := params["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少items参数或格式不正确")
	}
	
	items := make([]tools.InvoiceItem, 0, len(itemsInterface))
	for _, itemInterface := range itemsInterface {
		itemMap, ok := itemInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("商品项格式不正确")
		}
		
		name, ok := itemMap["name"].(string)
		if !ok {
			return nil, fmt.Errorf("商品缺少name字段")
		}
		
		quantity, ok := itemMap["quantity"].(int)
		if !ok {
			return nil, fmt.Errorf("商品缺少quantity字段")
		}
		
		unitPrice, ok := itemMap["unit_price"].(float64)
		if !ok {
			return nil, fmt.Errorf("商品缺少unit_price字段")
		}
		
		items = append(items, tools.InvoiceItem{
			Name:      name,
			Quantity:  quantity,
			UnitPrice: unitPrice,
		})
	}
	
	// 获取开票日期（可选）
	var issueDate time.Time
	if issueDateStr, ok := params["issue_date"].(string); ok && issueDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", issueDateStr)
		if err != nil {
			return nil, fmt.Errorf("开票日期格式不正确，应为YYYY-MM-DD格式")
		}
		issueDate = parsedDate
	}
	
	// 调用发票工具处理请求
	invoice, err := s.invoiceTool.CreateInvoice(ctx, customerName, customerTaxID, items, issueDate)
	if err != nil {
		s.logger.Error("处理发票请求失败", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("处理发票请求失败: %w", err)
	}
	
	// 将Invoice结构转换为map[string]interface{}
	result := map[string]interface{}{
		"invoice_id":     invoice.InvoiceID,
		"customer_name":  invoice.CustomerName,
		"customer_tax_id": invoice.CustomerTaxID,
		"items":          invoice.Items,
		"issue_date":     invoice.IssueDate,
		"due_date":       invoice.DueDate,
		"subtotal":       invoice.Subtotal,
		"tax_rate":       invoice.TaxRate,
		"tax_amount":     invoice.TaxAmount,
		"total_with_tax": invoice.TotalWithTax,
		"status":         invoice.Status,
	}
	
	s.logger.Info("发票请求处理成功", map[string]interface{}{
		"result": result,
	})
	
	return result, nil
}

// ProcessInvoiceQuery 处理发票查询
func (s *ConversationService) ProcessInvoiceQuery(ctx context.Context, invoiceID string) (map[string]interface{}, error) {
	s.logger.Info("处理发票查询", map[string]interface{}{
		"invoice_id": invoiceID,
	})
	
	// 调用发票工具查询发票
	invoice, err := s.invoiceTool.QueryInvoice(ctx, invoiceID)
	if err != nil {
		s.logger.Error("处理发票查询失败", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("处理发票查询失败: %w", err)
	}
	
	// 将Invoice结构转换为map[string]interface{}
	result := map[string]interface{}{
		"invoice_id":     invoice.InvoiceID,
		"customer_name":  invoice.CustomerName,
		"customer_tax_id": invoice.CustomerTaxID,
		"items":          invoice.Items,
		"issue_date":     invoice.IssueDate,
		"due_date":       invoice.DueDate,
		"subtotal":       invoice.Subtotal,
		"tax_rate":       invoice.TaxRate,
		"tax_amount":     invoice.TaxAmount,
		"total_with_tax": invoice.TotalWithTax,
		"status":         invoice.Status,
	}
	
	s.logger.Info("发票查询处理成功", map[string]interface{}{
		"result": result,
	})
	
	return result, nil
}

// ProcessRefundRequest 处理退款申请
func (s *ConversationService) ProcessRefundRequest(ctx context.Context, orderID, reason string) (map[string]interface{}, error) {
	s.logger.Info("处理退款申请", map[string]interface{}{
		"order_id": orderID,
		"reason":   reason,
	})
	
	// 检查退款资格
	eligible, message, err := s.refundTool.CheckRefundEligibility(ctx, orderID)
	if err != nil {
		s.logger.Error("检查退款资格失败", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("检查退款资格失败: %w", err)
	}
	
	if !eligible {
		s.logger.Info("不符合退款条件", map[string]interface{}{
			"order_id": orderID,
			"message":  message,
		})
		return map[string]interface{}{
			"success": false,
			"message": message,
		}, nil
	}
	
	// 提交退款申请
	refund, err := s.refundTool.SubmitRefund(ctx, orderID, reason)
	if err != nil {
		s.logger.Error("提交退款申请失败", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("提交退款申请失败: %w", err)
	}
	
	// 格式化退款信息
	formattedInfo := s.refundTool.FormatRefundInfo(refund)
	
	s.logger.Info("退款申请处理成功", map[string]interface{}{
		"refund_id": refund.RequestID,
		"order_id":  orderID,
	})
	
	return map[string]interface{}{
		"success": true,
		"refund":  refund,
		"message": "退款申请已提交",
		"details": formattedInfo,
	}, nil
}

// QueryRefundStatus 查询退款状态
func (s *ConversationService) QueryRefundStatus(ctx context.Context, refundID string) (map[string]interface{}, error) {
	s.logger.Info("查询退款状态", map[string]interface{}{
		"refund_id": refundID,
	})
	
	// 查询退款状态
	refund, err := s.refundTool.QueryRefund(ctx, refundID)
	if err != nil {
		s.logger.Error("查询退款状态失败", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("查询退款状态失败: %w", err)
	}
	
	// 格式化退款信息
	formattedInfo := s.refundTool.FormatRefundInfo(refund)
	
	s.logger.Info("退款状态查询成功", map[string]interface{}{
		"refund_id": refundID,
		"status":    refund.Status,
	})
	
	return map[string]interface{}{
		"success": true,
		"refund":  refund,
		"details": formattedInfo,
	}, nil
}

// ExecutePluginFunction 执行插件函数
func (s *ConversationService) ExecutePluginFunction(ctx context.Context, functionName string, params map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("执行插件函数", map[string]interface{}{
		"function": functionName,
		"params":   params,
	})
	
	// 获取插件函数
	function, err := s.pluginManager.GetPluginFunction(functionName)
	if err != nil {
		s.logger.Error("获取插件函数失败", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("获取插件函数失败: %w", err)
	}
	
	// 执行插件函数
	funcType := reflect.TypeOf(function)
	funcValue := reflect.ValueOf(function)
	
	// 检查函数签名
	if funcType.NumIn() != 2 || funcType.NumOut() != 2 {
		return nil, fmt.Errorf("插件函数签名不正确")
	}
	
	// 准备参数
	args := make([]reflect.Value, 2)
	args[0] = reflect.ValueOf(ctx)
	args[1] = reflect.ValueOf(params)
	
	// 调用函数
	results := funcValue.Call(args)
	
	// 处理结果
	if len(results) != 2 {
		return nil, fmt.Errorf("插件函数返回值数量不正确")
	}
	
	// 检查错误
	errInterface := results[1].Interface()
	if errInterface != nil {
		if err, ok := errInterface.(error); ok {
			return nil, err
		}
		return nil, fmt.Errorf("插件函数返回了非错误类型的错误")
	}
	
	// 获取结果
	resultInterface := results[0].Interface()
	if result, ok := resultInterface.(map[string]interface{}); ok {
		s.logger.Info("插件函数执行成功", map[string]interface{}{
			"function": functionName,
			"result":   result,
		})
		return result, nil
	}
	
	return nil, fmt.Errorf("插件函数返回值类型不正确")
}