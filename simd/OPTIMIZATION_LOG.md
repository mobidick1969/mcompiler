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
