package api

import (
	"github.com/simazhao/go-redis-group/config"
	"encoding/json"
	"os"
	"bufio"
	"bytes"
)

func load(configFile string) (*config.ForLoad, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	buf := new(bytes.Buffer)

	for  {
		if data,_, err := reader.ReadLine(); err != nil{
			break
		} else {
			buf.Write(data)
		}
	}

	if err != nil {
		return nil, err
	}

	allconfig := config.ForLoad{}
	json.Unmarshal(buf.Bytes(), &allconfig)

	return &allconfig, nil
}

func save(config *config.ForLoad, configFile string)  {
	data, err := json.MarshalIndent(*config, "", "    ")
	if err != nil {
		return
	}

	f, err := os.Create(configFile)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(data)
	f.Sync()
}
