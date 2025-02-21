# MagicQueue

MagicQueue is a powerful Go queue library that provides reliable message queue functionality with persistent storage and automatic recovery mechanisms. It supports both Redis and in-memory queues, with LevelDB for persistent storage, ensuring automatic recovery of unprocessed messages in case of system crashes or abnormal exits.

MagicQueue 是一个强大的 Go 语言队列库，提供可靠的消息队列功能，支持持久化存储和自动恢复机制。它支持 Redis 和内存队列两种实现，并使用 LevelDB 作为持久化存储，确保在系统崩溃或异常退出时能够自动恢复未处理的消息。

## Features

- Multiple queue implementations:
  - High-performance Redis queue
  - Fast in-memory queue for testing or small workloads
- Persistent storage using LevelDB for fault tolerance
- Support for message grouping and topics
- Automatic retry mechanism
- Concurrent processing capability
- Elegant chainable API
- Exception recovery mechanism
- Periodic queue statistics reporting

## 特性

- 多种队列实现：
  - 使用 Redis 作为高性能消息队列
  - 快速的内存队列，适用于测试或小规模工作负载
- LevelDB 持久化存储，支持故障恢复
- 支持消息分组和主题
- 自动重试机制
- 并发处理能力
- 优雅的链式调用 API
- 异常恢复机制
- 定时队列统计报告

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
    // 使用 Redis 队列
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        Password: "",
        DB: 0,
    })

    queue := MagicQueue.NewQueue("my_queue").
        UseRedis(rdb).
        UseLevelDb("./data/queue.db")

    // 或者使用内存队列
    memoryQueue := MagicQueue.NewQueue("my_queue").
        UseMemory(nil).  // 使用默认配置
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

查看 [examples/main.go](examples/main.go) 获取完整的示例代码，包括：
1. Redis 队列示例：展示在生产环境中使用 Redis 作为队列后端
2. 内存队列示例：展示如何使用内存队列进行开发和测试

以下是邮件队列的简化示例代码：

```go
package main

import (
    "MagicQueue"
)

// 定义邮件任务
type EmailTask struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Content string `json:"content"`
}

// 实现邮件处理器
type EmailHandler struct{}

func (h *EmailHandler) Execute(payload *MagicQueue.Payload) *MagicQueue.Result {
    var task EmailTask
    err := payload.ParseBody(&task)
    if err != nil {
        return MagicQueue.NewQueueResult(false, "Failed to parse task", nil)
    }
    // 处理邮件发送...
    return MagicQueue.NewQueueResult(true, "Email sent", nil)
}

func main() {
    // 示例 1：使用 Redis 队列（生产环境推荐）
    redisQueue := MagicQueue.NewQueue("email_service").
        UseRedis(redisClient).
        UseLevelDb("./data/queue.db")
    
    redisQueue.SetHandler("email", "notification", &EmailHandler{})
    go redisQueue.StartWorkers(2)
    
    // 发送任务到 Redis 队列
    redisQueue.Enqueue(&MagicQueue.Payload{
        Topic: "email",
        Group: "notification",
        Body:  EmailTask{To: "user@example.com"},
    })
    
    // 示例 2：使用内存队列（开发/测试环境）
    memoryQueue := MagicQueue.NewQueue("email_service").
        UseMemory(&MagicQueue.MemoryConfig{
            MaxQueueSize: 1000, // 可选：设置最大队列大小
        }).
        UseLevelDb("./data/queue.db") // 可选：使用 LevelDB 持久化
    
    memoryQueue.SetHandler("email", "notification", &EmailHandler{})
    go memoryQueue.StartWorkers(2)
    
    // 发送任务到内存队列
    memoryQueue.Enqueue(&MagicQueue.Payload{
        Topic: "email",
        Group: "notification",
        Body:  EmailTask{To: "test@example.com"},
    })
}
```

## API 文档

### NewQueue(name string) *MQueue
创建新的队列实例。

### MQueue 方法

- `UseRedis(client *redis.Client) *MQueue`: 设置 Redis 客户端作为队列实现
- `UseMemory(config *MemoryConfig) *MQueue`: 使用内存队列实现
- `UseLevelDb(path string) *MQueue`: 设置 LevelDB 存储路径
- `SetHandler(topic string, group string, handler Queueable) *MQueue`: 注册消息处理器
- `StartWorkers(workerNum int)`: 启动工作者处理消息
- `Enqueue(payload *Payload) (error, string)`: 发送消息到队列
- `GetQueueSize(topic string, group string) int64`: 获取队列大小

### 队列实现选择

1. **Redis 队列**
   - 适用于生产环境
   - 支持分布式部署
   - 可靠的消息传递
   - 高性能和可扩展性

2. **内存队列**
   - 适用于开发和测试环境
   - 无需额外依赖
   - 更快的消息处理速度
   - 重启后消息会丢失（除非使用 LevelDB 持久化）

### 队列统计

MagicQueue 会自动每分钟输出队列统计信息，包括：
- 每个主题/分组队列的当前消息数量
- 清晰的统计日志格式

统计信息示例：
```
=== Queue Statistics ===
Queue my_queue_group1::topic1: 42 messages
Queue my_queue_group2::topic2: 13 messages
=====================
```

统计功能会在调用 `StartWorkers()` 时自动启动，无需额外配置。这有助于监控队列的运行状况和及时发现潜在的消息堆积问题。

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

MIT License

Copyright (c) 2025 MagicQueue

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## 作者

Author JackyZhang8
