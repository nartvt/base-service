package repository

import (
	"base-service/internal/infra"
)

type CacheRepository interface{}

type cacheRepositoryImpl struct {
	rd *infra.RedisClient
}

func NewCacheRepository(rd *infra.RedisClient) *cacheRepositoryImpl {
	return &cacheRepositoryImpl{
		rd: rd,
	}
}
