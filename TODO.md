# mcompiler TODO List

This file documents the future direction of `mcompiler` and modern optimization techniques we wish to implement.

## 🚀 Parser Optimization and Enhancement Plan

- [ ] **1. Error Recovery (Panic Mode)**
    - Instead of stopping immediately on syntax errors, skip to the next semicolon (`;`) or brace (`}`) to detect multiple errors at once.
- [ ] **2. Incremental Parsing**
    - Optimize by updating only parts of the AST based on the changed range (Interval) or Hash, rather than re-parsing the entire code upon modification.
- [ ] **3. Memory and Resource Optimization**
    - **Arena Allocation**: Reduce GC load by allocating AST nodes in bulk from large memory blocks instead of individually.
    - **String Interning**: Optimize memory usage by sharing memory addresses for identical identifier names.
- [ ] **4. Advanced Parsing Techniques**
    - **Pratt Parsing**: Table-based parsing to efficiently handle operator precedence.
    - **Parallel Parsing**: Improve build speed for large projects by parsing multiple files in parallel.

## 🧠 Memory Management Strategy for Large Scale Projects (Future Consideration)

- [ ] **1. On-demand & Lazy Parsing**
    - Strategy to parse only when needed and immediately release AST nodes from memory when finished.
- [ ] **2. Compact AST Structure Design**
    - Minimize per-node memory usage by using integer indices instead of pointers or applying the Flyweight pattern.
- [ ] **3. Symbol Summary and Serialization**
    - Store and load only the information necessary for type checking in binary form (`.a`) instead of the entire AST.
- [ ] **4. Incremental Compilation**
    - Prevent redundant work in gigabyte-scale projects by recompiling only modified files and caching the results.

---
*Please feel free to suggest any additional ideas!*
