package chain

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go-smart/pkg/date"
)

// ConversationChain 对话链结构
type ConversationChain struct {
	chain      compose.Runnable[map[string]any, map[string]any]
	dateParser *date.DateProcessor
}

// NewConversationChain 创建新的对话链
func NewConversationChain(ctx context.Context, chatModel model.BaseChatModel) (*ConversationChain, error) {
	// 创建日期处理器
	dateParser := date.NewDateProcessor()
	
	// 创建对话模板
	chatTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个智能客服助手，专门帮助用户处理订单相关的问题。当前时间是 {current_date}。"),
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
	
	return &ConversationChain{
		chain:      chain,
		dateParser: dateParser,
	}, nil
}

// Invoke 执行对话链
func (c *ConversationChain) Invoke(ctx context.Context, query string) (map[string]any, error) {
	// 准备输入参数
	input := map[string]any{
		"query":        query,
		"current_date": time.Now().Format("2006-01-02"),
	}
	
	// 执行对话链
	result, err := c.chain.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("执行对话链失败: %w", err)
	}
	
	return result, nil
}

// ProcessOrderQuery 处理订单查询
func (c *ConversationChain) ProcessOrderQuery(ctx context.Context, query string) (string, error) {
	// 尝试从查询中提取订单号
	orderID := extractOrderID(query)
	
	// 尝试从查询中提取日期信息
	_, dateStr, err := c.dateParser.ExtractDateFromText(query)
	
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
		response.WriteString(fmt.Sprintf("订单号: %s\n", orderID))
		response.WriteString("订单状态: 已发货\n")
		response.WriteString("预计送达: 明天\n")
	} else {
		response.WriteString("请提供您的订单号，以便我为您查询具体的订单信息。\n")
	}
	
	return response.String(), nil
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