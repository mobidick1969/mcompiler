# mcompiler - High Performance Monkey Compiler

**mcompiler** is a modern, high-performance implementation of the Monkey programming language compiler, written in Go. It goes beyond the standard implementation by applying advanced optimization techniques such as Arena Allocation, SIMD processing, and Zero-Copy string handling.

## 🌟 Key Features

- **Zero-Allocation Philosophy**: Extensive use of Arena Allocators to minimize Garbage Collection (GC) overhead.
- **Modern Go**: Utilizes Go Generics and `unsafe` optimizations for maximum performance.
- **SIMD Acceleration**: Experimental SIMD-based JSON parser demonstrating vectorization techniques in Go.

## 📂 Modules

### `parser/`
The core of the compiler. Implements a **Pratt Parser** (Recursive Descent) to handle expressions with varying precedence.
- **Goal**: Fast, robust parsing with meaningful error reporting.

### `arena/`
A custom **Arena Allocator** implementation.
- **Why?**: Allocating millions of AST nodes individually causes massive GC pressure.
- **How?**: We pre-allocate large memory blocks and hand out pointers linearly. Deallocation is instant (resetting an offset).

### `simd/`
An experimental playground for **SIMD (Single Instruction, Multiple Data)** and **SWAR (SIMD Within A Register)** optimizations.
- **Highlights**: A JSON parser that is **~6x faster** than Go's standard library.
- **Tech**: see [simd/README.md](simd/README.md) for benchmarks and details.

## 🚀 Future Roadmap

We are actively working on:
- **Incremental Parsing**: Re-parsing only changed code regions.
- **Error Recovery**: Panic mode recovery to detect multiple errors at once.
- **Parallel Compilation**: Building large projects in parallel.

See [TODO.md](TODO.md) for the full roadmap.

## 🛠️ Usage

```bash
# Clone the repository
git clone https://github.com/yourusername/mcompiler.git

# Run tests
go test ./...

# Run benchmarks
cd simd
go test -bench . -benchmem
```
