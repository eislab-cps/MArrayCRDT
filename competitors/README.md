# CRDT Competitors Benchmark Suite

This directory contains performance benchmarks for competing CRDT libraries using the same Kleppmann editing trace used for MArrayCRDT testing.

## Structure

```
competitors/
├── automerge/         # Automerge CRDT benchmark
├── yjs/              # Yjs CRDT benchmark  
├── loro/             # Loro CRDT benchmark
└── run_all.js        # Master runner for all benchmarks
```

## Quick Start

Run all competitor benchmarks:

```bash
node run_all.js
```

This will:
1. 🔧 Install dependencies for all competitors
2. 🚀 Run benchmarks using the Kleppmann editing trace
3. 📊 Generate consolidated results in `../data/competitors_comparison.csv`
4. 📈 Display performance summary

## Individual Benchmarks

Run individual competitor benchmarks:

```bash
# Automerge
cd automerge && npm install && npm run benchmark

# Yjs  
cd yjs && npm install && npm run benchmark

# Loro
cd loro && npm install && npm run benchmark
```

## Benchmark Details

All benchmarks use:
- **Same editing trace**: Real 259k operations from writing a LaTeX paper
- **Multiple scales**: 1k, 5k, 10k, 20k, 30k, 40k, 50k operations
- **Real memory measurements**: Using Node.js `process.memoryUsage()`
- **Identical test conditions**: Same operations, same progression

## Metrics Collected

For each scale and system:
- **Operations per second** (throughput)
- **Memory usage** (MB)
- **Total execution time** (ms)
- **Final document length** (characters)

## Output Format

Results are saved as CSV with columns:
```
system,operations,time_ms,ops_per_sec,memory_mb,final_length
```

## Integration

Results are automatically integrated into the web visualization at `http://localhost:3000` after running the main MArrayCRDT benchmark.

## Libraries Tested

- **Automerge** v2.3.2 - JSON-like documents with automatic merging
- **Yjs** v13.6.18 - Modular building blocks for real-time collaboration  
- **Loro** v1.0.7 - High-performance CRDT with advanced features
- **MArrayCRDT** - Our implementation (benchmark separately)