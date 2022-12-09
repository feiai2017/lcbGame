package rand

import (
	"github.com/seehuhn/mt19937"
	"math/rand"
	"sync"
	"time"
)

var randPool = sync.Pool{
	New: func() interface{} {
		rng := rand.New(mt19937.New())
		rng.Seed(time.Now().UnixNano())
		return rng
	},
}

// RandomInt [0,n)
func RandomInt(n int) int {
	rng := randPool.Get()
	result := rng.(*rand.Rand).Int()
	randPool.Put(rng)
	return result % n
}

func Shuffle(slice []int) {
	r := randPool.Get().(*rand.Rand)
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}

	randPool.Put(r)
}
