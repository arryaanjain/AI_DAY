package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisQueue implements the queue.Queue interface using Redis.
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue creates a new Redis queue.
func NewRedisQueue(addr string) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisQueue{client: client}, nil
}

// Publish adds a message to the queue under the specified topic.
func (q *RedisQueue) Publish(ctx context.Context, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Use a Redis list (queue) with topic as the key
	queueKey := fmt.Sprintf("queue:%s", msg.Topic)
	if err := q.client.LPush(ctx, queueKey, string(data)).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Receive retrieves and removes a message from the specified topic.
// It blocks for up to 30 seconds waiting for a message.
func (q *RedisQueue) Receive(ctx context.Context, topic string) (Message, error) {
	queueKey := fmt.Sprintf("queue:%s", topic)

	// Use BRPOP to block and wait for messages
	result, err := q.client.BRPop(ctx, 30*time.Second, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return Message{}, fmt.Errorf("no messages available")
		}
		return Message{}, fmt.Errorf("failed to receive message: %w", err)
	}

	if len(result) < 2 {
		return Message{}, fmt.Errorf("unexpected redis response")
	}

	var msg Message
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return msg, nil
}

// Acknowledge marks a message as processed (Redis queue has no concept of acknowledgment, so this is a no-op).
func (q *RedisQueue) Acknowledge(ctx context.Context, msg Message) error {
	// Redis list doesn't have native acknowledgment; message is already removed by BRPOP
	// In a production system with durability requirements, you might use Redis Streams instead
	return nil
}
