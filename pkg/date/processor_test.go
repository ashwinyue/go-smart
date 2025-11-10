package date

import (
	"testing"
	"time"
)

func TestParseRelativeDate(t *testing.T) {
	// 设置一个固定的当前时间用于测试
	fixedTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	dp := NewDateProcessor()
	dp.SetCurrentTime(fixedTime)
	
	tests := []struct {
		name     string
		expr     string
		expected time.Time
		hasError bool
	}{
		{"昨天", "昨天", time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), false},
		{"前天", "前天", time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC), false},
		{"今天", "今天", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), false},
		{"明天", "明天", time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), false},
		{"3天前", "3天前", time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), false},
		{"5天后", "5天后", time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), false},
		{"无效表达式", "无效", time.Time{}, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dp.ParseRelativeDate(tt.expr)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("ParseRelativeDate(%s) expected error but got none", tt.expr)
				}
			} else {
				if err != nil {
					t.Errorf("ParseRelativeDate(%s) unexpected error: %v", tt.expr, err)
				}
				
				if !result.Equal(tt.expected) {
					t.Errorf("ParseRelativeDate(%s) = %v, expected %v", tt.expr, result, tt.expected)
				}
			}
		})
	}
}

func TestExtractDateFromText(t *testing.T) {
	// 设置一个固定的当前时间用于测试
	fixedTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	dp := NewDateProcessor()
	dp.SetCurrentTime(fixedTime)
	
	tests := []struct {
		name         string
		text         string
		expectedDate time.Time
		expectedStr  string
		hasError     bool
	}{
		{"我昨天下的单", "我昨天下的单", time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), "2024-01-14", false},
		{"前天的会议", "前天的会议", time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC), "2024-01-13", false},
		{"今天天气不错", "今天天气不错", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), "2024-01-15", false},
		{"3天前的报告", "3天前的报告", time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), "2024-01-12", false},
		{"5天后的安排", "5天后的安排", time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), "2024-01-20", false},
		{"没有日期", "没有日期", time.Time{}, "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, dateStr, err := dp.ExtractDateFromText(tt.text)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("ExtractDateFromText(%s) expected error but got none", tt.text)
				}
			} else {
				if err != nil {
					t.Errorf("ExtractDateFromText(%s) unexpected error: %v", tt.text, err)
				}
				
				if !date.Equal(tt.expectedDate) {
					t.Errorf("ExtractDateFromText(%s) date = %v, expected %v", tt.text, date, tt.expectedDate)
				}
				
				if dateStr != tt.expectedStr {
					t.Errorf("ExtractDateFromText(%s) dateStr = %v, expected %v", tt.text, dateStr, tt.expectedStr)
				}
			}
		})
	}
}