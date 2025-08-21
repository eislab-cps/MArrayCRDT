// Improved memory measurement utilities for JavaScript competitors
// Addresses issues with negative memory readings and GC timing

class MemoryMonitor {
  constructor() {
    this.baseline = 0;
    this.measurements = [];
  }
  
  // Establish a memory baseline after forcing garbage collection
  establishBaseline() {
    // Force multiple GC cycles to establish stable baseline
    for (let i = 0; i < 5; i++) {
      if (global.gc) global.gc();
    }
    
    // Wait a bit for async cleanup
    const start = Date.now();
    while (Date.now() - start < 50) {
      // Small delay
    }
    
    const memUsage = process.memoryUsage();
    this.baseline = memUsage.heapUsed;
    
    console.log(`Memory baseline established: ${(this.baseline / 1024 / 1024).toFixed(2)} MB`);
    return this.baseline;
  }
  
  // Measure current memory usage relative to baseline
  measureCurrent() {
    // Force GC before measurement
    if (global.gc) {
      global.gc();
    }
    
    const memUsage = process.memoryUsage();
    const currentUsed = memUsage.heapUsed;
    const deltaBytes = Math.max(0, currentUsed - this.baseline); // Ensure non-negative
    const deltaMB = deltaBytes / 1024 / 1024;
    
    this.measurements.push({
      timestamp: Date.now(),
      heapUsed: currentUsed,
      baseline: this.baseline,
      deltaMB: deltaMB,
      totalHeapMB: memUsage.heapTotal / 1024 / 1024
    });
    
    return deltaMB;
  }
  
  // Get peak memory usage during benchmark
  getPeakMemory() {
    if (this.measurements.length === 0) return 0;
    return Math.max(...this.measurements.map(m => m.deltaMB));
  }
  
  // Reset for new benchmark
  reset() {
    this.baseline = 0;
    this.measurements = [];
  }
  
  // Get detailed memory info for debugging
  getDetailedReport() {
    return {
      baseline: this.baseline,
      measurements: this.measurements,
      peak: this.getPeakMemory(),
      final: this.measurements.length > 0 ? this.measurements[this.measurements.length - 1] : null
    };
  }
}

// Isolated benchmark runner to minimize memory pollution
function runIsolatedBenchmark(benchmarkFn, targetOps) {
  const monitor = new MemoryMonitor();
  
  // Clear any existing state
  if (global.gc) {
    for (let i = 0; i < 3; i++) global.gc();
  }
  
  // Establish clean baseline
  monitor.establishBaseline();
  
  const startTime = process.hrtime.bigint();
  
  try {
    // Run the benchmark function
    const result = benchmarkFn(targetOps, monitor);
    
    const endTime = process.hrtime.bigint();
    const elapsedMs = Number(endTime - startTime) / 1_000_000;
    
    // Final memory measurement
    const finalMemory = monitor.measureCurrent();
    const peakMemory = monitor.getPeakMemory();
    
    return {
      ...result,
      timeMs: Math.round(elapsedMs),
      opsPerSec: Math.round((targetOps / elapsedMs) * 1000),
      memoryMb: parseFloat(Math.max(finalMemory, peakMemory).toFixed(2)),
      peakMemoryMb: parseFloat(peakMemory.toFixed(2)),
      memoryDetails: monitor.getDetailedReport()
    };
    
  } catch (error) {
    console.error(`Benchmark failed for ${targetOps} ops:`, error.message);
    return {
      operations: targetOps,
      timeMs: 0,
      opsPerSec: 0,
      memoryMb: 0,
      finalLength: 0,
      error: error.message
    };
  }
}

module.exports = {
  MemoryMonitor,
  runIsolatedBenchmark
};