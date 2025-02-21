/*
Copyright (C) MagicQueue
Author ZYW@
*/
package MagicQueue

import (
	"container/list"
	"sync"
)

// MemoryQueue 内存队列实现
type MemoryQueue struct {
	queues map[string]*list.List
	mutex  sync.RWMutex
}

// MemoryConfig 内存队列配置
type MemoryConfig struct {
	// 可以添加内存队列特定的配置，比如最大容量等
	MaxQueueSize int // 每个队列的最大容量，0表示无限制
}

func (c *MemoryConfig) GetDriver() string {
	return "memory"
}

// NewMemoryQueue 创建内存队列实例
func NewMemoryQueue(config *MemoryConfig) *MemoryQueue {
	return &MemoryQueue{
		queues: make(map[string]*list.List),
	}
}

func (q *MemoryQueue) getOrCreateQueue(queueKey string) *list.List {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if queue, exists := q.queues[queueKey]; exists {
		return queue
	}
	
	queue := list.New()
	q.queues[queueKey] = queue
	return queue
}

func (q *MemoryQueue) Push(queueKey string, message []byte) error {
	queue := q.getOrCreateQueue(queueKey)
	
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	queue.PushBack(message)
	return nil
}

func (q *MemoryQueue) Pop(queueKey string) ([]byte, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if queue, exists := q.queues[queueKey]; exists && queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		return element.Value.([]byte), nil
	}
	
	return nil, nil
}

func (q *MemoryQueue) Size(queueKey string) (int64, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	
	if queue, exists := q.queues[queueKey]; exists {
		return int64(queue.Len()), nil
	}
	
	return 0, nil
}

func (q *MemoryQueue) Clear(queueKey string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if queue, exists := q.queues[queueKey]; exists {
		queue.Init() // 清空队列
	}
	
	return nil
}

func (q *MemoryQueue) Close() error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	// 清空所有队列
	for _, queue := range q.queues {
		queue.Init()
	}
	q.queues = make(map[string]*list.List)
	
	return nil
}
