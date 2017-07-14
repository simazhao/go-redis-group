package config

import "time"

type ClientConfig struct {
	ConnectionTimeout time.Duration

	BufferSize int

	RequsetChanLength int
}
