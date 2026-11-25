package p03

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkMutex(b *testing.B) {
	var mu sync.Mutex
	c := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			c++
			mu.Unlock()
		}
	})
}

func BenchmarkAtomic(b *testing.B) {
	var c atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(1)
		}
	})
}

func BenchmarkRWMutexSoloLecturas(b *testing.B) {
	var mu sync.RWMutex
	c := 42
	var sink int
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			sink = c
			mu.RUnlock()
		}
	})
	_ = sink
}

// Sin sincronizar: NO es correcto, esta aqui solo para ver el suelo del coste.
func BenchmarkSinProteger(b *testing.B) {
	c := 0
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c++
		}
	})
}
