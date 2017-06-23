package bunch

type redisClient struct {
	pool *redisClientPool
	connections [][]redisConn
}

func NewRedisClient(address string, pool *redisClientPool) *redisClient{

	return nil
}

