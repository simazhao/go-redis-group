package config

type BunchConfig struct{
	Id int
	Primary string
	Backups []string
}
