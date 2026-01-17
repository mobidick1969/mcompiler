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
```

## 🛠️ Git & GitHub

**Q: What does "error: GH007: Your push would publish a private email address" mean?**
**A:** GitHub's "Block command line pushes that expose my email" setting is on, but your local git config (`user.email`) is using a private email (or one not verified).
**Fix:** Check email with `git config user.email` and change it to your public GitHub email or the `users.noreply.github.com` address.

**Q: How do I fix the author email of the last commit?**
**A:** `git commit --amend --reset-author` (after updating `git config user.email`).
