package infra

import (
	"context"
	"log/slog"

	"base-service/config"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	rd *redis.Client
}

func NewRedisClient(rd *config.RedisConfig) (*RedisClient, error) {
	client, err := initRedis(rd)
	if err != nil {
		return nil, err
	}
	return &RedisClient{rd: client}, nil
}

func initRedis(rd *config.RedisConfig) (*redis.Client, error) {
	opt, err := redis.ParseURL(rd.BuildRedisConnectionString())
	if err != nil {
		return nil, err
	}
	opt.PoolSize = rd.MaxIdle
	opt.DialTimeout = rd.DialTimeout
	opt.ReadTimeout = rd.ReadTimeout
	opt.WriteTimeout = rd.WriteTimeout
	opt.Password = rd.Password
	opt.DB = rd.DB

	client := redis.NewClient(opt)
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	slog.Info(pong)
	return client, nil
}

func (r *RedisClient) Redis() *redis.Client {
	return r.rd
}

func (r *RedisClient) Close() {
	if r.rd != nil {
		r.rd.Close()
	}
}
