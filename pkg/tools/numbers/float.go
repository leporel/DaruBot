package numbers

import (
	"math/rand"
	"time"
)

func GetRandFloat(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())
	n := rand.Int63n(int64(max - min))
	r := min + float64(n)
	return r
}
