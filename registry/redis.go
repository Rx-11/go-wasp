package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Rx-11/go-wasp/config"
	redis "github.com/redis/go-redis/v9"
)

type RedisRegistry struct {
	mu     sync.RWMutex
	client redis.Client
}

func NewRedisRegistry() *RedisRegistry {
	client := redis.NewClient(&redis.Options{Addr: config.Cfg.RedisAddr, Password: config.Cfg.RedisPassword, DB: config.Cfg.RedisDB})
	return &RedisRegistry{client: *client}
}

func (m *RedisRegistry) SaveFunction(name string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.client.Set(context.Background(), name, data, time.Duration(config.Cfg.DefaultTTL)*time.Second)
	return nil
}

func (m *RedisRegistry) GetFunction(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, err := m.client.Get(context.Background(), name).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (m *RedisRegistry) ListFunctions() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys, err := m.client.Keys(context.Background(), "*").Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}
