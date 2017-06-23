package group

import(
	"github.com/simazhao/go-redis-group/bunch"
)

type Group struct{
	Id int
	Name string
	Bunches []bunch.Bunch
}
