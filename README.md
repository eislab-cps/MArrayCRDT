# MArrayCRDT - Movable Array Conflict-free Replicated Data Type

A CRDT implementation in Go that supports array operations including move, sort, reverse, and swap, with comprehensive benchmarking against leading CRDT libraries.

## Repository Structure

```
├── crdt/                  # Core MArrayCRDT implementation (Go)
├── benchmarks/            # MArrayCRDT performance benchmarks (Go)
├── competitors/           # Competitor CRDT benchmarks (JavaScript)
│   ├── automerge/         # Automerge CRDT benchmarks
│   ├── yjs/               # Yjs CRDT benchmarks  
│   ├── loro/              # Loro CRDT benchmarks (Text + MovableList)
│   └── baseline/          # JavaScript Array baseline
├── web/                   # Interactive web visualization
├── data/                  # Benchmark data and editing traces
│   ├── paper.json         # Real editing trace (259k operations)
│   └── benchmark_runs/    # Timestamped benchmark results
└── run_all_benchmarks.sh  # Complete benchmark suite runner
```

## Quick Start

### Complete Benchmark Suite
```bash
# Run all benchmarks with proper isolation between libraries
./run_all_benchmarks.sh
```

This comprehensive suite:
- Runs MArrayCRDT (Go) with snapshot-based measurements
- Tests all JavaScript competitors with real memory tracking
- Uses 60-second cooldowns between runs for reliable measurements
- Generates timestamped results in `data/benchmark_runs/`
- Supports version comparison in the web UI

### Individual Benchmarks

**MArrayCRDT only:**
```bash
cd benchmarks
go run marraycrdt_simulation.go
```

**Specific competitor:**
```bash
cd competitors/automerge
node --expose-gc simulation.js
```

### Web Visualization

```bash
cd web
node server.js
```

Open http://localhost:3000 for interactive performance charts with:
- Real-time data loading from benchmark runs
- Version comparison between different benchmark sessions
- Memory usage and throughput visualization
- Operation count scaling analysis

## Key Features

### MArrayCRDT Capabilities
- **Array Operations**: Insert, delete, move, swap, sort, reverse, rotate
- **Strong Consistency**: Vector clock-based conflict resolution
- **Move Support**: Support for element repositioning

### Benchmarking Framework
- **Average Memory Tracking**: Samples memory every 100 operations for stable measurements
- **Process Isolation**: 60-second cooldowns between competitor runs
- **Academic Dataset**: Uses real editing trace from Kleppmann et al. research

### Competitor Analysis
- **Automerge**: Rich collaboration features with full operation history
- **Yjs**: High-performance text editing optimization
- **Loro**: Text CRDT with optional MovableList (comparable to MArrayCRDT)
- **Baseline**: Plain JavaScript Array for overhead comparison

## Architecture

### Memory Management
- **Go advantages**: Predictable memory layout, efficient GC, direct struct control
- **JavaScript overhead**: Object wrapping, prototype chains, V8 runtime costs
- **Real measurements**: Average memory sampling during execution, not endpoint diffs

### CRDT Design  
- **Element-level tracking**: Each array element has unique ID, vector clock, position metadata
- **Position-based ordering**: Floating-point positions enable efficient move operations
- **Conflict resolution**: Last-Writer-Wins with deterministic tiebreaking
- **Operation support**: Beyond text editing - full array manipulation capabilities
