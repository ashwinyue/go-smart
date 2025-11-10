package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用程序配置
type Config struct {
	Server    ServerConfig   `mapstructure:"server"`
	Logger    LoggerConfig   `mapstructure:"logger"`
	Database  DatabaseConfig `mapstructure:"database"`
	AI        AIConfig       `mapstructure:"ai"`
	PluginsDir string        `mapstructure:"plugins_dir"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Format      string `mapstructure:"format"`
	Output      string `mapstructure:"output"`
	FilePath    string `mapstructure:"file_path"`
	MaxSize     int    `mapstructure:"max_size"`
	MaxAge      int    `mapstructure:"max_age"`
	MaxBackups  int    `mapstructure:"max_backups"`
	Compress    bool   `mapstructure:"compress"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// AIConfig AI模型配置
type AIConfig struct {
	Provider string       `mapstructure:"provider"`
	OpenAI   OpenAIConfig `mapstructure:"openai"`
	Mock     MockConfig   `mapstructure:"mock"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	BaseURL     string  `mapstructure:"base_url"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// MockConfig Mock配置
type MockConfig struct {
	ResponseDelay  string `mapstructure:"response_delay"`
	DefaultResponse string `mapstructure:"default_response"`
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// 设置配置文件路径和名称
	viper.SetConfigFile(configPath)

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认值和环境变量
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("配置文件未找到，使用默认配置: %v\n", err)
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	
	// 打印配置信息用于调试
	fmt.Printf("OpenAI配置 - API Key: %s..., Base URL: %s, Model: %s\n", 
		config.AI.OpenAI.APIKey[:min(len(config.AI.OpenAI.APIKey), 10)], 
		config.AI.OpenAI.BaseURL, 
		config.AI.OpenAI.Model)

	return config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// 日志默认配置
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.filename", "logs/app.log")
	viper.SetDefault("logger.max_size", 100)
	viper.SetDefault("logger.max_age", 30)
	viper.SetDefault("logger.max_backups", 3)
	viper.SetDefault("logger.compress", true)

	// AI默认配置
	viper.SetDefault("ai.provider", "openai")
	viper.SetDefault("ai.temperature", 0.7)
	viper.SetDefault("ai.openai.model", "gpt-3.5-turbo")
	
	// 插件目录默认配置
	viper.SetDefault("plugins_dir", "plugins")
}