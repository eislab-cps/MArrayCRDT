# MArrayCRDT Simulation Results

This directory contains the results from MArrayCRDT performance simulations using the Kleppmann editing trace.

## Files Generated

- `marraycrdt_results.csv` - Performance results in CSV format for visualization
- `marraycrdt_comprehensive_benchmark.json` - Detailed benchmark results in JSON format  
- `marraycrdt_automerge_metrics.json` - Detailed metrics from trace simulation

## Data Format

### CSV Format (`marraycrdt_results.csv`)
```csv
system,operations,time_ms,ops_per_sec,memory_mb,insert_ops,delete_ops,final_length
MArrayCRDT,1000,71.4,14014.5,1.13,981,18,971
...
```

**Columns:**
- `system` - CRDT system name (MArrayCRDT, Baseline)
- `operations` - Number of operations processed
- `time_ms` - Total execution time in milliseconds
- `ops_per_sec` - Operations per second (throughput)
- `memory_mb` - Memory usage in megabytes
- `insert_ops` - Number of insert operations
- `delete_ops` - Number of delete operations  
- `final_length` - Final document length in characters

## Integration

Results are automatically loaded by the web visualization at http://localhost:3000

## Regeneration

To regenerate results, run:
```bash
cd ../benchmarks
go run .
```

This will overwrite existing files in this directory with fresh benchmark results.