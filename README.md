# MagicQueue

MagicQueue is a powerful Go queue library that provides reliable message queue functionality with persistent storage and automatic recovery mechanisms. It uses Redis as the message queue and LevelDB for persistent storage, ensuring automatic recovery of unprocessed messages in case of system crashes or abnormal exits.

MagicQueue 是一个强大的 Go 语言队列库，提供可靠的消息队列功能，支持持久化存储和自动恢复机制。它使用 Redis 作为消息队列，LevelDB 作为持久化存储，确保在系统崩溃或异常退出时能够自动恢复未处理的消息。

## Features

- High-performance message queue using Redis
- Persistent storage using LevelDB for fault tolerance
- Support for message grouping and topics
- Automatic retry mechanism
- Concurrent processing capability
- Elegant chainable API
- Exception recovery mechanism

## 特性

- 使用 Redis 作为高性能消息队列
- LevelDB 持久化存储，支持故障恢复
- 支持消息分组和主题
- 自动重试机制
- 并发处理能力
- 优雅的链式调用 API
- 异常恢复机制

## 安装

```bash
go get github.com/JackyZhang8/MagicQueue
```

## 快速开始

### 1. 基本使用

```go
package main

import (
    "github.com/go-redis/redis"
    "MagicQueue"
)

func main() {
    // 初始化 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        Password: "",
        DB: 0,
    })

    // 创建队列实例
    queue := MagicQueue.NewQueue("my_queue").
        UseRedis(rdb).
        UseLevelDb("./data/queue.db")

    // 启动工作者
    queue.StartWorkers(2)
}
```

### 2. 定义任务和处理器

```go
// 定义任务结构
type MyTask struct {
    Name    string `json:"name"`
    Data    string `json:"data"`
}

// 实现任务处理器
type MyHandler struct{}

func (h *MyHandler) Execute(payload *MagicQueue.Payload) *MagicQueue.Result {
    var task MyTask
    err := payload.ParseBody(&task)
    if err != nil {
        return MagicQueue.NewResult(false, "Failed to parse task", nil)
    }

    // 处理任务...
    return MagicQueue.NewResult(true, "Task completed", nil)
}
```

### 3. 注册处理器并发送任务

```go
// 注册处理器
queue.SetHandler("mytopic", "mygroup", &MyHandler{})

// 发送任务
task := MyTask{
    Name: "test",
    Data: "hello world",
}

err, id := queue.Enqueue(&MagicQueue.Payload{
    Topic:     "mytopic",
    Group:     "mygroup",
    Body:      task,
    MaxRetry:  3,
    IsPersist: true,
})

if err != nil {
    log.Printf("Failed to enqueue task: %v", err)
} else {
    log.Printf("Task enqueued with ID: %s", id)
}
```

## 完整示例

查看 [examples/main.go](examples/main.go) 获取完整的示例代码，包括邮件发送队列的实现。

## API 文档

### NewQueue(name string) *MQueue
创建新的队列实例。

### MQueue 方法

- `UseRedis(client *redis.Client) *MQueue`: 设置 Redis 客户端
- `UseLevelDb(path string) *MQueue`: 设置 LevelDB 存储路径
- `SetHandler(topic string, group string, handler Queueable) *MQueue`: 注册消息处理器
- `StartWorkers(workerNum int)`: 启动工作者处理消息
- `Enqueue(payload *Payload) (error, string)`: 发送消息到队列
- `GetQueueSize(topic string, group string) int64`: 获取队列大小

### Payload 结构

```go
type Payload struct {
    ID        string      `json:"id"`
    IsPersist bool       `json:"is_persist"`
    Topic     string     `json:"topic"`
    Group     string     `json:"group"`
    Body      interface{} `json:"body"`
    MaxRetry  int        `json:"max_retry"`
    Retry     int        `json:"retry"`
}
```

### Result 结构

```go
type Result struct {
    State   bool        `json:"state"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

## 最佳实践

1. **错误处理**
```go
err, id := queue.Enqueue(&MagicQueue.Payload{...})
if err != nil {
    // 处理错误
}
```

2. **重试机制**
```go
// 设置最大重试次数
payload := &MagicQueue.Payload{
    MaxRetry: 3,
    // ...
}
```

3. **持久化**
```go
// 启用消息持久化
payload := &MagicQueue.Payload{
    IsPersist: true,
    // ...
}
```

4. **并发控制**
```go
// 根据需求设置合适的工作者数量
queue.StartWorkers(runtime.NumCPU())
```

## 注意事项

1. 确保 Redis 服务正在运行
2. 为 LevelDB 存储选择合适的路径并确保有写入权限
3. 合理设置重试次数和工作者数量
4. 在生产环境中添加适当的错误处理和日志记录

## 许可证

Copyright (C) MagicQueue

## 作者

Author JackyZhang8
