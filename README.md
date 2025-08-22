# MArrayCRDT - Movable Array Conflict-free Replicated Data Type

A high-performance CRDT implementation in Go that supports full array operations including move, sort, reverse, and swap, with comprehensive benchmarking against leading JavaScript CRDT libraries.

## Repository Structure

```
├── crdt/                    # Core MArrayCRDT implementation (Go)
├── benchmarks/             # MArrayCRDT performance benchmarks (Go)
├── competitors/            # Competitor CRDT benchmarks (JavaScript)
│   ├── automerge/         # Automerge CRDT benchmarks
│   ├── yjs/               # Yjs CRDT benchmarks  
│   ├── loro/              # Loro CRDT benchmarks (Text + MovableList)
│   └── baseline/          # JavaScript Array baseline
├── web/                    # Interactive web visualization
├── data/                   # Benchmark data and editing traces
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
- **Full Array Operations**: Insert, delete, move, swap, sort, reverse, rotate
- **Strong Consistency**: Vector clock-based conflict resolution
- **Move Support**: First-class support for element repositioning
- **Memory Efficient**: 2.5-4x more memory efficient than JavaScript CRDTs
- **Replica-based**: Changed from "siteID" to "replicaID" terminology

### Benchmarking Framework
- **Real Memory Measurement**: No artificial scaling - actual `process.memoryUsage()` sampling
- **Single-run Snapshots**: Efficient benchmarking like MArrayCRDT (no re-running from scratch)
- **Average Memory Tracking**: Samples memory every 100 operations for stable measurements
- **Process Isolation**: 60-second cooldowns between competitor runs
- **Academic Dataset**: Uses real editing trace from Kleppmann et al. research

### Competitor Analysis
- **Automerge**: Rich collaboration features with full operation history
- **Yjs**: High-performance text editing optimization
- **Loro**: Text CRDT with optional MovableList (comparable to MArrayCRDT)
- **Baseline**: Plain JavaScript Array for overhead comparison

## Performance Results

**Memory Efficiency (at 30k+ operations):**
- **MArrayCRDT (Go)**: ~26-44 MB
- **JavaScript CRDTs**: ~250-425 MB  
- **Efficiency Gain**: 2.5-4x more memory efficient

**Throughput Characteristics:**
- **MArrayCRDT**: Optimized for shopping list use cases (<1k operations)
- **1k operations**: 13,050+ ops/sec - excellent for real-world usage
- **JavaScript CRDTs**: Various optimization profiles for text editing

**Use Case Validation:**
- Shopping lists typically have 10-30 items with ~500-1000 total operations
- MArrayCRDT performance is more than adequate for intended use case
- Scalability issues at 30k+ operations are not relevant for target scenarios

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

## Development Notes

- All artificial memory scaling has been removed for authentic benchmarking
- Cooldown periods ensure reliable measurements between JavaScript runs
- Web UI shows timestamped benchmark versions for comparison
- Single-run snapshot approach matches MArrayCRDT efficiency patterns