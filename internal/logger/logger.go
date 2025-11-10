package logger

import (
	"io"
	"log/slog"
	"os"

	"go-smart/internal/config"
)

// Logger 封装slog日志器
type Logger struct {
	*slog.Logger
}

// NewLogger 创建新的日志实例
func NewLogger(cfg *config.LoggerConfig) (*Logger, error) {
	// 设置日志级别
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// 创建日志处理器选项
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// 创建日志处理器
	var handler slog.Handler

	// 如果配置了文件日志
	if cfg.FilePath != "" {
		// 打开日志文件
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		// 创建多路写入器，同时写入文件和控制台
		multiWriter := io.MultiWriter(file, os.Stdout)
		
		// 根据配置的输出格式创建处理器
		if cfg.Format == "json" {
			handler = slog.NewJSONHandler(multiWriter, opts)
		} else {
			handler = slog.NewTextHandler(multiWriter, opts)
		}
	} else {
		// 仅控制台日志，根据配置的输出格式创建处理器
		if cfg.Format == "json" {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}
	}

	// 创建日志器
	logger := slog.New(handler)

	return &Logger{Logger: logger}, nil
}

// DefaultLogger 创建默认日志器
func DefaultLogger() (*Logger, error) {
	cfg := &config.LoggerConfig{
		Level:    "info",
		Format:   "text",
		Output:   "console",
		FilePath: "",
	}

	return NewLogger(cfg)
}

// WithFields 添加字段
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// WithError 添加错误字段
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err),
	}
}

// Info 记录信息日志
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	if fields != nil {
		l.Logger.Info(msg, l.mapToArgs(fields)...)
	} else {
		l.Logger.Info(msg)
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	if fields != nil {
		l.Logger.Debug(msg, l.mapToArgs(fields)...)
	} else {
		l.Logger.Debug(msg)
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	if fields != nil {
		l.Logger.Warn(msg, l.mapToArgs(fields)...)
	} else {
		l.Logger.Warn(msg)
	}
}

// Error 记录错误日志
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	if fields != nil {
		l.Logger.Error(msg, l.mapToArgs(fields)...)
	} else {
		l.Logger.Error(msg)
	}
}

// mapToArgs 将map转换为slog参数
func (l *Logger) mapToArgs(fields map[string]interface{}) []any {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
}

// Sync 同步日志缓冲区
func (l *Logger) Sync() error {
	// slog.Logger 本身没有 Sync 方法，这里提供一个空实现
	// 如果使用文件日志，可以考虑在 Handler 中实现同步
	return nil
}