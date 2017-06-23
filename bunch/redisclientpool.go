package bunch

import (
	"github.com/simazhao/go-redis-group/config"
)

type redisClientPool struct{
	config *config.PoolConfig
	pool map[string]*redisClient
}

func NewRedisClientPool(config *config.PoolConfig) *redisClientPool {
	pool := new(redisClientPool)
	pool.config = config
	pool.pool = make(map[string]*redisClient)
	return pool
}

func (pool *redisClientPool) Get(address string) *redisClient {
	return pool.pool[address]
}

func (pool *redisClientPool) Fetch(address string) *redisClient {
	if client := pool.Get(address); client != nil {
		return client
	} else if client := pool.rawFetch(address); client != nil {
		pool.pool[address] = client
		return client
	}

	return nil
}

func (pool *redisClientPool) rawFetch(address string) *redisClient {
	return NewRedisClient(address, pool)
}