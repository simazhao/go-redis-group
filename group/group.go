package group

import(
	"github.com/simazhao/go-redis-group/bunch"
	"github.com/simazhao/go-redis-group/datamodel"
	"github.com/simazhao/go-redis-group/config"
	"bytes"
	"hash/crc32"
	"github.com/simazhao/go-redis-group/log"
	//"fmt"
)

type Group struct{
	Id int
	Name string
	Bunches []*bunch.Bunch
	pool *bunch.RedisClientPool
}

func NewGroup(config *config.PoolConfig) *Group{
	return &Group{pool:bunch.NewRedisClientPool(config)}
}

func (g *Group) FillGroup(config *config.GroupConfig) {
	g.Id = config.Id
	g.Name = config.Name
	if g.Bunches != nil && len(g.Bunches) > 0 {
		for _, bunch := range g.Bunches {
			bunch.Reset()
		}
	}

	g.Bunches = make([]*bunch.Bunch, len(config.Bunches))

	for i, bunchConfig := range config.Bunches {
		bunch := new(bunch.Bunch)
		bunch.Id = bunchConfig.Id
		bunch.Init(&bunchConfig, g.pool)
		g.Bunches[i] =  bunch
	}
}

func (g *Group) Shutdown() {
	for _, bunch1 := range g.Bunches {
		bunch1.Reset()
	}
}

func (g *Group) Dispatch(request *datamodel.Request) error {
	hash := getHash(request)
	bunchno := hash % uint32(len(g.Bunches))
	bunch := g.Bunches[bunchno]

	log.Factory.GetLogger().InfoFormat("reqtrace: req select bunch #%d\r\n", bunch.Id)

	return bunch.HandleRequest(request, isReadonly(request))
}

func getHashKey(request *datamodel.Request) []byte  {
	index := 1
	return request.Clips[index].Value
}

// copied code
func getHash(request *datamodel.Request) uint32{
	hashKey := getHashKey(request)

	const (
		TagBeg = '{'
		TagEnd = '}'
	)
	if beg := bytes.IndexByte(hashKey, TagBeg); beg >= 0 {
		if end := bytes.IndexByte(hashKey[beg+1:], TagEnd); end >= 0 {
			hashKey = hashKey[beg+1 : beg+1+end]
		}
	}

	return crc32.ChecksumIEEE(hashKey)
}

func isReadonly(request *datamodel.Request) bool {
	cmd := string(request.Clips[0].Value)
	return cmd == datamodel.GET || cmd == datamodel.MGET
}