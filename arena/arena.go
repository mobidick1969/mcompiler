package arena

import (
	"fmt"
	"unsafe"
)

type SimpleArena struct {
	buffer []byte
	offset int
}

func NewSimpleArena(size int) *SimpleArena {
	return &SimpleArena{
		buffer: make([]byte, size),
		offset: 0,
	}
}

func (a *SimpleArena) Allocate(size int) unsafe.Pointer {
	if a.offset+size > len(a.buffer) {
		panic("OOM: Arena memory exhausted")
	}

	ptr := unsafe.Pointer(&a.buffer[a.offset])
	a.offset += size
	return ptr
}

func (a *SimpleArena) Reset() {
	a.offset = 0
}

func main() {
	myArena := NewSimpleArena(1024 * 1024 * 32) //32MB
	ptr := myArena.Allocate(8)
	val := (*int64)(ptr)
	*val = 1234567890
	fmt.Println(*val)
	myArena.Reset()
	fmt.Println(*val)
}

type BetterArena struct {
	buffer []byte
	offset int
}

func NewBetterArena(size int) *BetterArena {
	return &BetterArena{
		buffer: make([]byte, size),
		offset: 0,
	}
}

func Allocate[T any](a *BetterArena) *T {
	var zero T
	size := int(unsafe.Sizeof(zero))
	align := int(unsafe.Alignof(zero))
	padding := (align - a.offset&(align-1)) & (align - 1)
	if a.offset+padding+size > len(a.buffer) {
		panic("OOM: Arena memory exhausted")
	}
	a.offset += padding
	ptr := unsafe.Pointer(&a.buffer[a.offset])
	a.offset += size
	return (*T)(ptr)
}

func (a *BetterArena) Reset() {
	a.offset = 0
}

const initialChunkSize = 4096

type Chunk struct {
	Data []byte
	Next *Chunk
}

type BestArena struct {
	Head    *Chunk
	Current *Chunk
	Offset  int
}

func NewBestArena() *BestArena {
	c := &Chunk{Data: make([]byte, initialChunkSize)}
	return &BestArena{
		Head:    c,
		Current: c,
		Offset:  0,
	}
}

func Alloc[T any](a *BestArena) *T {
	var zero T
	size := int(unsafe.Sizeof(zero))
	align := int(unsafe.Alignof(zero))

	padding := (align - (a.Offset & (align - 1))) & (align - 1)
	if a.Offset+padding+size > len(a.Current.Data) {
		a.grow(size)
		padding = 0
	}

	a.Offset += padding
	ptr := unsafe.Pointer(&a.Current.Data[a.Offset])
	a.Offset += size
	return (*T)(ptr)
}

func (a *BestArena) grow(requiredSize int) {
	newSize := len(a.Current.Data) * 2
	if requiredSize > newSize {
		newSize = requiredSize
	}

	if a.Current.Next != nil {
		if len(a.Current.Next.Data) >= newSize {
			a.Current = a.Current.Next
			a.Offset = 0
			return
		}
	}
	newChunk := &Chunk{Data: make([]byte, newSize)}
	a.Current.Next = newChunk
	a.Current = newChunk
	a.Offset = 0
}

func (a *BestArena) Reset() {
	a.Current = a.Head
	a.Offset = 0
}

func (a *BestArena) AllocUnsafe(size, align int) unsafe.Pointer {
	padding := (align - (a.Offset & (align - 1))) & (align - 1)
	if a.Offset+padding+size > len(a.Current.Data) {
		a.grow(size)
		padding = 0
	}

	a.Offset += padding
	ptr := unsafe.Pointer(&a.Current.Data[a.Offset])
	a.Offset += size
	return ptr
}
