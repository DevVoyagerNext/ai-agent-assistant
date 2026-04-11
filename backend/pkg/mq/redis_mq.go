package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Message 消息结构体
type Message struct {
	Topic   string      `json:"topic"`   // 消息主题/队列名
	Payload interface{} `json:"payload"` // 消息负载
	Retry   int         `json:"retry"`   // 重试次数
}

// Handler 任务处理函数定义
type Handler func(ctx context.Context, payload string) error

// RedisMQ Redis消息队列
type RedisMQ struct {
	client   *redis.Client
	logger   *zap.Logger
	handlers map[string]Handler
}

func NewRedisMQ(client *redis.Client, logger *zap.Logger) *RedisMQ {
	return &RedisMQ{
		client:   client,
		logger:   logger,
		handlers: make(map[string]Handler),
	}
}

// Register 注册主题处理函数
func (m *RedisMQ) Register(topic string, handler Handler) {
	m.handlers[topic] = handler
}

// Publish 发送消息到队列
func (m *RedisMQ) Publish(ctx context.Context, topic string, payload interface{}) error {
	msg := Message{
		Topic:   topic,
		Payload: payload,
		Retry:   0,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message error: %v", err)
	}

	// 使用 LPUSH 将消息推入列表左侧
	key := fmt.Sprintf("mq:queue:%s", topic)
	return m.client.LPush(ctx, key, data).Err()
}

// Start 启动消费者监听
func (m *RedisMQ) Start(ctx context.Context) {
	for topic, handler := range m.handlers {
		// 为每个 topic 开启一个或多个协程消费
		go m.consume(ctx, topic, handler)
	}
}

func (m *RedisMQ) consume(ctx context.Context, topic string, handler Handler) {
	key := fmt.Sprintf("mq:queue:%s", topic)
	m.logger.Info("MQ 消费者启动", zap.String("topic", topic))

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("MQ 消费者停止", zap.String("topic", topic))
			return
		default:
			// 使用 BRPOP 阻塞式从列表右侧弹出消息，超时时间设为 5 秒
			result, err := m.client.BRPop(ctx, 5*time.Second, key).Result()
			if err != nil {
				// redis.Nil 表示超时没拿到数据，继续循环即可
				continue
			}

			// result[0] 是 key 名，result[1] 是消息内容
			payloadStr := result[1]

			// 执行处理逻辑
			if err := handler(ctx, payloadStr); err != nil {
				m.logger.Error("MQ 任务处理失败",
					zap.String("topic", topic),
					zap.Error(err),
					zap.String("payload", payloadStr),
				)
			}
		}
	}
}
