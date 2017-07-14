package untity

import (
	"time"
	"math/rand"
)

func RandUInt(max int32) uint {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	u := r.Int31n(max)
	return uint(u)
}
