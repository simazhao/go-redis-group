package api

import (
	"testing"
	"github.com/simazhao/go-redis-group/config"
	"time"
	"path"
)

func TestSaveSample(t *testing.T) {
	singleBunchIps :=[]string {
		"10.0.30.135:6379",
		"10.0.30.140:6379",
		"10.0.30.170:6379",
		"10.0.11.115:6379",
		"10.0.13.153:6379",
		"10.0.10.107:6379",
	}

	singleBunchIps2 :=[]string {
		"192.168.144.128:6379",
	}

	singleBunchIps = singleBunchIps2

	config1 := &config.ForLoad{}
	config1.Pool = config.PoolConfig{MaxDataBases:4, MaxConnectionsInDataBase:10,
		ClientConfig:config.ClientConfig{
			ConnectionTimeout:time.Second * 30,
			BufferSize:4096,
			RequsetChanLength:1024,
		}}

	config1.Groups = make([]*config.GroupConfig, 1)
	config1.Groups[0] = &config.GroupConfig{
		Id:1,Name:"default",
		Bunches:make([]config.BunchConfig, len(singleBunchIps)),
	}



	for i, singleip := range singleBunchIps {
		config1.Groups[0].Bunches[i] = config.BunchConfig{
			Id:i+1,
			Primary:singleip,
			Backups:make([]string, 0),
		}
	}

	save(config1, "d:\\all.config")
}

func TestLoadSample(t *testing.T) {
	if config, err := load(path.Join("d:\\",configfilename)); err == nil {
		print(config.Show())
	} else {
		t.Error("failed to load")
	}
}