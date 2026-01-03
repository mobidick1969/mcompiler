package main

import (
	"fmt"
	"math/bits"
	"mcompiler/arena"
	"unsafe"
)

type NodeType uint8

const (
	Null NodeType = iota
	False
	True
	Number
	String
	Object
	Array
)

const (
	lsb = 0x0101010101010101
	msb = 0x8080808080808080
)

type Node struct {
	ValueStr string
	Key      string
	Children *Node
	Next     *Node
	Type     NodeType
}

func hasQuote(v uint64) uint64 {
	lo := uint64(lsb)
	hi := uint64(msb)
	diff := v ^ (0x22 * lo)
	return (diff - lo) & (^diff) & hi
}

type QuoteScanner struct {
	data []byte
}

func NewQuoteScanner(input []byte) *QuoteScanner {
	return &QuoteScanner{
		data: input,
	}
}

func (s *QuoteScanner) FindQuoteFast(startIdx int) int {
	limit := len(s.data) - 8
	for i := startIdx; i <= limit; i += 8 {
		// This conversion is safe because we are taking the address of an element
		// in the slice which is live at this point.
		val := *(*uint64)(unsafe.Pointer(&s.data[i]))
		xor := val ^ 0x2222222222222222
		hasZero := (xor - lsb) & (^xor) & msb
		if hasZero != 0 {
			zeros := bits.TrailingZeros64(hasZero)
			idx := zeros >> 3
			return i + idx
		}
	}
	for i := startIdx; i < len(s.data); i++ {
		if s.data[i] == '"' {
			return i
		}
	}
	return -1
}

type Parser struct {
	input  []byte
	cursor int
	arena  *arena.BestArena
}

func NewParser(input []byte, arena *arena.BestArena) *Parser {
	return &Parser{
		input:  input,
		cursor: 0,
		arena:  arena,
	}
}

func (p *Parser) skipWhitespace() {
	for p.cursor < len(p.input) {
		c := p.input[p.cursor]
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			p.cursor++
		} else {
			break
		}
	}
}

func (p *Parser) ParseAny() *Node {
	p.skipWhitespace()
	if p.cursor >= len(p.input) {
		panic("Unexpected EOF")
	}
	char := p.input[p.cursor]

	switch char {
	case '"':
		return p.ParseString()
	case '{':
		return p.ParseObject()
	case '[':
		return p.ParseArray()
	case 't', 'f', 'n', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return p.parsePrimitive()
	default:
		panic("Unexpected character: " + string(char))
	}
}

func (p *Parser) parsePrimitive() *Node {
	start := p.cursor
	c := p.input[start]

	node := arena.Alloc[Node](p.arena)
	*node = Node{}

	switch c {
	case 't':
		if !p.match("true") {
			panic("Expected true")
		}
		node.Type = True
		node.ValueStr = "true"
	case 'f':
		if !p.match("false") {
			panic("Expected false")
		}
		node.Type = False
		node.ValueStr = "false"
	case 'n':
		if !p.match("null") {
			panic("Expected null")
		}
		node.Type = Null
		node.ValueStr = "null"
	default:
		if c == '-' || (c >= '0' && c <= '9') {
			node.Type = Number
			numView := p.scanNumber()
			node.ValueStr = numView
		} else {
			panic(fmt.Sprintf("Unexpected character: %c", c))
		}
	}

	return node
}

func (p *Parser) scanStringBoundary() (int, bool) {
	data := p.input
	curr := p.cursor
	end := len(data)
	baseAddr := uintptr(unsafe.Pointer(unsafe.SliceData(data)))

	// Track if we saw any escape characters so far
	seenEscape := false

	for {
		quoteIdx := -1
		limit := end - 8
		scanCursor := curr

		// ---------------------------------------------------------------------
		// SIMD Loop: Check both " (0x22) and \ (0x5C) simultaneously
		// ---------------------------------------------------------------------
		for scanCursor <= limit {
			ptr := unsafe.Pointer(baseAddr + uintptr(scanCursor))
			val := *(*uint64)(ptr)

			// 1. Detect Quote (")
			xorQuote := val ^ 0x2222222222222222
			maskQuote := (xorQuote - lsb) & (^xorQuote) & msb

			// 2. Detect Backslash (\) - The cost is just a few CPU cycles
			xorEscape := val ^ 0x5C5C5C5C5C5C5C5C
			maskEscape := (xorEscape - lsb) & (^xorEscape) & msb

			// Case A: Quote Found
			if maskQuote != 0 {
				zeros := bits.TrailingZeros64(maskQuote)
				quoteIdx = scanCursor + (zeros >> 3)
				
				// Critical: Did we see an escape inside THIS chunk, BEFORE the quote?
				// Create a mask for bits before the quote position
				// (Go implies Little Endian logic here)
				bitPos := zeros // 0, 8, 16...
				
				// Create a mask that covers bytes strictly BEFORE the quote byte
				// e.g. if quote is at 3rd byte, mask allows 1st and 2nd byte
				// If bitPos is 0 (1st byte), mask is 0.
				var checkMask uint64
				if bitPos > 0 {
					checkMask = (1 << bitPos) - 1
				}
				
				if (maskEscape & checkMask) != 0 {
					seenEscape = true
				}
				
				break // Stop SIMD loop
			}

			// Case B: No Quote, but check Escape
			if maskEscape != 0 {
				seenEscape = true
			}

			scanCursor += 8
		}

		// Fallback: Scan remaining bytes (if SIMD loop finished without finding quote)
		if quoteIdx == -1 {
			for i := scanCursor; i < end; i++ {
				b := data[i]
				if b == '"' {
					quoteIdx = i
					break
				}
				if b == '\\' {
					seenEscape = true
				}
			}
		}

		if quoteIdx == -1 {
			panic("String not closed") // or return error
		}

		// ---------------------------------------------------------------------
		// Escape Verification (Odd/Even Rule) - Same as before
		// ---------------------------------------------------------------------
		bsCount := 0
		// Backward scan only for the quote validity
		for i := quoteIdx - 1; i >= p.cursor; i-- {
			if data[i] == '\\' {
				bsCount++
			} else {
				break
			}
		}

		if bsCount%2 == 0 {
			// Real closing quote found!
			return quoteIdx - p.cursor, seenEscape
		}

		// It was an escaped quote (\"). Treat it as a regular character.
		seenEscape = true // Since it's escaped, we definitely have an escape char
		curr = quoteIdx + 1 // Continue searching from next char
	}
}

func (p *Parser) scanNumber() string {
	start := p.cursor
	for p.cursor < len(p.input) {
		c := p.input[p.cursor]
		if isNumChar(c) {
			p.cursor++
		} else {
			break
		}
	}

	len := p.cursor - start
	basePtr := unsafe.SliceData(p.input)
	strStartPtr := unsafe.Pointer(uintptr(unsafe.Pointer(basePtr)) + uintptr(start))
	view := unsafe.String((*byte)(strStartPtr), len)
	return view
}

func isNumChar(c byte) bool {
	return c == '-' || (c >= '0' && c <= '9')
}

func (p *Parser) match(target string) bool {
	if p.cursor+len(target) > len(p.input) {
		return false
	}
	for i := 0; i < len(target); i++ {
		if p.input[p.cursor+i] != target[i] {
			return false
		}
	}
	p.cursor += len(target)
	return true
}

func (p *Parser) ParseObject() *Node {
	p.cursor++
	p.skipWhitespace()

	if p.cursor < len(p.input) && p.input[p.cursor] == '}' {
		p.cursor++
		node := arena.Alloc[Node](p.arena)
		*node = Node{}
		node.Type = Object
		return node
	}

	objNode := arena.Alloc[Node](p.arena)
	*objNode = Node{} // Zero-initialize
	objNode.Type = Object

	var tail *Node
	for {
		p.skipWhitespace()
		if p.input[p.cursor] != '"' {
			panic("Expected string key")
		}

		keyView := p.parseStringValueOnly()
		p.skipWhitespace()

		if p.input[p.cursor] != ':' {
			panic("Expected ':' after key")
		}

		p.cursor++

		valNode := p.ParseAny()
		valNode.Key = keyView
		if objNode.Children == nil {
			objNode.Children = valNode
		} else {
			tail.Next = valNode
		}
		tail = valNode
		p.skipWhitespace()

		c := p.input[p.cursor]
		if c == '}' {
			p.cursor++
			break
		} else if c == ',' {
			p.cursor++
		} else {
			panic(fmt.Sprintf("Expected ',' or '}', got %c", c))
		}
	}

	return objNode
}

func (p *Parser) parseStringValueOnly() string {
	if p.input[p.cursor] != '"' {
		panic("Expected quote")
	}
	p.cursor++

	startIdx := p.cursor

	len := p.findClosingQuoteLength()
	if len == -1 {
		panic("Key string not closed")
	}

	basePtr := unsafe.SliceData(p.input)
	strStartPtr := unsafe.Pointer(uintptr(unsafe.Pointer(basePtr)) + uintptr(startIdx))
	view := unsafe.String((*byte)(strStartPtr), len)

	p.cursor += len + 1

	return view
}

func (p *Parser) ParseArray() *Node {
	p.cursor++
	p.skipWhitespace()

	if p.cursor < len(p.input) && p.input[p.cursor] == ']' {
		p.cursor++
		node := arena.Alloc[Node](p.arena)
		*node = Node{}
		node.Type = Array
		return node
	}

	arrNode := arena.Alloc[Node](p.arena)
	*arrNode = Node{} // Zero-initialize
	arrNode.Type = Array

	var tail *Node

	for {
		p.skipWhitespace()
		valNode := p.ParseAny()
		if arrNode.Children == nil {
			arrNode.Children = valNode
		} else {
			tail.Next = valNode
		}
		tail = valNode
		p.skipWhitespace()

		c := p.input[p.cursor]
		if c == ']' {
			p.cursor++
			break
		} else if c == ',' {
			p.cursor++
		} else {
			panic("Expected ',' or ']'")
		}
	}

	return arrNode
}

// func (p *Parser) ParseString() *Node {
// 	p.cursor++ // Skip opening quote

// 	// One pass to find length AND escape status
// 	strLen, hasEscape := p.scanStringBoundary()

// 	node := arena.Alloc[Node](p.arena)
// 	*node = Node{}
// 	node.Type = String

// 	if !hasEscape {
// 		// [FAST PATH] Zero-Copy
// 		// No escapes found, so raw bytes == string value.
// 		basePtr := unsafe.SliceData(p.input)
// 		strStartPtr := unsafe.Pointer(uintptr(unsafe.Pointer(basePtr)) + uintptr(p.cursor))
// 		node.ValueStr = unsafe.String((*byte)(strStartPtr), strLen)
// 	} else {
// 		// [SLOW PATH] Unescape required
// 		// We found backslashes, so we must allocate new memory to remove them.
// 		// ex: "Hello \"World\"" (15 bytes) -> "Hello "World"" (13 bytes)
		
// 		// For simplicity, using Go's standard conversion which causes copy.
// 		// In a real high-perf parser, you would write a custom `Unescape(src, dst)` 
// 		// that writes directly into the Arena memory.
// 		node.ValueStr = string(p.input[p.cursor : p.cursor+strLen]) 
// 	}

// 	p.cursor += strLen + 1
// 	return node
// }

func (p *Parser) ParseString() *Node {
	if p.cursor >= len(p.input) {
		panic("Unexpected end of input")
	}

	if p.input[p.cursor] != '"' {
		panic("Expected string start '\"'")
	}

	p.cursor++
	startIdx := p.cursor

	strLen := p.findClosingQuoteLength()
	if strLen == -1 {
		panic("String not closed")
	}

	basePtr := unsafe.SliceData(p.input)
	strStartPtr := unsafe.Pointer(uintptr(unsafe.Pointer(basePtr)) + uintptr(startIdx))
	view := unsafe.String((*byte)(strStartPtr), strLen)

	node := arena.Alloc[Node](p.arena)
	*node = Node{} // Zero-initialize
	node.Type = String
	node.ValueStr = view
	p.cursor += strLen + 1
	return node
}

func main() {
	jsonInput := []byte(`{
		"name": "SimdParser",
		"description": "It says \"Hello\" to the world",
		"path": "C:\\Windows\\System32",
		"config": {
			"version": "1.0",
			"turbo": "on"
		}
	}`)

	arena := arena.NewBestArena()
	p := NewParser(jsonInput, arena)

	root := p.ParseAny()

	fmt.Printf("Root Type: %d (Object)\n", root.Type)

	child1 := root.Children
	fmt.Printf("Field 1: Key=%s, Val=%s\n", child1.Key, child1.ValueStr)

	child2 := child1.Next
	fmt.Printf("Field 2: Key=%s, Val=%s\n", child2.Key, child2.ValueStr)

	child3 := child2.Next
	fmt.Printf("Field 3: Key=%s, Val=%s\n", child3.Key, child3.ValueStr)

	child4 := child3.Next
	fmt.Printf("Field 4: Key=%s, Type=%d\n", child4.Key, child4.Type)

	configChild1 := child4.Children
	fmt.Printf("  -> Config Field 1: Key=%s, Val=%s\n", configChild1.Key, configChild1.ValueStr)

	configChild2 := configChild1.Next
	fmt.Printf("  -> Config Field 2: Key=%s, Val=%s\n", configChild2.Key, configChild2.ValueStr)
}

func (p *Parser) findClosingQuoteLength() int {
	data := p.input
	curr := p.cursor
	end := len(data)

	limit := end - 8

	baseAddr := uintptr(unsafe.Pointer(unsafe.SliceData(data)))

	for curr <= limit {
		ptr := unsafe.Pointer(baseAddr + uintptr(curr))

		val := *(*uint64)(ptr) // Unaligned Load

		// SWAR Logic
		xor := val ^ 0x2222222222222222
		hasZero := (xor - 0x0101010101010101) & (^xor) & 0x8080808080808080

		if hasZero != 0 {
			zeros := bits.TrailingZeros64(hasZero)
			idx := zeros >> 3
			realIdx := (curr - p.cursor) + idx

			// Check for even number of backslashes
			bsCount := 0
			// realIdx is relative to p.cursor, so absolute index in data is p.cursor + realIdx
			checkIdx := p.cursor + realIdx - 1
			for checkIdx >= 0 && data[checkIdx] == '\\' {
				bsCount++
				checkIdx--
			}

			if bsCount%2 == 0 {
				return realIdx
			}
			// It was escaped, continue searching from next byte
			// We can't just skip 8 bytes because there might be another quote in this block
			// But since we are iterating by 8, and we found one, we need to carefully proceed.
			// Actually, the simplest way after finding an escaped quote in a SWAR block
			// is to fallback to linear scan for the rest of this block or just tricky bit masking.
			// Ideally we clear this bit and try again, but just breaking to linear scan from here is defined and safe.
			curr = p.cursor + realIdx + 1
			goto LinearScan
		}
		curr += 8
	}

LinearScan:
	for curr < end {
		if data[curr] == '"' {
			// Check for even number of backslashes
			bsCount := 0
			checkIdx := curr - 1
			for checkIdx >= 0 && data[checkIdx] == '\\' {
				bsCount++
				checkIdx--
			}
			if bsCount%2 == 0 {
				return curr - p.cursor
			}
		}
		curr++
	}

	return -1
}
