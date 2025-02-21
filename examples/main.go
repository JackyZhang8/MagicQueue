package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"time"
	"MagicQueue"
)

// EmailTask 表示一个邮件任务
type EmailTask struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

// EmailHandler 处理邮件发送任务
type EmailHandler struct{}

func (h *EmailHandler) Execute(payload *MagicQueue.Payload) *MagicQueue.Result {
	var task EmailTask
	err := payload.ParseBody(&task)
	if err != nil {
		return MagicQueue.NewQueueResult(false, fmt.Sprintf("Failed to parse task: %v", err), nil)
	}

	// 模拟发送邮件
	log.Printf("Sending email to: %s, subject: %s", task.To, task.Subject)
	time.Sleep(time.Second) // 模拟网络延迟

	return MagicQueue.NewQueueResult(true, "Email sent successfully", nil)
}

// sendTestEmails 发送测试邮件
func sendTestEmails(queue *MagicQueue.MQueue, count int) {
	for i := 1; i <= count; i++ {
		task := EmailTask{
			To:      fmt.Sprintf("user%d@example.com", i),
			Subject: fmt.Sprintf("Test Email %d", i),
			Content: fmt.Sprintf("This is test email content %d", i),
		}

		err, id := queue.Enqueue(&MagicQueue.Payload{
			Topic:     "email",
			Group:     "notification",
			Body:      task,
			IsPersist: true,
		})

		if err != nil {
			log.Printf("Failed to enqueue email %d: %v", i, err)
		} else {
			log.Printf("Enqueued email %d with ID: %s", i, id)
		}
	}
}

// Example1_RedisQueue 展示如何使用 Redis 队列
func Example1_RedisQueue() {
	log.Println("=== Redis Queue Example ===")
	
	// 初始化 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 没有密码
		DB:       0,  // 使用默认 DB
	})

	// 创建 Redis 队列实例
	queue := MagicQueue.NewQueue("email_service").
		UseRedis(rdb).
		UseLevelDb("./data/queue.db")

	// 注册邮件处理器
	queue.SetHandler("email", "notification", &EmailHandler{})

	// 启动工作者
	go queue.StartWorkers(2)

	// 发送测试邮件
	sendTestEmails(queue, 5)

	// 等待所有邮件发送完成
	time.Sleep(10 * time.Second)
}

// Example2_MemoryQueue 展示如何使用内存队列
func Example2_MemoryQueue() {
	log.Println("=== Memory Queue Example ===")
	
	// 创建内存队列实例
	queue := MagicQueue.NewQueue("email_service").
		UseMemory(&MagicQueue.MemoryConfig{
			MaxQueueSize: 1000, // 设置最大队列大小
		}).
		UseLevelDb("./data/queue.db") // 可选：使用 LevelDB 持久化

	// 注册邮件处理器
	queue.SetHandler("email", "notification", &EmailHandler{})

	// 启动工作者
	go queue.StartWorkers(2)

	// 发送测试邮件
	sendTestEmails(queue, 3)

	// 等待所有邮件发送完成
	time.Sleep(6 * time.Second)
}

func main() {
	// 运行 Redis 队列示例
	Example1_RedisQueue()
	
	log.Println("\nWaiting 5 seconds before starting memory queue example...\n")
	time.Sleep(5 * time.Second)
	
	// 运行内存队列示例
	Example2_MemoryQueue()
}
