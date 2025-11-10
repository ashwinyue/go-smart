package main

import (
	"fmt"
)

// AddNumbers 是一个简单的加法函数，符合插件系统的签名要求
func AddNumbers(args map[string]interface{}) (map[string]interface{}, error) {
	// 从参数中获取数值
	a, aOk := args["a"].(float64)
	b, bOk := args["b"].(float64)
	
	if !aOk || !bOk {
		return map[string]interface{}{
			"error": "参数a和b必须是数字",
		}, fmt.Errorf("参数a和b必须是数字")
	}
	
	// 执行加法
	result := a + b
	
	// 返回结果
	return map[string]interface{}{
		"result": result,
		"message": fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result),
	}, nil
}

// MultiplyNumbers 是一个简单的乘法函数
func MultiplyNumbers(args map[string]interface{}) (map[string]interface{}, error) {
	// 从参数中获取数值
	a, aOk := args["a"].(float64)
	b, bOk := args["b"].(float64)
	
	if !aOk || !bOk {
		return map[string]interface{}{
			"error": "参数a和b必须是数字",
		}, fmt.Errorf("参数a和b必须是数字")
	}
	
	// 执行乘法
	result := a * b
	
	// 返回结果
	return map[string]interface{}{
		"result": result,
		"message": fmt.Sprintf("%.2f * %.2f = %.2f", a, b, result),
	}, nil
}