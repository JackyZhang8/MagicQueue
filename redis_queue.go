/*
Copyright (C) MagicQueue
Author ZYW@
*/
package MagicQueue

import (
	"github.com/go-redis/redis"
)

// RedisQueue Redis队列实现
type RedisQueue struct {
	client *redis.Client
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func (c *RedisConfig) GetDriver() string {
	return "redis"
}

// NewRedisQueue 创建Redis队列实例
func NewRedisQueue(config *RedisConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &RedisQueue{
		client: client,
	}, nil
}

func (q *RedisQueue) Push(queueKey string, message []byte) error {
	return q.client.RPush(queueKey, message).Err()
}

func (q *RedisQueue) Pop(queueKey string) ([]byte, error) {
	result, err := q.client.LPop(queueKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return result, err
}

func (q *RedisQueue) Size(queueKey string) (int64, error) {
	return q.client.LLen(queueKey).Result()
}

func (q *RedisQueue) Clear(queueKey string) error {
	return q.client.Del(queueKey).Err()
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}
