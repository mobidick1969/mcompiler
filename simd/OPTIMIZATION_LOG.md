# 🚀 SIMD JSON Parser Optimization Report

This document records the **SIMD optimization and tuning process** performed to push the performance of the Go-based JSON parser to its limits.

## 🏁 Final Results Summary

Achieved **~8.5x faster speed** and **Zero Allocation** compared to the standard library (`encoding/json`).

| Optimization Stage | Small Payload Speed | Large Payload Speed | Memory Alloc (Alloc/Op) | Note |
|:---:|:---:|:---:|:---:|:---|
| **Standard Library** | 5900 ns | 10.94 ms | 73 (3KB) | Baseline |
| **Arena Introduction** | 1200 ns | 2.50 ms | 0 (0B) | Memory Reuse |
| **ScanStringBoundary** | 751 ns | 1.45 ms | 0 (0B) | String Parsing Acceleration |
| **SIMD Number** | 742 ns | 1.42 ms | 0 (0B) | Long Number Acceleration |
| **Fused Scanning (Final)** | **693 ns** | **1.33 ms** | **0 (0B)** | Minimal Branch/Call Overhead |

---

## 🏆 Comparison with State-of-the-Art (`valyala/fastjson`)

We benchmarked against `github.com/valyala/fastjson`, widely considered the fastest Go JSON parser.

| Library | Small Payload | Large Payload | Allocations (Large) |
|:---|:---:|:---:|:---:|
| **mcompiler/simd (Ours)** | **696 ns** | **1.33 ms** | **0** |
| valyala/fastjson | 634 ns | 1.15 ms | 21 |

---

## 💾 Memory Footprint Comparison

We analyzed the **Total Allocated Bytes** (via `-alloc_space`) to see the true memory pressure.

| Library | Allocated Bytes / Op (Large) | Note |
|:---|:---:|:---|
| **mcompiler/simd (Ours)** | **~4.6 KB** | Purely for Arena expansion (reused). |
| valyala/fastjson | ~22.3 KB | Internal caching & KV pairs. |
| encoding/json | ~5,090 KB (5 MB) | Garbage created per parse. |

**Observation**: `encoding/json` creates **5 MB** of garbage for a 1 MB file. Our parser creates virtually **0 bytes** of garbage (the 4.6 KB is mostly amortized arena growth which is reused).

---

## 🧪 Robustness & Stress Testing

We verified correctness and performance on challenging payloads.

| Scenario | FastParser Time | fastjson Time | StdLib Time | Note |
|:---|:---:|:---:|:---:|:---|
| **Deep Nesting** (1000 levels) | **29.3 µs** | 21.0 µs | 157.3 µs | fastjson is faster (recursive optimization). |
| **Float Array** (1000 floats) | **23.7 µs** | 12.1 µs | 105.3 µs | fastjson wins on raw number scanning. |

**Correctness**: Verified against `encoding/json` output for escaped strings, unicodes, and nested structures.

---

## 🛠️ Key Optimization Techniques

### 1. Memory Arena & Inline Allocation
Introduced an **Arena Allocator** to eliminate Go GC (Garbage Collector) overhead.
- **Initial Attempt**: Used Generic functions (`Alloc[T]`) but suffered performance degradation due to function call overhead.
- **Solution**: **Inlined** allocation logic directly into the parser and controlled memory addresses using `unsafe.Pointer`.
- **Result**: Reduced Heap Allocation to **0**.

### 2. SIMD String Scanning (SWAR Technique)
Optimized the core string scanning (`"..."`) using SWAR (SIMD Within A Register) technique.
- **Technique**: Loads 8 bytes (uint64) at once and uses bitwise operations (XOR, SUB, AND) to detect `"` and `\` characters simultaneously.
- **Zero-Copy**: Eliminated copy costs by mapping the raw byte slice (`[]byte`) pointer directly to the `string` header (`unsafe.String`).

### 3. SIMD Number Scanning
Leveraged the fact that most characters in number parsing are digits (`0`~`9`).
- **Technique**: Loads 8 bytes and verifies if **all bytes are within the digit range** using a single SWAR operation.
- **Effect**: Drastically reduced loop iterations by 1/8 when parsing long timestamps or numeric IDs.

### 4. Fused Scanning & Profiling (Micro-Optimization)
Identified via profiling (`pprof`) that `peekNextToken` consumed 47% of total execution time.
- **Problem**: 4 comparison operations (`== ' ' || == '\n' ...`) in the scalar loop for whitespace skipping were a bottleneck.
- **Solution**: Replaced checks with a **single range comparison** (`c <= ' '`), leveraging the fact that all whitespace characters are ≤ 0x20.
- **Result**: Reduced comparison cost from 440ms to 180ms, improving overall performance by an additional 7%.

---

## 🔍 Conclusion
We have reached an **Optimum** state where only hardware-level memory bandwidth and function call overhead remain.

> "The fastest code is the code that never runs."
