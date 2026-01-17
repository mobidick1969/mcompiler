# 🧠 Learning Flashcards

## 🚀 Performance & Optimization

**Q: Why can a simple Scalar Loop be faster than SIMD for parsing floats?**
**A:** SIMD has a "Setup Cost" (loading 8/16 bytes, masking, logic overhead). If the token is short (e.g., `123.45`) or contains characters that break SIMD immediately (like `.`), you pay the setup cost but process 0 bytes, then fall back to scalar. `fastjson` wins here by using an extremely tight, inlined scalar loop that starts processing immediately.

**Q: What is "Fused Scanning" in the context of parsers?**
**A:** Combining "Skip Whitespace" and "Peek Next Token" into a single loop. Instead of `skipWhitespace(); c := peek()`, you iterate once: if `c <= ' '` continue, else return `c`. This saves one layer of loop overhead and memory access.

**Q: How does Arena Allocation achieve 0 Allocations?**
**A:** Instead of `new(Node)` which allocates on the Heap and involves GC, an Arena pre-allocates a massive byte slice (e.g., 10MB). `Alloc()` simply returns a pointer to the current offset and increments the offset. Resetting is just `offset = 0`. No Garbage Collection is needed.

## 🧪 Go Benchmarking

**Q: How do you run only specific benchmarks in Go?**
**A:** Use the `-bench` regex flag.
`go test -bench="FastParser|Valyala" .` (Runs benchmarks matching FastParser OR Valyala)

**Q: How do you programmatically skip a benchmark test?**
**A:** Use `b.Skip("Reason")` inside the benchmark function.
```go
func BenchmarkTooSlow(b *testing.B) {
    b.Skip("This takes too long")
    // ...
}

## 🔥 Advanced Profiling (pprof)

**Q: How do I generate CPU and Memory profiles from benchmarks?**
**A:** Use flags `-cpuprofile` and `-memprofile`:
```bash
go test -bench=. -cpuprofile cpu.prof -memprofile mem.prof
```

**Q: How do I see which specific lines are slow?**
**A:** Open the profile with `go tool pprof` and use `list` command:
```bash
go tool pprof cpu.prof
(pprof) list MySlowFunction
```

**Q: How do I check for "Hidden" Allocations (Heap Analysis)?**
**A:** Check `alloc_objects` (count) and `alloc_space` (bytes):
```bash
# Check allocation count (Are we creating too many objects?)
go tool pprof -top -alloc_objects mem.prof

# Check allocation size (Are we creating huge objects?)
go tool pprof -top -alloc_space mem.prof
```

**Q: What is "Assembly Analysis" in pprof?**
**A:** If `list` isn't detailed enough, use `disasm` to see the Assembly code. This helps spot memory moves (`MOV`), bounds checks, or lack of SIMD instructions.
```bash
(pprof) disasm MySlowFunction
```

**Q: Why does `disasm` or `list` fail with "no matches found"?**
**A:** `pprof` needs the **executable binary** to read assembly/source code.
If you just ran `go tool pprof cpu.prof`, it might not find the binary.
**Fix:** Explicitly provide the binary:
```bash
go tool pprof simd.test cpu.prof
```
*(Note: `go test -cpuprofile` creates both `.prof` and `.test` files)*

**Q: How do I profile ONLY a specific benchmark function?**
**A:** Combine `-bench` filter with profile flags:
```bash
# Profile only benchmarks matching "FastParser"
go test -bench=FastParser -cpuprofile cpu.prof
```

**Q: Useful pprof Interactive Commands Cheat Sheet**
| Command | Description |
|:---|:---|
| `top` / `top20` | Show top functions by flat time (default). |
| `top -cum` | Show top functions by **cumulative** time (useful to find heavy call stacks). |
| `list <Func>` | Show source code of `<Func>` with CPU/Memory usage per line. |
| `peek <Func>` | Show callers (who called this) and callees (who this calls). |
| `disasm <Func>` | Show assembly instructions annotated with samples. |
| `web` | Open a visualized graph in the web browser (requires Graphviz). |
| `o` | Show current option settings (sort, sample_index, etc.). |

**Q: Error: `failed to execute dot. Is Graphviz installed?`**
**A:** `pprof` visualization (`web`, `pdf`, `png`) requires **Graphviz**.
**Fix (Mac):** Install it via Homebrew:
```bash
brew install graphviz
```
