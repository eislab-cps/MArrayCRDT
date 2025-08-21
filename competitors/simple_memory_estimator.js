// Simple memory estimator for CRDT comparisons
// Instead of unreliable heap measurements, estimate based on data structure size

function estimateMemoryUsage(dataStructure, operations) {
  // Base estimation strategies for different CRDT types
  
  if (typeof dataStructure === 'string') {
    // Simple string - just the character count
    return dataStructure.length * 2 / 1024 / 1024; // 2 bytes per char, convert to MB
  }
  
  if (Array.isArray(dataStructure)) {
    // JavaScript Array - estimate based on length and element size
    const avgElementSize = 50; // Rough estimate for mixed content
    return (dataStructure.length * avgElementSize) / 1024 / 1024;
  }
  
  // For complex CRDT objects, estimate based on operations and final size
  const operationCount = operations || 1000;
  const estimatedOverhead = Math.max(1, Math.log10(operationCount)) * 0.5; // Logarithmic overhead
  
  return estimatedOverhead;
}

function createReliableMemoryMeasurement() {
  return {
    // Force a clean baseline
    baseline: null,
    measurements: [],
    
    establishBaseline() {
      // Multiple GC attempts
      for (let i = 0; i < 3; i++) {
        if (global.gc) global.gc();
      }
      
      const mem = process.memoryUsage();
      this.baseline = mem.heapUsed;
      console.log(`Established baseline: ${(this.baseline / 1024 / 1024).toFixed(1)} MB`);
    },
    
    measureDelta(operationCount, finalDataSize) {
      // Strategy 1: Try heap measurement with protection against negative values
      let heapDelta = 0;
      
      if (global.gc && this.baseline) {
        global.gc();
        const current = process.memoryUsage().heapUsed;
        heapDelta = Math.max(0, (current - this.baseline) / 1024 / 1024);
      }
      
      // Strategy 2: Size-based estimation as fallback/verification
      const sizeEstimate = Math.max(
        operationCount * 0.0001, // 0.1KB per operation minimum
        finalDataSize * 0.000002  // 2 bytes per character
      );
      
      // Use heap measurement if reasonable, otherwise fall back to estimation
      if (heapDelta > 0 && heapDelta < 1000) { // Sanity check: less than 1GB
        return Math.max(heapDelta, sizeEstimate);
      } else {
        console.log(`Using size estimation (heap delta was ${heapDelta.toFixed(2)} MB)`);
        return sizeEstimate;
      }
    }
  };
}

module.exports = {
  estimateMemoryUsage,
  createReliableMemoryMeasurement
};