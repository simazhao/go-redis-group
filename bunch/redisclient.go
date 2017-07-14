package bunch

import (
	"github.com/simazhao/go-redis-group/untity"
	"github.com/simazhao/go-redis-group/log"
)

type redisClient struct {
	pool *RedisClientPool
	connections [][]*redisConn
}

func NewRedisClient(address string, pool *RedisClientPool) *redisClient{
	client := new(redisClient)
	client.pool = pool

	client.connections = make([][]*redisConn, pool.config.MaxDataBases)

	for db := range client.connections {
		client.connections[db] = make([]*redisConn, pool.config.MaxConnectionsInDataBase)

		for connectno := range client.connections[db] {
			client.connections[db][connectno] = NewRedisConn(address, pool.config.ClientConfig)
		}
	}

	return client
}

func (r *redisClient) Reset() {
	for _, dbconn := range r.connections {
		for _, conn := range dbconn {
			conn.Stop()
		}
	}
}

func (r *redisClient) selectConnection(database int) *redisConn {
	if database < 0 || database > len(r.connections)-1 {
		return nil
	}

	connNo := untity.RandUInt(100000)
	return r.selectConnection2(connNo, database)
}

func (r *redisClient) selectConnection2(rnd uint, database int) *redisConn {
	if database < 0 || database > len(r.connections)-1 {
		return nil
	}

	connNo := rnd
	db := r.connections[database]

	for _ = range db {
		connNo = (connNo + 1) % uint(len(db))

		if conn := db[connNo]; conn.IsRunning() {
			log.Factory.GetLogger().InfoFormat("reqtrace: req select connection #%d of %d\r\n", connNo, len(db))
			return conn
		}
	}

	return nil
}

