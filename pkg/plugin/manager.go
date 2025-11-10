package plugin

import (
	"context"
	"fmt"
	"path/filepath"
	"plugin"
	"reflect"
	"sync"

	"go-smart/internal/logger"
)

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Function    interface{} `json:"-"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Tools       []ToolInfo  `json:"tools"`
	Plugin      interface{} `json:"-"`
}

// PluginManager 插件管理器
type PluginManager struct {
	plugins         map[string]*PluginInfo
	pluginFunctions map[string]interface{}
	mu              sync.RWMutex
	logger          *logger.Logger
	pluginsDir      string
}

// NewPluginManager 创建插件管理器
func NewPluginManager(log *logger.Logger, pluginsDir string) *PluginManager {
	pm := &PluginManager{
		plugins:         make(map[string]*PluginInfo),
		pluginFunctions: make(map[string]interface{}),
		logger:          log,
		pluginsDir:      pluginsDir,
	}
	
	// 加载所有插件
	pm.LoadAllPlugins()
	
	return pm
}

// LoadAllPlugins 加载所有插件
func (pm *PluginManager) LoadAllPlugins() {
	pm.logger.Info("开始加载插件", map[string]interface{}{
		"plugins_dir": pm.pluginsDir,
	})
	
	// 获取插件目录下的所有.so文件
	files, err := filepath.Glob(filepath.Join(pm.pluginsDir, "*.so"))
	if err != nil {
		pm.logger.Error("获取插件文件失败", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	// 加载每个插件
	for _, file := range files {
		pluginName := filepath.Base(file)
		pluginName = pluginName[:len(pluginName)-3] // 移除.so扩展名
		
		if err := pm.LoadPlugin(pluginName, file); err != nil {
			pm.logger.Error("加载插件失败", map[string]interface{}{
				"plugin": pluginName,
				"error":  err.Error(),
			})
		}
	}
	
	pm.logger.Info("插件加载完成", map[string]interface{}{
		"count": len(pm.plugins),
	})
}

// LoadPlugin 加载单个插件
func (pm *PluginManager) LoadPlugin(pluginName, pluginPath string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 检查插件是否已加载
	if _, exists := pm.plugins[pluginName]; exists {
		return fmt.Errorf("插件 %s 已加载", pluginName)
	}
	
	// 加载插件
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("打开插件失败: %w", err)
	}
	
	// 查找插件实例
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("查找插件入口失败: %w", err)
	}
	
	// 类型断言
	newPluginFunc, ok := symbol.(func() interface{})
	if !ok {
		return fmt.Errorf("无效的插件入口函数类型")
	}
	
	// 创建插件实例
	pluginInstance := newPluginFunc()
	
	// 查找插件中的工具函数
	tools := pm.extractTools(pluginInstance)
	
	// 保存插件信息
	pluginInfo := &PluginInfo{
		Name:   pluginName,
		Path:   pluginPath,
		Tools:  tools,
		Plugin: p,
	}
	
	pm.plugins[pluginName] = pluginInfo
	
	// 保存工具函数引用
	for _, tool := range tools {
		funcName := fmt.Sprintf("%s.%s", pluginName, tool.Name)
		pm.pluginFunctions[funcName] = tool.Function
	}
	
	pm.logger.Info("插件加载成功", map[string]interface{}{
		"plugin":      pluginName,
		"tools_count": len(tools),
	})
	
	return nil
}

// extractTools 从插件实例中提取工具函数
func (pm *PluginManager) extractTools(pluginInstance interface{}) []ToolInfo {
	var tools []ToolInfo
	
	// 获取插件实例的反射值
	val := reflect.ValueOf(pluginInstance)
	
	// 如果是指针，获取其指向的值
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	// 遍历所有方法
	for i := 0; i < val.NumMethod(); i++ {
		method := val.Method(i)
		methodType := method.Type()
		
		// 检查方法是否符合工具函数的要求
		// 1. 公共方法（首字母大写）
		// 2. 返回值只有一个，类型为func(context.Context, map[string]interface{}) (map[string]interface{}, error)
		if methodType.NumOut() == 1 {
			outType := methodType.Out(0)
			if outType.Kind() == reflect.Func {
				// 获取函数签名
				funcType := outType
				if funcType.NumIn() == 2 && 
				   funcType.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() &&
				   funcType.In(1) == reflect.TypeOf(map[string]interface{}{}) &&
				   funcType.NumOut() == 2 &&
				   funcType.Out(0) == reflect.TypeOf(map[string]interface{}{}) &&
				   funcType.Out(1) == reflect.TypeOf((*error)(nil)).Elem() {
					
					// 获取方法名
					methodName := val.Type().Method(i).Name
					
					// 创建工具信息
					tool := ToolInfo{
						Name:        methodName,
						Description: pm.getMethodDescription(pluginInstance, methodName),
						Function:    method.Interface(),
					}
					
					tools = append(tools, tool)
				}
			}
		}
	}
	
	return tools
}

// getMethodDescription 获取方法描述
func (pm *PluginManager) getMethodDescription(pluginInstance interface{}, methodName string) string {
	// 这里可以通过反射或注释解析获取方法描述
	// 暂时返回默认描述
	return fmt.Sprintf("工具函数: %s", methodName)
}

// GetPlugin 获取指定插件
func (pm *PluginManager) GetPlugin(pluginName string) (*PluginInfo, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	plugin, exists := pm.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("插件 %s 不存在", pluginName)
	}
	
	return plugin, nil
}

// GetAllPlugins 获取所有插件
func (pm *PluginManager) GetAllPlugins() map[string]*PluginInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	// 创建副本以避免并发问题
	result := make(map[string]*PluginInfo)
	for k, v := range pm.plugins {
		result[k] = v
	}
	
	return result
}

// GetPluginFunction 获取指定的插件函数
func (pm *PluginManager) GetPluginFunction(functionName string) (interface{}, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	function, exists := pm.pluginFunctions[functionName]
	if !exists {
		return nil, fmt.Errorf("插件函数 %s 不存在", functionName)
	}
	
	return function, nil
}

// ReloadPlugin 重新加载指定插件
func (pm *PluginManager) ReloadPlugin(pluginName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 检查插件是否存在
	pluginInfo, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("插件 %s 不存在", pluginName)
	}
	
	// 保存旧的工具函数名
	oldFunctionNames := make([]string, 0, len(pluginInfo.Tools))
	for _, tool := range pluginInfo.Tools {
		funcName := fmt.Sprintf("%s.%s", pluginName, tool.Name)
		oldFunctionNames = append(oldFunctionNames, funcName)
	}
	
	// 删除旧的插件函数引用
	for _, funcName := range oldFunctionNames {
		delete(pm.pluginFunctions, funcName)
	}
	
	// 重新加载插件
	if err := pm.LoadPlugin(pluginName, pluginInfo.Path); err != nil {
		return fmt.Errorf("重新加载插件失败: %w", err)
	}
	
	pm.logger.Info("插件重新加载成功", map[string]interface{}{
		"plugin": pluginName,
	})
	
	return nil
}

// ReloadAllPlugins 重新加载所有插件
func (pm *PluginManager) ReloadAllPlugins() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 保存所有插件名称
	pluginNames := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		pluginNames = append(pluginNames, name)
	}
	
	// 清空当前插件
	pm.plugins = make(map[string]*PluginInfo)
	pm.pluginFunctions = make(map[string]interface{})
	
	// 重新加载所有插件
	successCount := 0
	failedPlugins := make([]string, 0)
	
	for _, pluginName := range pluginNames {
		pluginPath := filepath.Join(pm.pluginsDir, pluginName+".so")
		if err := pm.LoadPlugin(pluginName, pluginPath); err != nil {
			failedPlugins = append(failedPlugins, pluginName)
			pm.logger.Error("重新加载插件失败", map[string]interface{}{
				"plugin": pluginName,
				"error":  err.Error(),
			})
		} else {
			successCount++
		}
	}
	
	pm.logger.Info("插件重新加载完成", map[string]interface{}{
		"success_count": successCount,
		"total_count":   len(pluginNames),
		"failed_count":  len(failedPlugins),
	})
	
	if len(failedPlugins) > 0 {
		return fmt.Errorf("部分插件重新加载失败: %v", failedPlugins)
	}
	
	return nil
}

// UnloadPlugin 卸载指定插件
func (pm *PluginManager) UnloadPlugin(pluginName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 检查插件是否存在
	pluginInfo, exists := pm.plugins[pluginName]
	if !exists {
		return fmt.Errorf("插件 %s 不存在", pluginName)
	}
	
	// 删除插件函数引用
	for _, tool := range pluginInfo.Tools {
		funcName := fmt.Sprintf("%s.%s", pluginName, tool.Name)
		delete(pm.pluginFunctions, funcName)
	}
	
	// 删除插件
	delete(pm.plugins, pluginName)
	
	pm.logger.Info("插件卸载成功", map[string]interface{}{
		"plugin": pluginName,
	})
	
	return nil
}