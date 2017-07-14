package api

import (
	"github.com/simazhao/go-redis-group/datamodel"
	"github.com/simazhao/go-redis-group/group"
	"github.com/CodisLabs/codis/pkg/utils/errors"
	"github.com/simazhao/go-redis-group/config"
	"time"
	"strconv"
)

type webapi struct {
	names map[string]int
	Groups map[int]*group.Group
}

func (p *webapi) Init(allconfig *config.ForLoad) {
	p.names = make(map[string]int)
	p.Groups = make(map[int]*group.Group)

	for _, groupConfig := range allconfig.Groups {
		group := group.NewGroup(&allconfig.Pool)
		group.FillGroup(groupConfig)
		p.Groups[groupConfig.Id] =  group

		p.names[groupConfig.Name] =  groupConfig.Id
	}
}

func (p *webapi) Close() {
	for _, group1 := range p.Groups {
		group1.Shutdown()
	}
}

func (p *webapi) process(gkey GroupKey, request *datamodel.Request) error {
	var found *group.Group
	var ok bool
	ok = false
	groupId := gkey.GroupId

	if p.Groups == nil {
		return errors.New("group do not init yet")
	}

	if found, ok = p.Groups[groupId]; !ok {
		if groupId, ok = p.names[gkey.GroupName]; !ok {
			return errors.New("cannot find group")
		} else if found, ok = p.Groups[groupId]; !ok {
			return errors.New("cannot find group")
		}
	}

	found.Dispatch(request)

	request.Wait.Wait()
	
	if request.Err != nil {
		return request.Err
	}

	return nil
}

func (p *webapi) Get(gkey GroupKey, key string) (string, error) {
	if request, err := convertGetKeyValue(key); err != nil {
		return "", err
	} else if err := p.process(gkey, request); err != nil {
		return "", err
	} else {
		if len(request.Clips) == 0 || len(request.Clips[0].Value) == 0{
			return "", nil
		}

		return string(request.Clips[0].Value), nil
	}
}

func (p *webapi) Set(gkey GroupKey, key string, val string) (bool, error) {
	if request, err := convertSetKeyValue(key, val); err != nil {
		return false, err
	} else if err := p.process(gkey, request); err != nil {
		return false, err
	} else {
		return len(request.Clips) > 0 && string(request.Clips[0].Value) == "OK", nil
	}
}

func (p *webapi) SetExp(gkey GroupKey, key string, val string, dur string) (bool, error) {
	if n, err := strconv.ParseInt(dur, 10, 0); err != nil {
		return false, errors.New("incorrect duration format")
	} else if time.Duration(n) < time.Second {
		return false, errors.New("duration less than one second")
	} else {
		return p.setExp(gkey, key ,val, time.Duration(n))
	}
}

func (p *webapi) setExp(gkey GroupKey, key string, val string, dur time.Duration) (bool, error){
	if request, err := convertSetKeyValueExpire(key, val, dur); err != nil {
		return false, err
	} else if err := p.process(gkey, request); err != nil {
		return false, err
	} else {
		return len(request.Clips) > 0 && string(request.Clips[0].Value) == "OK", nil
	}
}

