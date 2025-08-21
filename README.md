# MArrayCRDT - Movable Array Conflict-free Replicated Data Type

A high-performance CRDT implementation with comprehensive benchmarking against Automerge.

## Repository Structure

```
├── crdt/          # Core CRDT implementation (Go)
├── benchmarks/    # Performance benchmarks and simulations (Go)
├── simulation/    # MArrayCRDT simulation results (CSV/JSON)
├── competitors/   # Competitor CRDT benchmarks (Node.js/JavaScript)  
├── web/           # Web-based visualization (Node.js/JavaScript)
├── data/          # Input data and competitor results
└── go.mod         # Go module definition
```

## Quick Start

### Running Benchmarks

**MArrayCRDT Benchmark:**
```bash
cd benchmarks
go run .
```

**Competitor Benchmarks:**
```bash  
cd competitors
node run_all.js
```

**Complete Comparison:**
```bash
# Run both MArrayCRDT and competitors
cd benchmarks && go run . && cd ../competitors && node run_all.js
```

This will:
- Run comprehensive benchmarks at multiple scales (1k-50k operations)
- Test MArrayCRDT, Automerge, Yjs, and Loro with identical conditions
- Generate performance comparison data in `data/` directory
- Use real editing traces from academic research (Kleppmann et al.)

### Web Visualization

```bash
# From the project root
cd web
npm install
npm start
```

Then open http://localhost:3000 to view interactive performance charts.

## Components

### Core CRDT (`crdt/`)
- `marraycrdt.go` - Main CRDT implementation
- `marraycrdt_test.go` - Unit tests
- `example_test.go` - Usage examples

### Benchmarks (`benchmarks/`)
- `main.go` - Main benchmark runner
- `comprehensive_benchmark.go` - Multi-scale performance tests
- `automerge_trace_simulation.go` - Real editing trace simulation

### Simulation Results (`simulation/`)
- `marraycrdt_results.csv` - MArrayCRDT performance results
- `marraycrdt_comprehensive_benchmark.json` - Detailed benchmark data
- `marraycrdt_automerge_metrics.json` - Trace simulation metrics

### Competitors (`competitors/`)
- `automerge/` - Automerge CRDT benchmark
- `yjs/` - Yjs CRDT benchmark  
- `loro/` - Loro CRDT benchmark
- `run_all.js` - Master runner for all competitor tests

### Web Visualization (`web/`)
- `server.js` - Express server for serving data
- `public/app.js` - Chart.js visualization
- `public/index.html` - Web interface

### Data (`data/`)
- `paper.json` - Real editing trace (259k operations from writing a LaTeX paper)
- `competitors_comparison.csv` - Competitor benchmark results

## Key Features

- **Real Memory Measurements**: Actual runtime memory usage (not estimates)
- **Academic Benchmarks**: Uses real editing traces from CRDT research
- **Multi-Scale Testing**: Performance analysis from 1k to 50k operations  
- **Interactive Visualization**: Browser-based charts and comparisons
- **Comprehensive Metrics**: Throughput, memory usage, scalability analysis

## Performance Highlights

MArrayCRDT vs Automerge (at 50k operations):
- **Memory Efficiency**: 10x less memory usage (46MB vs 466MB)
- **Scalability**: Consistent performance characteristics
- **Real Workloads**: Tested with actual collaborative editing scenarios