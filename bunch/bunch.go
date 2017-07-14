package bunch

import (
	"github.com/simazhao/go-redis-group/config"
	"github.com/simazhao/go-redis-group/datamodel"
	"github.com/simazhao/go-redis-group/untity"
	"github.com/CodisLabs/codis/pkg/utils/errors"
)

type Bunch struct {
	Id int
	Primary *redisClient
	Backups []*redisClient
}

func (b *Bunch) Init(config *config.BunchConfig, pool *RedisClientPool) {
	b.Id = config.Id
	b.Primary = pool.Fetch(config.Primary)
	b.Backups = make([]*redisClient, len(config.Backups))
	for i, backup := range config.Backups {
		b.Backups[i] = pool.Fetch(backup)
	}
}

func (b *Bunch) Reset() {
	if b.Primary != nil {
		b.Primary.Reset()
	}

	if b.Backups != nil {
		for _, backup := range b.Backups {
			backup.Reset()
		}
	}
}

func (b *Bunch) HandleRequest(request *datamodel.Request, readonly bool) error {
	if conn := b.getConn(request, readonly); conn == nil {
		return errors.New("no valid redis to provide")
	} else {
		conn.Put(request)
	}

	return nil
}

func (b *Bunch) getConn(request *datamodel.Request, readonly bool) *redisConn {
	if readonly && len(b.Backups) > 0 {
		bno := untity.RandUInt(100000)

		for _ = range b.Backups {
			bno := (bno + 1) % uint(len(b.Backups))
			if conn := b.Backups[bno].selectConnection2(bno, request.Database); conn != nil {
				return conn
			}
		}
	}

	return b.Primary.selectConnection(request.Database)
}