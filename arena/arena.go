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

type chunk struct {
	data []byte
	next *chunk
}

type BestArena struct {
	head    *chunk
	current *chunk
	offset  int
}

func NewBestArena() *BestArena {
	c := &chunk{data: make([]byte, initialChunkSize)}
	return &BestArena{
		head:    c,
		current: c,
		offset:  0,
	}
}

func Alloc[T any](a *BestArena) *T {
	var zero T
	size := int(unsafe.Sizeof(zero))
	align := int(unsafe.Alignof(zero))

	padding := (align - (a.offset & (align - 1))) & (align - 1)
	if a.offset+padding+size > len(a.current.data) {
		a.grow(size)
		padding = 0
	}

	a.offset += padding
	ptr := unsafe.Pointer(&a.current.data[a.offset])
	a.offset += size
	return (*T)(ptr)
}

func (a *BestArena) grow(requiredSize int) {
	newSize := len(a.current.data) * 2
	if requiredSize > newSize {
		newSize = requiredSize
	}

	if a.current.next != nil {
		if len(a.current.next.data) >= newSize {
			a.current = a.current.next
			a.offset = 0
			return
		}
	}
	newChunk := &chunk{data: make([]byte, newSize)}
	a.current.next = newChunk
	a.current = newChunk
	a.offset = 0
}

func (a *BestArena) Reset() {
	a.current = a.head
	a.offset = 0
}
