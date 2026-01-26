# mcompiler: VM & Compiler 구현 로드맵

## 현재 상태 (Phase 0)
- ✅ Lexer (Tokenizer)
- ✅ Parser (Pratt Parser)
- ✅ AST 정의
- ✅ REPL (토큰만 출력)
- ✅ Arena Allocator (메모리 최적화)
- ✅ SIMD JSON Parser (실험적)

**다음 단계**: Tree-Walking Interpreter → Bytecode VM → Compiler → 고급 최적화

---

## Phase 1: Tree-Walking Interpreter (기초 구현)

### 목표
AST를 직접 순회하며 실행하는 기본 인터프리터 구현

### 구현 항목
1. **Object System 설계** (`/object/object.go`)
   - Value 타입 정의 (Integer, Boolean, Null, Function, String, Array, Hash)
   - Object interface 설계
   - Environment (변수 바인딩) 구현

2. **Evaluator 구현** (`/evaluator/evaluator.go`)
   - Expression 평가
     - Integer/Boolean/String literals
     - Prefix expressions (!, -)
     - Infix expressions (+, -, *, /, ==, !=, <, >)
     - If-else expressions
   - Statement 실행
     - Let statements (변수 선언)
     - Return statements
   - Function definitions & calls
   - Built-in functions (len, first, last, rest, push, puts)

3. **REPL 업그레이드**
   - 토큰 출력 → AST 파싱 → 평가 결과 출력으로 변경

### 예상 소요
- 핵심 기능: 1-2주
- 테스트 & 디버깅: 1주

---

## Phase 2: Bytecode 설계 & VM 기초

### 목표
Stack-based VM을 위한 bytecode instruction set 설계 및 기본 VM 구현

### 구현 항목
1. **Instruction Set 설계** (`/code/code.go`)
   ```
   Stack-based instructions:
   - OpConstant       // 상수를 스택에 푸시
   - OpAdd, OpSub, OpMul, OpDiv
   - OpTrue, OpFalse, OpNull
   - OpEqual, OpNotEqual, OpGreaterThan
   - OpMinus, OpBang  // 단항 연산자
   - OpJump, OpJumpNotTruthy  // 제어 흐름
   - OpGetGlobal, OpSetGlobal  // 전역 변수
   - OpGetLocal, OpSetLocal    // 지역 변수
   - OpArray, OpHash  // 복합 타입
   - OpIndex          // 인덱싱
   - OpCall, OpReturn // 함수 호출
   - OpGetBuiltin     // built-in 함수
   - OpClosure        // 클로저
   - OpGetFree        // free 변수 (클로저용)
   - OpReturnValue    // 명시적 return
   - OpPop            // 스택 정리
   ```

2. **Bytecode Encoding/Decoding**
   - Instruction format: [Opcode][Operand...]
   - Operand width 정의 (1-byte, 2-byte)
   - Disassembler (디버깅용)

3. **VM 구현** (`/vm/vm.go`)
   - Stack machine (1024 slots)
   - Constant pool
   - Global store (65536 slots)
   - Instruction pointer (IP)
   - Stack pointer (SP)
   - Fetch-Decode-Execute 사이클

4. **초기 통합**
   - Evaluator와 성능 비교 벤치마크
   - REPL에서 VM 모드 지원

### 예상 소요
- Bytecode 설계: 1주
- VM 구현: 2-3주
- 테스트: 1-2주

---

## Phase 3: Compiler 구현

### 목표
AST → Bytecode 컴파일러 구현

### 구현 항목
1. **Compiler 기본 구조** (`/compiler/compiler.go`)
   - AST 순회 (Visitor 패턴)
   - Instruction 생성
   - Constant pool 관리
   - Symbol table (변수 이름 → 인덱스 매핑)

2. **Expression Compilation**
   - Literals (integer, boolean, string)
   - Infix expressions
   - Prefix expressions
   - If-else (conditional jumps)
   - Array/Hash literals
   - Index expressions

3. **Statement Compilation**
   - Let statements (symbol table 업데이트)
   - Return statements
   - Expression statements

4. **Function Compilation**
   - Function literals → CompiledFunction 객체
   - Parameter binding
   - Closure 지원 (free 변수 캡처)
   - Scope 관리 (global/local/free)

5. **최적화 Phase 3.5**
   - Constant folding (컴파일 타임 상수 계산)
   - Dead code elimination (unreachable code 제거)
   - Peephole optimization (연속된 명령어 패턴 최적화)

### 예상 소요
- 기본 컴파일러: 2-3주
- 함수 & 클로저: 2주
- 초기 최적화: 1주

---

## Phase 4: 저수준 최적화 (고급)

### 4.1 Register-based VM 전환 (선택사항)
**목표**: Stack-based에서 Register-based VM으로 전환하여 성능 향상

**장점**:
- Instruction 수 감소 (stack manipulation overhead 제거)
- Lua VM 스타일의 고성능 달성

**구현**:
- Register allocation algorithm
- SSA (Static Single Assignment) 형태의 IR 도입
- Instruction set 재설계 (3-address code)

**예상 소요**: 3-4주

---

### 4.2 NaN Boxing
**목표**: Value representation 최적화 (포인터 추적 감소, 캐시 효율 향상)

**기법**:
```
64-bit value encoding:
- Integers: XXXX XXXX XXXX XXXX (직접 인코딩)
- Pointers: 0000 PPPP PPPP PPPP (하위 48비트)
- Booleans: FFFF FFFF FFFF FFF0/FFF1
- Null:     FFFF FFFF FFFF FFF2
- NaN:      7FF8 0000 0000 0001+
```

**장점**:
- Heap allocation 감소
- Type checking이 비트 연산으로 단순화
- Cache-friendly (8바이트 고정)

**예상 소요**: 2-3주

---

### 4.3 Inline Caching
**목표**: 동적 디스패칭 비용 감소 (property access, method calls)

**구현**:
- Monomorphic inline cache (단일 타입 추정)
- Polymorphic inline cache (2-4개 타입)
- Megamorphic 처리 (dictionary fallback)

**예상 소요**: 2주

---

### 4.4 Direct Threading / Computed Goto
**목표**: Instruction dispatch overhead 최소화

**기법**:
```go
// 기존 switch-case dispatch
switch opcode {
case OpAdd: ...
case OpSub: ...
}

// Computed goto (Go에서는 function pointer table 사용)
var dispatchTable []func()
dispatchTable[OpAdd] = func() { /* add logic */ }
dispatchTable[ip]()  // 직접 점프
```

**장점**:
- Branch prediction 향상
- 5-25% 성능 향상

**예상 소요**: 1-2주

---

### 4.5 SIMD Optimizations
**목표**: 배열/문자열 연산을 벡터화

**적용 영역**:
- String comparison (이미 JSON parser에서 사용 중)
- Array operations (map, filter, reduce)
- Numeric array arithmetic

**구현**:
- Go의 `golang.org/x/sys/cpu` 활용
- SWAR (SIMD Within A Register) 확장
- AVX2/AVX512 intrinsics (cgo를 통한 C 코드)

**예상 소요**: 2-3주

---

### 4.6 JIT Compilation (Tier 1)
**목표**: Hot code를 native machine code로 컴파일

**아키텍처**:
```
Bytecode → Profiling → Hot path detection → JIT compile → Execute native
```

**구현 단계**:
1. **Profiling Infrastructure**
   - Instruction counter
   - Hot loop detection (backward jumps)
   - Function call counters

2. **Simple JIT (Template JIT)**
   - Pre-compiled assembly templates
   - x86-64 instruction encoding (amd64만 지원)
   - 간단한 expressions만 JIT (add, sub, mul, div)

3. **Register Allocation**
   - Linear scan allocation
   - x86-64 registers 매핑 (rax, rbx, rcx, rdx)

4. **Deoptimization**
   - Type guard 실패 시 bytecode로 fallback
   - OSR (On-Stack Replacement)

**예상 소요**: 6-8주

---

### 4.7 Advanced JIT (Tier 2 - 야심찬 목표)
**목표**: 최적화 컴파일러 수준의 JIT

**기법**:
- SSA-based IR (LLVM 참고)
- GVN (Global Value Numbering)
- Loop unrolling
- Inlining (함수 호출 제거)
- Escape analysis (stack allocation 결정)
- Trace-based JIT (LuaJIT 스타일)

**통합**:
- LLVM bindings 사용 (기존 최적화 인프라 활용)
- Cranelift (Rust JIT library, cgo 바인딩)

**예상 소요**: 3-4개월 (장기 프로젝트)

---

### 4.8 Garbage Collection 최적화
**목표**: 현재 Go GC 의존도 낮추고 custom allocator 확장

**기법**:
1. **Generational GC**
   - Young generation (arena allocator - 이미 구현됨)
   - Old generation (mark-sweep)
   - Write barrier

2. **Incremental GC**
   - Pause time 최소화
   - Tri-color marking

3. **Arena 확장**
   - Per-thread arenas (lock-free allocation)
   - Size class segregation (small/medium/large objects)

**예상 소요**: 4-6주

---

### 4.9 Profile-Guided Optimization (PGO)
**목표**: 실행 프로파일 기반 최적화

**구현**:
1. **Profiling Data 수집**
   - Branch frequencies
   - Hot functions
   - Type profiles

2. **Optimization**
   - Basic block reordering (hot path를 연속 메모리에 배치)
   - Speculative inlining
   - Devirtualization

**예상 소요**: 3-4주

---

## Phase 5: 생산성 & 도구

### 구현 항목
1. **Debugger**
   - Breakpoints
   - Step execution
   - Stack trace
   - Variable inspection

2. **Profiler**
   - CPU profiling (pprof 통합)
   - Memory profiling
   - Flame graphs

3. **Benchmarking Suite**
   - Standard benchmark programs
   - Comparative analysis (vs Python, Lua, JavaScript)

4. **Documentation**
   - VM architecture guide
   - Optimization techniques guide
   - Contributor guide

**예상 소요**: 4-6주

---

## 성능 목표 & 벤치마크

### 비교 대상
- Tree-walking interpreter (baseline: 1x)
- Bytecode VM: **5-10x faster**
- VM + NaN boxing: **8-15x faster**
- VM + Inline caching: **10-20x faster**
- Simple JIT: **20-40x faster**
- Advanced JIT: **40-100x faster** (최적 케이스)

### 벤치마크 프로그램
1. Fibonacci (재귀)
2. Array manipulation
3. Object property access
4. Function calls (overhead 측정)
5. String concatenation

---

## 우선순위 추천 순서

### 필수 구현 (순차적으로 진행)
1. ✅ Phase 1: Tree-Walking Interpreter
2. ✅ Phase 2: Bytecode VM
3. ✅ Phase 3: Compiler

### 고급 최적화 (병렬 가능)
**Quick Wins (높은 효과, 낮은 난이도)**:
4. 🔥 Phase 4.4: Direct Threading (5-25% 향상, 2주)
5. 🔥 Phase 4.2: NaN Boxing (2-3배 향상, 3주)
6. 🔥 Phase 3.5: Compiler Optimizations (10-30% 향상, 1주)

**중간 난이도**:
7. Phase 4.3: Inline Caching (2-3배 향상, 2주)
8. Phase 4.5: SIMD Extensions (특정 workload 5-10배 향상, 3주)

**고난이도 장기 프로젝트**:
9. Phase 4.1: Register-based VM (1.5-2배 향상, 4주)
10. Phase 4.6: Simple JIT (5-10배 향상, 8주)
11. Phase 4.8: Custom GC (안정성 향상, 6주)
12. Phase 4.7: Advanced JIT (10-50배 향상, 4개월)

---

## 참고 자료

### 책
- "Writing An Interpreter In Go" (Thorsten Ball) - Phase 1
- "Writing A Compiler In Go" (Thorsten Ball) - Phase 2-3
- "Crafting Interpreters" (Robert Nystrom) - VM 설계
- "Advanced Compiler Design and Implementation" (Steven Muchnick)

### 오픈소스 구현
- **Lua VM**: Register-based VM의 교과서
- **LuaJIT**: Trace-based JIT의 정점
- **CPython**: Bytecode VM 참고
- **V8 (JavaScript)**: 현대 JIT 구현
- **Wren**: 작고 빠른 VM (NaN boxing 사용)
- **Gravity**: Swift-like 언어, 깔끔한 VM 구조

### 논문
- "Optimizing Dynamically-Typed Object-Oriented Languages With Polymorphic Inline Caches" (1991)
- "An Inline Caching Strategy for Dynamic Typing"
- "Fast Native Functions for JavaScript" (V8)
- "Trace-based Just-in-Time Type Specialization for Dynamic Languages" (LuaJIT)

---

## 예상 전체 일정

| Phase | 기간 | 누적 |
|-------|------|------|
| Phase 1: Interpreter | 3-4주 | 1개월 |
| Phase 2: Bytecode VM | 4-6주 | 2.5개월 |
| Phase 3: Compiler | 4-5주 | 3.5개월 |
| Phase 4 Quick Wins | 6-8주 | 5.5개월 |
| Phase 4 중급 | 5-7주 | 7개월 |
| Phase 4 고급 (JIT) | 3-6개월 | 12-14개월 |

**최소 실용 버전 (Bytecode VM + 기본 최적화)**: 5.5개월
**고성능 버전 (Simple JIT 포함)**: 10개월
**최첨단 버전 (Advanced JIT 포함)**: 14개월+

---

## 다음 단계

Phase 1부터 시작하려면:

```bash
mkdir -p object evaluator
touch object/object.go evaluator/evaluator.go evaluator/evaluator_test.go
```

첫 번째 구현 목표: Integer와 Boolean을 평가하는 기본 Evaluator 작성

**시작 코드 예시**: Integer literal 평가
```go
// evaluator/evaluator.go
func Eval(node ast.Node) object.Object {
    switch node := node.(type) {
    case *ast.IntegerLiteral:
        return &object.Integer{Value: node.Value}
    }
    return nil
}
```

즐거운 컴파일러 개발 되세요! 🚀
