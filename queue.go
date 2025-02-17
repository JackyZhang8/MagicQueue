/*
Copyright (C) MagicQueue
Author ZYW@
*/
package MagicQueue

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"runtime"
	"strings"
	"time"
)

type Queueable interface {
	Execute(*Payload) *Result
}

type Payload struct {
	ID        string      `json:"id"`
	IsPersist bool        `json:"is_persist"`
	Topic     string      `json:"topic"`
	Group     string      `json:"group"`
	Body      interface{} `json:"body"`
	MaxRetry  int         `json:"max_retry"`
	Retry     int         `json:"retry"`
}

type RecoveryListener func(stack string)

type Result struct {
	State   bool        `json:"state"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewQueueResult(state bool, msg string, data interface{}) *Result {
	return &Result{State: state, Message: msg, Data: data}
}

type MQueue struct {
	db             *redis.Client
	ldb            *leveldb.DB
	MaxRetry       int
	Retry          int
	ReadyQueues    chan Payload
	FallbackQueues chan Payload
	WorkerNum      int
	OnRecovery     RecoveryListener
	Handlers       map[string]Queueable
	Name           string
}

func (r *MQueue) TestQueue() error {
	_, err := r.db.Ping().Result()
	return err
}

func (r *MQueue) RegisterOnInterrupt(listener RecoveryListener) *MQueue {
	r.OnRecovery = listener
	return r
}

func (r *MQueue) SetHandler(topic string, group string, e Queueable) *MQueue {
	name := r.formatHandlerKey(topic, group)
	r.Handlers[name] = e
	return r
}

func (r *MQueue) UseRedis(client *redis.Client) *MQueue {
	r.db = client
	return r
}

func (r *MQueue) UseLevelDb(ldbPath string) *MQueue {
	ldb, err := leveldb.OpenFile(ldbPath, nil)
	if err != nil {
		return r
	}
	r.ldb = ldb
	return r
}

func (r *MQueue) formatQueueKey(topic string, group string) string {
	var name string
	if len(group) > 0 {
		name = fmt.Sprintf("%s_%s::%s", r.Name, group, topic)
	} else {
		name = fmt.Sprintf("%s_%s", r.Name, topic)
	}
	return name
}

func (r *MQueue) GetQueueSize(topic string, group string) int64 {
	cmd := r.db.LLen(r.formatQueueKey(topic, group))
	length, err := cmd.Result()
	if err != nil {
		log.Info(err)
		return 0
	}
	return length
}

func (r *MQueue) dequeue(topic string, group string) (*Payload, []byte, error) {
	var payload Payload
	cmd := r.db.LPop(r.formatQueueKey(topic, group))
	ret, err := cmd.Bytes()
	if err != nil {
		return nil, nil, cmd.Err()
	}
	err = json.Unmarshal(ret, &payload)
	if err != nil {
		return nil, nil, err
	}
	return &payload, ret, nil
}

func (r *MQueue) Enqueue(payload *Payload) (error, string) {
	if len(payload.Topic) <= 0 {
		return errors.New("TopicId can not be empty"), ""
	}
	id, err := uuid.NewUUID()
	if err != nil {
		return err, ""
	}
	payload.ID = id.String()

	payloadStr, err := json.Marshal(payload)
	if err != nil {
		return err, ""
	}

	// 存储到 LevelDB 用于持久化
	if r.ldb != nil {
		err = r.ldb.Put([]byte(payload.ID), payloadStr, nil)
		if err != nil {
			return err, ""
		}
		payload.IsPersist = true
	}

	_, err = r.db.RPush(r.formatQueueKey(payload.Topic, payload.Group), payloadStr).Result()
	if err != nil {
		return err, ""
	}

	return nil, payload.ID
}

func (r *MQueue) recoverPersistentMessages() {
	if r.ldb == nil {
		return
	}
	
	iter := r.ldb.NewIterator(nil, nil)
	defer iter.Release()
	
	for iter.Next() {
		var payload Payload
		ret := iter.Value()
		err := json.Unmarshal(ret, &payload)
		if err != nil {
			log.Info("Failed to unmarshal payload during recovery:", err)
			continue
		}
		
		exists, err := r.db.LRange(r.formatQueueKey(payload.Topic, payload.Group), 0, -1).Result()
		if err != nil {
			log.Info("Failed to check queue during recovery:", err)
			continue
		}
		
		payloadExists := false
		for _, item := range exists {
			var existingPayload Payload
			if err := json.Unmarshal([]byte(item), &existingPayload); err == nil {
				if existingPayload.ID == payload.ID {
					payloadExists = true
					break
				}
			}
		}
		
		if !payloadExists {
			payload.IsPersist = true
			r.ReadyQueues <- payload
			log.Info("Recovered message:", payload.ID)
		}
	}
	
	if err := iter.Error(); err != nil {
		log.Info("Error during recovery iteration:", err)
	}
}

func (r *MQueue) cleanup() {
	if r.ldb != nil {
		r.ldb.Close()
	}
}

func (r *MQueue) startMessageConsumer(topic string, group string) {
	for {
		queueLen := r.GetQueueSize(topic, group)
		if queueLen > 0 {
			for i := 0; i < int(queueLen); i++ {
				payload, serialized, err := r.dequeue(topic, group)
				if err == nil {
					//to persis file
					r.ldb.Put([]byte(payload.ID), serialized, nil)
					log.Info("persis:", payload.ID)
					//to queue
					r.ReadyQueues <- *payload
				} else {
					log.Info(err)
				}
			}
		} else {
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (r *MQueue) formatHandlerKey(topic string, group string) string {
	var handleName string
	if len(topic) > 0 && len(group) > 0 {
		handleName = fmt.Sprintf("%s::%s", group, topic)
	} else if len(topic) > 0 {
		handleName = topic
	}
	return handleName
}

func (r *MQueue) parseHandlerKey(name string) (string, string) {
	index := strings.Index(name, "::")
	if index == -1 {
		return name, ""
	}
	return name[index+2:], name[:index]
}

func (r *MQueue) StartWorkers(workerNum int) {
	if workerNum > 0 {
		r.WorkerNum = workerNum
	}
	defer r.cleanup()
	
	go r.recoverPersistentMessages()
	
	for key := range r.Handlers {
		topic, group := r.parseHandlerKey(key)
		go r.startMessageConsumer(topic, group)
	}
	
	for n := 0; n < r.WorkerNum; n++ {
		go r.processMessages(n)
	}
	
	log.Info("Queue workers started successfully.")
}

func (r *MQueue) processMessages(workerID int) {
	defer func() {
		if err := recover(); err != nil {
			var stacktrace string
			for i := 1; ; i++ {
				_, f, l, got := runtime.Caller(i)
				if !got {
					break
				}
				stacktrace += fmt.Sprintf("%s:%d\n", f, l)
			}
			logMessage := fmt.Sprintf("Trace: %s\n", err)
			logMessage += fmt.Sprintf("\n%s", stacktrace)
			log.Info(logMessage)
			
			if r.OnRecovery != nil {
				r.OnRecovery(logMessage)
			}
		}
	}()
	
	for {
		select {
		case job := <-r.ReadyQueues:
			log.Info("Worker", workerID, "processing job", job.ID)
			
			handler, exists := r.Handlers[r.formatHandlerKey(job.Topic, job.Group)]
			if !exists {
				log.Info("No handler found for topic:", job.Topic, "group:", job.Group)
				continue
			}
			
			result := handler.Execute(&job)
			
			if result.State {
				if job.IsPersist && r.ldb != nil {
					err := r.ldb.Delete([]byte(job.ID), nil)
					if err != nil {
						log.Info("Failed to delete completed job from LevelDB:", err)
					}
				}
			} else {
				if job.Retry < job.MaxRetry {
					job.Retry++
					r.FallbackQueues <- job
					log.Info("Job failed, scheduled for retry:", job.ID)
				} else {
					log.Info("Job failed permanently after max retries:", job.ID)
					if job.IsPersist && r.ldb != nil {
						err := r.ldb.Delete([]byte(job.ID), nil)
						if err != nil {
							log.Info("Failed to delete failed job from LevelDB:", err)
						}
					}
				}
			}
			
		case job := <-r.FallbackQueues:
			time.Sleep(time.Second * time.Duration(job.Retry*2))
			r.ReadyQueues <- job
		}
	}
}

func (r *MQueue) funcName(job Payload) string {
	return r.formatHandlerKey(job.Topic, job.Group)
}

// Payload 结构体添加 ParseBody 方法
func (p *Payload) ParseBody(v interface{}) error {
	jsonData, err := json.Marshal(p.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, v)
}
