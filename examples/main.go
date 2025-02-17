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
		return MagicQueue.NewResult(false, fmt.Sprintf("Failed to parse task: %v", err), nil)
	}

	// 模拟发送邮件
	log.Printf("Sending email to: %s, subject: %s", task.To, task.Subject)
	time.Sleep(time.Second) // 模拟网络延迟

	return MagicQueue.NewResult(true, "Email sent successfully", nil)
}

func main() {
	// 初始化 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 没有密码
		DB:       0,  // 使用默认 DB
	})

	// 创建队列实例
	queue := MagicQueue.NewQueue("email_service").
		UseRedis(rdb).
		UseLevelDb("./data/queue.db")

	// 注册邮件处理器
	queue.SetHandler("email", "notification", &EmailHandler{})

	// 启动工作者
	go queue.StartWorkers(2)

	// 发送一些测试邮件
	for i := 1; i <= 5; i++ {
		task := EmailTask{
			To:      fmt.Sprintf("user%d@example.com", i),
			Subject: fmt.Sprintf("Test Email %d", i),
			Content: fmt.Sprintf("This is test email content %d", i),
		}

		err, id := queue.Enqueue(&MagicQueue.Payload{
			Topic:     "email",
			Group:     "notification",
			Body:      task,
			MaxRetry:  3,
			IsPersist: true,
		})

		if err != nil {
			log.Printf("Failed to enqueue task: %v", err)
		} else {
			log.Printf("Task enqueued with ID: %s", id)
		}

		time.Sleep(time.Second * 2) // 间隔发送
	}

	// 保持程序运行
	select {}
}
