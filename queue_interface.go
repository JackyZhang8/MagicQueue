/*
Copyright (C) MagicQueue
Author ZYW@
*/
package MagicQueue

// QueueDriver 定义队列驱动接口
type QueueDriver interface {
	// Push 将消息推入队列
	Push(queueKey string, message []byte) error
	
	// Pop 从队列中取出消息
	Pop(queueKey string) ([]byte, error)
	
	// Size 获取队列大小
	Size(queueKey string) (int64, error)
	
	// Clear 清空队列
	Clear(queueKey string) error
	
	// Close 关闭队列连接
	Close() error
}

// QueueConfig 队列配置接口
type QueueConfig interface {
	// GetDriver 获取队列驱动类型
	GetDriver() string
}
