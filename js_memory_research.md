# JavaScript Memory Monitoring Research

## The Fundamental Problem

JavaScript memory measurement using `process.memoryUsage().heapUsed` is inherently unreliable for several reasons:

### 1. **Garbage Collection Timing**
- V8 garbage collector runs unpredictably
- `global.gc()` forces collection but doesn't guarantee complete cleanup
- Memory may be freed but not returned to OS immediately

### 2. **Memory Fragmentation**
- Heap memory can be fragmented
- Available memory ≠ usable contiguous memory
- Memory regions may be reserved but not used

### 3. **V8 Internal Optimizations**
- Hidden classes and inline caching
- Deoptimization can cause memory spikes
- Internal data structures not visible to measurement

### 4. **Shared Memory Issues**
- Node.js runtime overhead
- Shared objects between measurements
- Module cache pollution

## Current Issues in Our Benchmarks

Looking at the data:
- Yjs: -144.95 MB (impossible negative value)
- Loro: -145.87 MB (impossible negative value)  
- These suggest baseline measurement > final measurement

## Potential Solutions

### Option 1: Separate Process Isolation
Run each benchmark in completely isolated Node.js processes:

```bash
# Each competitor in separate process
node --expose-gc --max-old-space-size=512 automerge_bench.js > automerge_results.csv
sleep 5  # Let system clear memory
node --expose-gc --max-old-space-size=512 yjs_bench.js > yjs_results.csv
```

### Option 2: Statistical Sampling
Run multiple iterations and take median/mean:

```javascript
const results = [];
for (let i = 0; i < 5; i++) {
  // Fresh process each time or significant cleanup
  results.push(runBenchmark());
}
return median(results);
```

### Option 3: Size-based Estimation (Current Approach)
Estimate based on data structure characteristics:
- Text length × 2 bytes (UTF-16)
- Operation count × CRDT overhead factor
- More predictable but less accurate

### Option 4: External Memory Monitoring
Monitor process from outside using system tools:
```bash
# Monitor from shell
ps -o pid,vsz,rss,comm -p $NODE_PID
```

## Operations/Second Reliability

**Timing measurement** (`Date.now()` or `process.hrtime.bigint()`) is generally reliable because:

1. **System Clock**: OS provides consistent timing
2. **No GC Impact**: Timing not affected by memory management  
3. **Large Scale**: Effects average out over thousands of operations
4. **Monotonic**: Time only moves forward

However, potential issues:
- **JIT Warmup**: First iterations may be slower (V8 optimization)
- **System Load**: Other processes can affect timing
- **Thermal Throttling**: CPU frequency changes

## Recommendations

1. **Trust Operations/Sec**: Timing measurements are generally reliable
2. **Memory Monitoring**: Use process isolation for accurate memory measurement
3. **Multiple Runs**: Average results across multiple benchmark runs
4. **Validation**: Cross-check with external monitoring tools

## Implementation Strategy

For reliable memory measurement:
1. Run each competitor in separate Node.js process
2. Use system memory monitoring alongside JS measurements  
3. Take multiple samples and use statistical methods
4. Accept that memory measurement will never be perfectly accurate in JS