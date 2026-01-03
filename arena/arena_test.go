package arena

import (
	"testing"
	"unsafe"
)

type Data struct {
	A, B int64
	C    float64
}

var Sink *Data

func BenchmarkStandardAllocation(b *testing.B) {
	var temp *Data
	for i := 0; i < b.N; i++ {
		d := &Data{A: 1, B: 2, C: 3.0}
		temp = d
	}
	Sink = temp
}

func BenchmarkArenaAllocation(b *testing.B) {
	arena := NewSimpleArena(1024 * 1024 * 32)

	var temp *Data
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr := arena.Allocate(int(unsafe.Sizeof(Data{})))
		val := (*Data)(ptr)
		val.A = 1
		val.B = 2
		val.C = 3.0
		temp = val
		if i%1024*1024 == 0 {
			arena.Reset()
		}
	}
	Sink = temp
}

func BenchmarkBetterArenaAllocation(b *testing.B) {
	arena := NewBetterArena(1024 * 1024 * 32)

	var temp *Data
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr := Allocate[Data](arena)
		val := (*Data)(ptr)
		val.A = 1
		val.B = 2
		val.C = 3.0
		temp = val
		if i%1000000 == 0 {
			arena.Reset()
		}
	}
	Sink = temp
}

func BenchmarkBestArenaAllocation(b *testing.B) {
	arena := NewBestArena()

	var temp *Data
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr := Alloc[Data](arena)
		val := (*Data)(ptr)
		val.A = 1
		val.B = 2
		val.C = 3.0
		temp = val
		if i%1000000 == 0 {
			arena.Reset()
		}
	}
	Sink = temp
}

func BenchmarkGC_Pressure(b *testing.B) {
	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]*Data, 10000)
			for j := 0; j < 10000; j++ {
				slice[j] = &Data{A: 1}
			}
		}
	})

	arena := NewSimpleArena(1024 * 1024 * 10) // 10MB
	b.Run("Arena", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			arena.Reset()
			for j := 0; j < 10000; j++ {
				ptr := arena.Allocate(int(unsafe.Sizeof(Data{})))
				val := (*Data)(ptr)
				val.A = 1
			}
		}
	})

	betterArena := NewBetterArena(1024 * 1024 * 10) // 10MB
	b.Run("BetterArena", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			betterArena.Reset()
			for j := 0; j < 10000; j++ {
				ptr := Allocate[Data](betterArena)
				val := (*Data)(ptr)
				val.A = 1
			}
		}
	})

	bestArena := NewBestArena()
	b.Run("BestArena", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bestArena.Reset()
			for j := 0; j < 10000; j++ {
				ptr := Alloc[Data](bestArena)
				val := (*Data)(ptr)
				val.A = 1
			}
		}
	})
}
