package random

import (
	"math/rand"
	"sync"
	"time"
)

type Rand struct {
	mu sync.Mutex
	r  *rand.Rand
}

func New() *Rand {
	return &Rand{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (r *Rand) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.r.Intn(n)
}

