package date

import (
	"time"
	"regexp"
	"strconv"
	"errors"
)

// DateProcessor 日期处理器，用于解析和计算相对日期
type DateProcessor struct {
	currentTime time.Time
}

// NewDateProcessor 创建新的日期处理器
func NewDateProcessor() *DateProcessor {
	return &DateProcessor{
		currentTime: time.Now(),
	}
}

// SetCurrentTime 设置当前时间（主要用于测试）
func (dp *DateProcessor) SetCurrentTime(t time.Time) {
	dp.currentTime = t
}

// ParseRelativeDate 解析相对日期表达式，返回具体日期
// 支持的表达式：
// - "昨天" -> 昨天的日期
// - "前天" -> 前天的日期
// - "今天" -> 今天的日期
// - "明天" -> 明天的日期
// - "N天前" -> N天前的日期
// - "N天后" -> N天后的日期
func (dp *DateProcessor) ParseRelativeDate(expr string) (time.Time, error) {
	// 处理"昨天"
	if expr == "昨天" {
		return dp.currentTime.AddDate(0, 0, -1), nil
	}
	
	// 处理"前天"
	if expr == "前天" {
		return dp.currentTime.AddDate(0, 0, -2), nil
	}
	
	// 处理"今天"
	if expr == "今天" {
		return dp.currentTime, nil
	}
	
	// 处理"明天"
	if expr == "明天" {
		return dp.currentTime.AddDate(0, 0, 1), nil
	}
	
	// 处理"N天前"模式
	reDaysBefore := regexp.MustCompile(`(\d+)天前`)
	matches := reDaysBefore.FindStringSubmatch(expr)
	if len(matches) == 2 {
		days, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, errors.New("无效的天数")
		}
		return dp.currentTime.AddDate(0, 0, -days), nil
	}
	
	// 处理"N天后"模式
	reDaysAfter := regexp.MustCompile(`(\d+)天后`)
	matches = reDaysAfter.FindStringSubmatch(expr)
	if len(matches) == 2 {
		days, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, errors.New("无效的天数")
		}
		return dp.currentTime.AddDate(0, 0, days), nil
	}
	
	return time.Time{}, errors.New("不支持的日期表达式")
}

// ExtractDateFromText 从文本中提取日期表达式并转换为具体日期
func (dp *DateProcessor) ExtractDateFromText(text string) (time.Time, string, error) {
	// 定义需要匹配的日期表达式模式
	patterns := []string{
		`昨天`,
		`前天`,
		`今天`,
		`明天`,
		`(\d+)天前`,
		`(\d+)天后`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(text) {
			// 提取匹配的日期表达式
			match := re.FindString(text)
			
			// 转换为具体日期
			date, err := dp.ParseRelativeDate(match)
			if err != nil {
				return time.Time{}, "", err
			}
			
			// 返回日期和格式化的日期字符串
			return date, date.Format("2006-01-02"), nil
		}
	}
	
	return time.Time{}, "", errors.New("文本中未找到日期表达式")
}

// FormatDate 格式化日期为指定格式
func (dp *DateProcessor) FormatDate(date time.Time, format string) string {
	if format == "" {
		format = "2006-01-02"
	}
	return date.Format(format)
}