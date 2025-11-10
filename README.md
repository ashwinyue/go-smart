# Go Smart

基于Go语言开发的智能对话系统，使用Gin、Viper和Slog构建。

## 项目结构

```
go-smart/
├── cmd/
│   └── server/
│       └── main.go              # 应用程序入口
├── configs/
│   └── config.yaml             # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go           # 配置管理
│   ├── handler/
│   │   └── chat.go             # HTTP处理器
│   ├── logger/
│   │   └── logger.go           # 日志管理
│   ├── model/
│   │   └── service.go          # AI模型服务
│   ├── server/
│   │   └── server.go           # HTTP服务器
│   └── service/
│       └── conversation.go     # 对话服务
├── pkg/
│   ├── chain/
│   │   └── conversation.go     # 对话链实现
│   └── model/
│       └── config.go           # 模型配置
├── go.mod
├── go.sum
└── README.md
```

## 功能特性

- 基于Gin框架的RESTful API
- 支持多种AI模型（OpenAI、讯飞星火、Mock）
- 使用Viper进行配置管理
- 使用Go官方slog进行日志记录
- 支持对话和订单查询功能
- 优雅关闭机制

## 快速开始

### 安装依赖

```bash
go mod download
```

### 配置环境变量

根据需要设置以下环境变量：

```bash
# OpenAI配置
export OPENAI_API_KEY="your-openai-api-key"
export OPENAI_BASE_URL="https://api.openai.com/v1"  # 可选

# 讯飞星火配置
export SPARK_APP_ID="your-spark-app-id"
export SPARK_API_KEY="your-spark-api-key"
export SPARK_API_SECRET="your-spark-api-secret"
```

### 运行应用

```bash
go run cmd/server/main.go
```

应用将在 http://localhost:8080 启动。

## API文档

### 健康检查

```
GET /health
```

### 聊天接口

```
POST /api/v1/chat
Content-Type: application/json

{
  "message": "你好，今天天气怎么样？"
}
```

### 订单查询接口

```
POST /api/v1/order/query
Content-Type: application/json

{
  "query": "查询我的订单"
}
```

### 测试接口

```
GET /api/v1/ping
```

## 配置说明

配置文件位于 `configs/config.yaml`，主要包含以下部分：

- `server`: 服务器配置（端口、模式等）
- `logger`: 日志配置（级别、格式、输出等）
- `ai`: AI模型配置（提供商、API密钥等）
- `database`: 数据库配置
- `app`: 应用程序配置

## 开发指南

### 添加新的AI模型

1. 在 `internal/model/service.go` 中添加新的模型创建逻辑
2. 在 `configs/config.yaml` 中添加相应的配置项
3. 更新 `config.Config` 结构体

### 添加新的API接口

1. 在 `internal/handler/` 中创建新的处理器
2. 在 `internal/server/server.go` 中添加路由
3. 更新 `main.go` 中的依赖注入

## 许可证

MIT License