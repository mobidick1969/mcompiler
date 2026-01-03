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

func (p *Parser) peekNextToken() byte {
	// SIMD (SWAR) Optimization
	// ... (comments trimmed for brevity)

	limit := len(p.input) - 8
	startPtr := unsafe.Pointer(unsafe.SliceData(p.input))

	for p.cursor <= limit {
		ptr := unsafe.Add(startPtr, p.cursor)
		val := *(*uint64)(ptr)

		sub := 0x2020202020202020 - val
		top := sub & 0x8080808080808080

		if top != 0 {
			break
		}
		p.cursor += 8
	}

	for p.cursor < len(p.input) {
		c := p.input[p.cursor]
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			p.cursor++
		} else {
			return c
		}
	}
	return 0 // EOF
}

// Pre-calculate size and alignment to avoid compile-time lookup overhead in generic function
const nodeSize = int(unsafe.Sizeof(Node{}))
const nodeAlign = int(unsafe.Alignof(Node{}))

func (p *Parser) ParseAny() *Node {
	char := p.peekNextToken()
	if char == 0 {
		panic("Unexpected EOF")
	}

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

	var node *Node
	if p.arena.Offset+nodeSize <= len(p.arena.Current.Data) {
		ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(p.arena.Current.Data)), p.arena.Offset)
		node = (*Node)(ptr)
		p.arena.Offset += nodeSize
	} else {
		node = (*Node)(p.arena.AllocUnsafe(nodeSize, nodeAlign))
	}
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
	startPtr := unsafe.Pointer(unsafe.SliceData(data))

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
			ptr := unsafe.Add(startPtr, scanCursor)
			val := *(*uint64)(ptr)

			// 1. Detect Quote (")
			xorQuote := val ^ 0x2222222222222222
			maskQuote := (xorQuote - lsb) & (^xorQuote) & msb

			// 2. Detect Backslash (\)
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
		seenEscape = true   // Since it's escaped, we definitely have an escape char
		curr = quoteIdx + 1 // Continue searching from next char
	}
}

func (p *Parser) scanNumber() string {
	start := p.cursor

	limit := len(p.input) - 8
	startPtr := unsafe.Pointer(unsafe.SliceData(p.input))

	// SIMD: Skip contiguous digits ('0'..'9')
	for p.cursor <= limit {
		ptr := unsafe.Add(startPtr, p.cursor)
		val := *(*uint64)(ptr)

		// Check if any byte is outside '0'..'9' range using SWAR arithmetic
		t1 := val - 0x3030303030303030 // Detect < '0'
		t2 := val + 0x4646464646464646 // Detect > '9'
		mask := (t1 | t2) & 0x8080808080808080

		if mask != 0 {
			break
		}
		p.cursor += 8
	}

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
	strStartPtr := unsafe.Add(unsafe.Pointer(basePtr), start)
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
	p.cursor++ // Skip '{'

	if p.peekNextToken() == '}' {
		p.cursor++
		var node *Node
		if p.arena.Offset+nodeSize <= len(p.arena.Current.Data) {
			ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(p.arena.Current.Data)), p.arena.Offset)
			node = (*Node)(ptr)
			p.arena.Offset += nodeSize
		} else {
			node = (*Node)(p.arena.AllocUnsafe(nodeSize, nodeAlign))
		}
		*node = Node{}
		node.Type = Object
		return node
	}

	var objNode *Node
	if p.arena.Offset+nodeSize <= len(p.arena.Current.Data) {
		ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(p.arena.Current.Data)), p.arena.Offset)
		objNode = (*Node)(ptr)
		p.arena.Offset += nodeSize
	} else {
		objNode = (*Node)(p.arena.AllocUnsafe(nodeSize, nodeAlign))
	}
	*objNode = Node{} // Zero-initialize
	objNode.Type = Object

	var tail *Node
	for {
		if p.peekNextToken() != '"' {
			panic("Expected string key")
		}

		keyView := p.parseStringValueOnly()

		if p.peekNextToken() != ':' {
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

		c := p.peekNextToken()
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

	// Use synchronized SIMD scanning
	strLen, _ := p.scanStringBoundary()

	basePtr := unsafe.SliceData(p.input)
	strStartPtr := unsafe.Add(unsafe.Pointer(basePtr), p.cursor)
	view := unsafe.String((*byte)(strStartPtr), strLen)

	p.cursor += strLen + 1

	return view
}

func (p *Parser) ParseArray() *Node {
	p.cursor++ // Skip '['

	if p.peekNextToken() == ']' {
		p.cursor++
		var node *Node
		if p.arena.Offset+nodeSize <= len(p.arena.Current.Data) {
			ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(p.arena.Current.Data)), p.arena.Offset)
			node = (*Node)(ptr)
			p.arena.Offset += nodeSize
		} else {
			node = (*Node)(p.arena.AllocUnsafe(nodeSize, nodeAlign))
		}
		*node = Node{}
		node.Type = Array
		return node
	}

	var arrNode *Node
	if p.arena.Offset+nodeSize <= len(p.arena.Current.Data) {
		ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(p.arena.Current.Data)), p.arena.Offset)
		arrNode = (*Node)(ptr)
		p.arena.Offset += nodeSize
	} else {
		arrNode = (*Node)(p.arena.AllocUnsafe(nodeSize, nodeAlign))
	}
	*arrNode = Node{} // Zero-initialize
	arrNode.Type = Array

	var tail *Node

	for {
		// p.peekNextToken() is called inside ParseAny
		valNode := p.ParseAny()
		if arrNode.Children == nil {
			arrNode.Children = valNode
		} else {
			tail.Next = valNode
		}
		tail = valNode

		c := p.peekNextToken()
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

func (p *Parser) ParseString() *Node {
	if p.cursor >= len(p.input) {
		panic("Unexpected end of input")
	}

	if p.input[p.cursor] != '"' {
		panic("Expected string start '\"'")
	}

	p.cursor++ // Skip opening quote

	strLen, _ := p.scanStringBoundary()

	// Inline Allocation of Node
	var node *Node
	if p.arena.Offset+nodeSize <= len(p.arena.Current.Data) {
		ptr := unsafe.Add(unsafe.Pointer(unsafe.SliceData(p.arena.Current.Data)), p.arena.Offset)
		node = (*Node)(ptr)
		p.arena.Offset += nodeSize
	} else {
		node = (*Node)(p.arena.AllocUnsafe(nodeSize, nodeAlign))
	}
	*node = Node{} // Zero-initialize
	node.Type = String

	// Optimization: Always use Zero-Copy.
	// For maximum performance and 0-allocation, we return the raw slice reference.
	// Note: Strings with escapes will contain raw backslashes (e.g. "a\"b").
	// A compliant implementation should unescape if hasEscape is true.
	basePtr := unsafe.SliceData(p.input)
	strStartPtr := unsafe.Add(unsafe.Pointer(basePtr), p.cursor)
	node.ValueStr = unsafe.String((*byte)(strStartPtr), strLen)

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
