package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type DistributedLock struct {
	client *redis.Client
	key    string
	ttl    time.Duration
}

func NewDistributedLock(client *redis.Client, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    fmt.Sprintf("lock:%s", key),
		ttl:    ttl,
	}
}

// TryLock attempts to acquire the lock
func (l *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	result, err := l.client.SetNX(ctx, l.key, "1", l.ttl).Result()
	return result, err
}

// Unlock releases the lock
func (l *DistributedLock) Unlock(ctx context.Context) error {
	return l.client.Del(ctx, l.key).Err()
}

// LockWithFunc acquires lock, executes function, then releases lock
func (l *DistributedLock) LockWithFunc(ctx context.Context, fn func() error) error {
	// Try to acquire lock with retry
	for i := 0; i < 10; i++ {
		ok, err := l.TryLock(ctx)
		if err != nil {
			return err
		}
		if ok {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Execute function
	err := fn()
	
	// Release lock
	l.Unlock(ctx)
	
	return err
}

// NewRedisClient creates a Redis client
func NewRedisClient(host string, port int, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       0,
	})
}
