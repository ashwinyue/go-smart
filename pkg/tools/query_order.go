package tools

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// OrderInfo 订单信息
type OrderInfo struct {
	OrderID      string    `json:"order_id"`
	Status       string    `json:"status"`
	CreateTime   time.Time `json:"create_time"`
	PayTime      time.Time `json:"pay_time"`
	ShipTime     time.Time `json:"ship_time"`
	ProductList  []Product `json:"product_list"`
	TotalAmount  float64   `json:"total_amount"`
	ShipAddress  string    `json:"ship_address"`
	TrackingInfo string    `json:"tracking_info"`
	EstDelivery  time.Time `json:"est_delivery"`
}

// Product 商品信息
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// QueryOrder 订单查询工具
type QueryOrder struct {
	// 模拟数据库
	orders map[string]OrderInfo
}

// NewQueryOrder 创建订单查询工具
func NewQueryOrder() *QueryOrder {
	q := &QueryOrder{
		orders: make(map[string]OrderInfo),
	}
	
	// 初始化模拟数据
	q.initMockData()
	
	return q
}

// initMockData 初始化模拟数据
func (q *QueryOrder) initMockData() {
	now := time.Now()
	
	// 创建一些模拟订单
	orders := []OrderInfo{
		{
			OrderID:     "ORD123456",
			Status:      "已发货",
			CreateTime:  now.Add(-72 * time.Hour),
			PayTime:     now.Add(-71 * time.Hour),
			ShipTime:    now.Add(-24 * time.Hour),
			ProductList: []Product{
				{ID: "P001", Name: "智能手表", Price: 1299.00, Quantity: 1},
				{ID: "P002", Name: "手机壳", Price: 49.00, Quantity: 2},
			},
			TotalAmount:  1397.00,
			ShipAddress:  "北京市朝阳区某某街道123号",
			TrackingInfo: "顺丰快递，单号SF123456789",
			EstDelivery:  now.Add(24 * time.Hour),
		},
		{
			OrderID:     "ORD789012",
			Status:      "已送达",
			CreateTime:  now.Add(-120 * time.Hour),
			PayTime:     now.Add(-119 * time.Hour),
			ShipTime:    now.Add(-96 * time.Hour),
			ProductList: []Product{
				{ID: "P003", Name: "蓝牙耳机", Price: 399.00, Quantity: 1},
			},
			TotalAmount:  399.00,
			ShipAddress:  "上海市浦东新区某某路456号",
			TrackingInfo: "顺丰快递，单号SF987654321",
			EstDelivery:  now.Add(-48 * time.Hour),
		},
		{
			OrderID:     "ORD345678",
			Status:      "待发货",
			CreateTime:  now.Add(-12 * time.Hour),
			PayTime:     now.Add(-11 * time.Hour),
			ProductList: []Product{
				{ID: "P004", Name: "平板电脑", Price: 2999.00, Quantity: 1},
				{ID: "P005", Name: "保护膜", Price: 29.00, Quantity: 3},
			},
			TotalAmount:  3086.00,
			ShipAddress:  "广州市天河区某某大道789号",
			TrackingInfo: "暂无物流信息",
			EstDelivery:  now.Add(48 * time.Hour),
		},
	}
	
	for _, order := range orders {
		q.orders[order.OrderID] = order
	}
}

// Query 查询订单
func (q *QueryOrder) Query(ctx context.Context, orderID string) (*OrderInfo, error) {
	// 模拟查询延迟
	time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(200)))
	
	order, exists := q.orders[orderID]
	if !exists {
		return nil, fmt.Errorf("订单不存在: %s", orderID)
	}
	
	return &order, nil
}

// FormatOrderInfo 格式化订单信息
func (q *QueryOrder) FormatOrderInfo(order *OrderInfo) string {
	var result string
	
	result += fmt.Sprintf("订单号: %s\n", order.OrderID)
	result += fmt.Sprintf("订单状态: %s\n", order.Status)
	result += fmt.Sprintf("下单时间: %s\n", order.CreateTime.Format("2006-01-02 15:04:05"))
	
	if !order.PayTime.IsZero() {
		result += fmt.Sprintf("支付时间: %s\n", order.PayTime.Format("2006-01-02 15:04:05"))
	}
	
	if !order.ShipTime.IsZero() {
		result += fmt.Sprintf("发货时间: %s\n", order.ShipTime.Format("2006-01-02 15:04:05"))
	}
	
	result += "\n商品列表:\n"
	for _, product := range order.ProductList {
		result += fmt.Sprintf("- %s (数量: %d, 单价: %.2f)\n", product.Name, product.Quantity, product.Price)
	}
	
	result += fmt.Sprintf("\n订单总额: %.2f\n", order.TotalAmount)
	result += fmt.Sprintf("收货地址: %s\n", order.ShipAddress)
	
	if order.TrackingInfo != "" {
		result += fmt.Sprintf("物流信息: %s\n", order.TrackingInfo)
	}
	
	if !order.EstDelivery.IsZero() {
		if order.EstDelivery.After(time.Now()) {
			result += fmt.Sprintf("预计送达: %s\n", order.EstDelivery.Format("2006-01-02"))
		} else {
			result += fmt.Sprintf("送达时间: %s\n", order.EstDelivery.Format("2006-01-02"))
		}
	}
	
	return result
}

// GetToolInfo 获取工具信息
func (q *QueryOrder) GetToolInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "query_order",
		"description": "查询订单信息，包括订单状态、物流信息等",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"order_id": map[string]interface{}{
					"type":        "string",
					"description": "订单号，通常以'ORD'开头",
				},
			},
			"required": []string{"order_id"},
		},
	}
}