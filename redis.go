package redis

import (
	"context"
	"github.com/go-tron/config"
	"github.com/go-tron/redis/script"
	"github.com/redis/go-redis/v9"
	"time"
)

const Nil = redis.Nil

func NewScript(src string) *redis.Script {
	return redis.NewScript(src)
}

type Config struct {
	Addr         string `json:"addr"`
	Password     string `json:"password"`
	Database     int    `json:"database"`
	PoolSize     int    `json:"poolSize"`
	MinIdleConns int    `json:"minIdleConns"`
}

func NewWithConfig(c *config.Config) *Redis {
	return New(&Config{
		Addr:         c.GetString("redis.addr"),
		Password:     c.GetString("redis.password"),
		Database:     c.GetInt("redis.database"),
		PoolSize:     c.GetInt("redis.poolSize"),
		MinIdleConns: c.GetInt("redis.minIdleConns"),
	})
}

func New(config *Config) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.PoolSize,
	})

	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		panic(err)
	}
	return &Redis{
		client,
	}
}

type Redis struct {
	*redis.Client
}

func (r *Redis) GetDel(ctx context.Context, key string) (string, error) {
	return script.GetDel.Run(ctx, r, []string{key}).Text()
}

func (r *Redis) HGetDel(ctx context.Context, key string, field string) (string, error) {
	return script.HGetDel.Run(ctx, r, []string{key, field}).Text()
}

func (r *Redis) Lock(ctx context.Context, key string, expireInSeconds int) bool {
	s, err := r.SetNX(ctx, key, 1, time.Second*time.Duration(expireInSeconds)).Result()
	if err != nil {
		return false
	}
	return s
}

func (r *Redis) Unlock(ctx context.Context, key string) bool {
	n, err := r.Del(ctx, key).Uint64()
	if err != nil {
		return false
	}
	return n == uint64(1)
}

func (r *Redis) IncrExpire(ctx context.Context, key string, expiration time.Duration) (int, error) {
	return script.IncrExpire.Run(ctx, r, []string{key}, int(expiration/time.Second)).Int()
}

func (r *Redis) HIncrLimit(ctx context.Context, key string, field string, incr int, max int, interval time.Duration) (int, error) {
	return script.HIncrLimit.Run(ctx, r, []string{key, field}, incr, max, int(interval/time.Second)).Int()
}

func (r *Redis) FrequencyLimit(ctx context.Context, key string, max int, interval time.Duration) (int, error) {
	return script.FrequencyLimit.Run(ctx, r, []string{key}, max, int(interval/time.Second)).Int()
}

func (r *Redis) BatchGet(ctx context.Context, key string, fields ...interface{}) (interface{}, error) {
	return script.BatchGet.Run(ctx, r, []string{key}, fields...).Result()
}

func (r *Redis) BatchLock(ctx context.Context, keys []string, expireInSeconds int) (string, error) {
	result, err := script.BatchLock.Run(ctx, r, keys, expireInSeconds).Text()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (r *Redis) BatchUnlock(ctx context.Context, keys []string) bool {
	_, err := script.BatchUnlock.Run(ctx, r, keys).Result()
	return err != nil
}

func (r *Redis) LockWithSecret(ctx context.Context, key string, secret string, expireInSeconds int) bool {
	s, err := r.SetNX(ctx, key, secret, time.Second*time.Duration(expireInSeconds)).Result()
	if err != nil {
		return false
	}
	return s
}

func (r *Redis) UnlockWithSecret(ctx context.Context, key string, secret string) bool {
	v, err := r.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return false
	}
	if v != secret {
		return false
	}
	n, err := r.Del(ctx, key).Uint64()
	if err != nil {
		return false
	}
	return n == uint64(1)
}

func (r *Redis) HDelIsKeyDeleted(ctx context.Context, key string, field string) bool {
	keyDeleted, err := script.HDelIsKeyDeleted.Run(ctx, r, []string{key}, field).Int()
	if err != nil {
		return false
	}
	return keyDeleted == 1
}
