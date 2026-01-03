# SIMD JSON Parser for Go

A high-performance, validating JSON parser for Go, optimized using SIMD (SWAR) techniques and Arena allocation.

## Features

- **🚀 High Performance**: Up to **~6x faster** than the standard `encoding/json` library for struct unmarshaling.
- **⚡ SIMD / SWAR Optimization**: Uses "SIMD Within A Register" techniques to process 8 bytes at a time for finding quotes and structural characters, dramatically reducing CPU cycles.
- **💾 Zero-Allocation**: Built on top of a custom **Arena Allocator**, achieving **0 allocations** per parse operation.
- **🚫 Zero-Copy Strings**: Creates string views directly from the input buffer for unescaped strings, avoiding expensive memory copies.
- **🛡️ Validating**: Correctly handles complexities like escaped quotes (`\"`) and nested structures.

## Benchmarks

Benchmarks run on Apple M1 Max.

### Small Payload (~1KB)

| Benchmark | Time/Op | Bytes/Op | Allocs/Op | Speedup |
|-----------|---------|----------|-----------|---------|
| Standard (Map) | 5900 ns | 3120 B | 73 | 1x |
| Standard (Struct) | 4700 ns | 472 B | 11 | 1.25x |
| **FastParser (SIMD)** | **693 ns** | **0 B** | **0** | **~8.5x** |

### Large Payload (~1MB)

| Benchmark | Time/Op | Bytes/Op | Allocs/Op | Speedup |
|-----------|---------|----------|-----------|---------|
| Standard (Map) | 10.94 ms | 5.09 MB | 122k | 1x |
| Standard (Struct) | 8.35 ms | 0.28 MB | 2022 | 1.3x |
| **FastParser (SIMD)** | **1.45 ms** | **0.005 MB** | **0*** | **~5.8x** |

*\*Allocations for FastParser are for arena growth only ( amortized to 0 per op in pre-sized scenarios).*

### String Scanning Micro-Benchmark

| Function | Time/Op | Approach |
|----------|---------|----------|
| `findClosingQuoteLength` | 25.46 ns | Two-pass (Scan + Backward Check) |
| `scanStringBoundary` | **16.52 ns** | **Single-pass SIMD (Forward)** |

## Usage

```go
package main

import (
	"fmt"
	"mcompiler/arena"
	"mcompiler/simd"
)

func main() {
	jsonInput := []byte(`{
		"name": "SuperFast",
		"config": { "turbo": true }
	}`)

	// 1. Create an Arena (Reusable)
	a := arena.NewBestArena()
	
	// 2. Create Parser
	p := simd.NewParser(jsonInput, a)
	
	// 3. Parse
	root := p.ParseAny()
	
	fmt.Printf("Parsed: %s\n", root.Children.ValueStr)
	
	// 4. Reset Arena for next request (Instant deallocation)
	a.Reset()
}
```

## How it works

### SIMD (SWAR)
The parser reads data in 8-byte chunks (uint64) and uses bitwise XOR and subtraction logic to identify calling characters (quotes, backslashes) in parallel, rather than checking byte-by-byte.

### Arena Allocation
Instead of letting the Go GC manage millions of small `Node` objects, we allocate them linearly in a pre-allocated byte slice. Resetting the parser is as simple as setting an integer offset to 0.
