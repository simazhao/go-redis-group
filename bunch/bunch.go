package bunch

import "github.com/simazhao/go-redis-group/config"

type Bunch struct {
	Id int
	Primary *redisClient
	Backups []*redisClient
}

func (b *Bunch) Init(config *config.BunchConfig, pool *redisClientPool) {
	b.Id = config.Id
	b.Primary = pool.Fetch(config.Primary)
	b.Backups = make([]*redisClient, len(config.Backups))
	for i, bk := range config.Backups {
		b.Backups[i] = pool.Fetch(bk)
	}
}