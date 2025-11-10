package business

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// InvoiceItem 发票项目
type InvoiceItem struct {
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// Invoice 发票
type Invoice struct {
	InvoiceID      string        `json:"invoice_id"`
	CustomerName   string        `json:"customer_name"`
	CustomerTaxID  string        `json:"customer_tax_id"`
	Items          []InvoiceItem `json:"items"`
	IssueDate      time.Time     `json:"issue_date"`
	DueDate        time.Time     `json:"due_date"`
	Subtotal       float64       `json:"subtotal"`
	TaxRate        float64       `json:"tax_rate"`
	TaxAmount      float64       `json:"tax_amount"`
	TotalWithTax   float64       `json:"total_with_tax"`
	Status         string        `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// InvoiceTool 发票工具
type InvoiceTool struct {
	// 模拟数据库
	invoices      map[string]Invoice
	invoiceCounter int
}

// NewInvoiceTool 创建发票工具
func NewInvoiceTool() *InvoiceTool {
	it := &InvoiceTool{
		invoices:      make(map[string]Invoice),
		invoiceCounter: 1000,
	}
	
	// 初始化模拟数据
	it.initMockData()
	
	return it
}

// initMockData 初始化模拟数据
func (it *InvoiceTool) initMockData() {
	now := time.Now()
	
	// 创建一些模拟发票
	invoices := []Invoice{
		{
			InvoiceID:     "INV20231101001",
			CustomerName:  "张三",
			CustomerTaxID: "110101199001011234",
			Items: []InvoiceItem{
				{Name: "智能手表", Quantity: 1, UnitPrice: 1299.00, Total: 1299.00},
				{Name: "手机壳", Quantity: 2, UnitPrice: 49.00, Total: 98.00},
			},
			IssueDate:    now.Add(-72 * time.Hour),
			DueDate:      now.Add(-42 * time.Hour),
			Subtotal:     1397.00,
			TaxRate:      0.13,
			TaxAmount:    181.61,
			TotalWithTax: 1578.61,
			Status:       "已支付",
			CreatedAt:    now.Add(-72 * time.Hour),
			UpdatedAt:    now.Add(-42 * time.Hour),
		},
		{
			InvoiceID:     "INV20231102002",
			CustomerName:  "李四",
			CustomerTaxID: "110101199002022345",
			Items: []InvoiceItem{
				{Name: "蓝牙耳机", Quantity: 1, UnitPrice: 399.00, Total: 399.00},
			},
			IssueDate:    now.Add(-48 * time.Hour),
			DueDate:      now.Add(48 * time.Hour),
			Subtotal:     399.00,
			TaxRate:      0.13,
			TaxAmount:    51.87,
			TotalWithTax: 450.87,
			Status:       "已开具",
			CreatedAt:    now.Add(-48 * time.Hour),
			UpdatedAt:    now.Add(-48 * time.Hour),
		},
	}
	
	for _, invoice := range invoices {
		it.invoices[invoice.InvoiceID] = invoice
	}
}

// GetName 获取工具名称
func (it *InvoiceTool) GetName() string {
	return "invoice_tool"
}

// GetDescription 获取工具描述
func (it *InvoiceTool) GetDescription() string {
	return "创建或查询发票，支持发票开具和状态查询"
}

// GetParameters 获取工具参数
func (it *InvoiceTool) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "操作类型：create（创建发票）或 query（查询发票）",
				"enum":        []string{"create", "query"},
			},
			"invoice_id": map[string]interface{}{
				"type":        "string",
				"description": "发票ID，查询时必需",
			},
			"customer_name": map[string]interface{}{
				"type":        "string",
				"description": "客户名称，创建发票时必需",
			},
			"customer_tax_id": map[string]interface{}{
				"type":        "string",
				"description": "客户税号，创建发票时必需",
			},
			"items": map[string]interface{}{
				"type":        "array",
				"description": "商品列表，创建发票时必需",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "商品名称",
						},
						"quantity": map[string]interface{}{
							"type":        "integer",
							"description": "商品数量",
						},
						"unit_price": map[string]interface{}{
							"type":        "number",
							"description": "商品单价",
						},
					},
					"required": []string{"name", "quantity", "unit_price"},
				},
			},
		},
		"required": []string{"action"},
	}
}

// Call 实现工具调用接口
func (it *InvoiceTool) Call(args map[string]interface{}) (map[string]interface{}, error) {
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
	case "create":
		// 获取创建发票所需参数
		customerName, _ := args["customer_name"].(string)
		customerTaxID, _ := args["customer_tax_id"].(string)
		
		// 处理items参数
		var items []InvoiceItem
		if itemsInterface, ok := args["items"].([]interface{}); ok {
			for _, itemInterface := range itemsInterface {
				if itemMap, ok := itemInterface.(map[string]interface{}); ok {
					item := InvoiceItem{}
					if name, ok := itemMap["name"].(string); ok {
						item.Name = name
					}
					if quantity, ok := itemMap["quantity"].(float64); ok {
						item.Quantity = int(quantity)
					}
					if unitPrice, ok := itemMap["unit_price"].(float64); ok {
						item.UnitPrice = unitPrice
					}
					item.Total = float64(item.Quantity) * item.UnitPrice
					items = append(items, item)
				}
			}
		}
		
		// 创建发票
		invoice, err := it.CreateInvoice(ctx, customerName, customerTaxID, items, time.Time{})
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, err
		}
		
		formattedInfo := it.FormatInvoiceInfo(invoice)
		
		return map[string]interface{}{
			"success":        true,
			"invoice":        invoice,
			"formatted_info": formattedInfo,
		}, nil
		
	case "query":
		// 获取查询发票所需参数
		invoiceID, _ := args["invoice_id"].(string)
		
		// 查询发票
		invoice, err := it.QueryInvoice(ctx, invoiceID)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}, err
		}
		
		formattedInfo := it.FormatInvoiceInfo(invoice)
		
		return map[string]interface{}{
			"success":        true,
			"invoice":        invoice,
			"formatted_info": formattedInfo,
		}, nil
		
	default:
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("不支持的操作: %s", action),
		}, fmt.Errorf("不支持的操作: %s", action)
	}
}

// generateInvoiceID 生成发票ID
func (it *InvoiceTool) generateInvoiceID() string {
	it.invoiceCounter++
	return fmt.Sprintf("INV%s%04d", time.Now().Format("20060102"), it.invoiceCounter)
}

// CreateInvoice 创建发票
func (it *InvoiceTool) CreateInvoice(ctx context.Context, customerName, customerTaxID string, items []InvoiceItem, issueDate time.Time) (*Invoice, error) {
	// 验证输入参数
	if customerName == "" || customerTaxID == "" {
		return nil, fmt.Errorf("客户名称和税号不能为空")
	}
	
	if len(items) == 0 {
		return nil, fmt.Errorf("商品列表不能为空")
	}
	
	// 验证商品信息
	for i := range items {
		item := &items[i]
		if item.Name == "" || item.Quantity <= 0 || item.UnitPrice <= 0 {
			return nil, fmt.Errorf("商品信息不完整，必须包含名称、数量和单价")
		}
		
		// 计算商品总价
		item.Total = float64(item.Quantity) * item.UnitPrice
	}
	
	// 如果未指定开票日期，使用当前日期
	if issueDate.IsZero() {
		issueDate = time.Now()
	}
	
	// 生成发票ID
	invoiceID := it.generateInvoiceID()
	
	// 计算总金额
	subtotal := 0.0
	for _, item := range items {
		subtotal += item.Total
	}
	
	// 计算税额（假设税率为13%）
	taxRate := 0.13
	taxAmount := subtotal * taxRate
	
	// 计算价税合计
	totalWithTax := subtotal + taxAmount
	
	// 创建发票
	invoice := Invoice{
		InvoiceID:     invoiceID,
		CustomerName:  customerName,
		CustomerTaxID: customerTaxID,
		Items:         items,
		IssueDate:     issueDate,
		DueDate:       issueDate.AddDate(0, 0, 30), // 30天后到期
		Subtotal:      subtotal,
		TaxRate:       taxRate,
		TaxAmount:     taxAmount,
		TotalWithTax:  totalWithTax,
		Status:        "已开具",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// 保存发票
	it.invoices[invoiceID] = invoice
	
	// 模拟处理延迟
	time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(200)))
	
	return &invoice, nil
}

// QueryInvoice 查询发票
func (it *InvoiceTool) QueryInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	// 模拟查询延迟
	time.Sleep(time.Millisecond * time.Duration(50+rand.Intn(100)))
	
	invoice, exists := it.invoices[invoiceID]
	if !exists {
		return nil, fmt.Errorf("发票不存在: %s", invoiceID)
	}
	
	return &invoice, nil
}

// UpdateInvoiceStatus 更新发票状态
func (it *InvoiceTool) UpdateInvoiceStatus(ctx context.Context, invoiceID, status string) (*Invoice, error) {
	invoice, exists := it.invoices[invoiceID]
	if !exists {
		return nil, fmt.Errorf("发票不存在: %s", invoiceID)
	}
	
	// 验证状态
	validStatuses := map[string]bool{
		"已开具": true,
		"已发送": true,
		"已支付": true,
		"已作废": true,
	}
	
	if !validStatuses[status] {
		return nil, fmt.Errorf("无效的发票状态: %s", status)
	}
	
	// 更新状态
	invoice.Status = status
	invoice.UpdatedAt = time.Now()
	it.invoices[invoiceID] = invoice
	
	return &invoice, nil
}

// FormatInvoiceInfo 格式化发票信息
func (it *InvoiceTool) FormatInvoiceInfo(invoice *Invoice) string {
	var result string
	
	result += fmt.Sprintf("发票号: %s\n", invoice.InvoiceID)
	result += fmt.Sprintf("客户名称: %s\n", invoice.CustomerName)
	result += fmt.Sprintf("客户税号: %s\n", invoice.CustomerTaxID)
	result += fmt.Sprintf("开票日期: %s\n", invoice.IssueDate.Format("2006-01-02"))
	result += fmt.Sprintf("到期日期: %s\n", invoice.DueDate.Format("2006-01-02"))
	result += fmt.Sprintf("发票状态: %s\n", invoice.Status)
	
	result += "\n商品明细:\n"
	for _, item := range invoice.Items {
		result += fmt.Sprintf("- %s (数量: %d, 单价: %.2f, 小计: %.2f)\n", 
			item.Name, item.Quantity, item.UnitPrice, item.Total)
	}
	
	result += fmt.Sprintf("\n不含税金额: %.2f\n", invoice.Subtotal)
	result += fmt.Sprintf("税率: %.0f%%\n", invoice.TaxRate*100)
	result += fmt.Sprintf("税额: %.2f\n", invoice.TaxAmount)
	result += fmt.Sprintf("价税合计: %.2f\n", invoice.TotalWithTax)
	
	return result
}