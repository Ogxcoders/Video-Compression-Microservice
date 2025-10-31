package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yourusername/video-compressor/internal/models"
)

const (
	QueueKey          = "compression:queue"
	ProcessingKey     = "compression:processing"
	ProcessingJobsKey = "compression:processing:jobs"
)

type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisQueue(redisURL string) (*RedisQueue, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisQueue{
		client: client,
		ctx:    ctx,
	}, nil
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}

func (q *RedisQueue) Enqueue(jobID string, priority int) error {
	score := float64(time.Now().Unix())
	if priority > 5 {
		score -= float64(priority * 1000)
	}

	err := q.client.ZAdd(q.ctx, QueueKey, &redis.Z{
		Score:  score,
		Member: jobID,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

func (q *RedisQueue) Dequeue() (string, error) {
	result, err := q.client.ZPopMin(q.ctx, QueueKey, 1).Result()
	if err != nil {
		return "", fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) == 0 {
		return "", nil
	}

	jobID := result[0].Member.(string)

	err = q.client.SAdd(q.ctx, ProcessingJobsKey, jobID).Err()
	if err != nil {
		return "", fmt.Errorf("failed to mark job as processing: %w", err)
	}

	return jobID, nil
}

func (q *RedisQueue) MarkComplete(jobID string) error {
	return q.client.SRem(q.ctx, ProcessingJobsKey, jobID).Err()
}

func (q *RedisQueue) GetQueueLength() (int64, error) {
	return q.client.ZCard(q.ctx, QueueKey).Result()
}

func (q *RedisQueue) GetProcessingCount() (int64, error) {
	return q.client.SCard(q.ctx, ProcessingJobsKey).Result()
}

func (q *RedisQueue) RemoveJob(jobID string) error {
	err := q.client.ZRem(q.ctx, QueueKey, jobID).Err()
	if err != nil {
		return err
	}
	return q.client.SRem(q.ctx, ProcessingJobsKey, jobID).Err()
}

func (q *RedisQueue) CacheJobStatus(jobID string, status *models.StatusResponse, ttl time.Duration) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("job:status:%s", jobID)
	return q.client.Set(q.ctx, key, data, ttl).Err()
}

func (q *RedisQueue) GetCachedJobStatus(jobID string) (*models.StatusResponse, error) {
	key := fmt.Sprintf("job:status:%s", jobID)
	data, err := q.client.Get(q.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var status models.StatusResponse
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}

	return &status, nil
}
