package config

type PoolConfig struct{
	MaxDataBases int
	MaxConnectionsInDataBase int
	ClientConfig ClientConfig
}