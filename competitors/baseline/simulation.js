// Baseline performance simulation using plain JavaScript Array
// No CRDT overhead - pure JavaScript operations for comparison
const fs = require('fs');
const path = require('path');

// Load the editing trace data
function loadEditingTrace() {
  const tracePath = path.join(__dirname, '../../data/paper.json');
  const fileContent = fs.readFileSync(tracePath, 'utf8');
  const lines = fileContent.trim().split('\n');
  return lines.map(line => JSON.parse(line));
}

// Extract operations from the trace
function extractOperations(trace, maxOps = 50000) {
  const operations = [];
  let currentLength = 0;
  
  for (const change of trace) {
    if (operations.length >= maxOps) break;
    
    if (change.ops && Array.isArray(change.ops)) {
      for (const op of change.ops) {
        if (operations.length >= maxOps) break;
        
        if (op.action === 'set' && op.insert === true && typeof op.value === 'string') {
          // Convert position-based insert to array index
          const insertIndex = Math.floor(Math.random() * (currentLength + 1));
          operations.push({
            type: 'insert',
            position: insertIndex,
            value: op.value
          });
          currentLength++;
        } else if (op.action === 'del') {
          if (currentLength > 0) {
            // Random delete for baseline
            const deleteIndex = Math.floor(Math.random() * currentLength);
            operations.push({
              type: 'delete',
              position: deleteIndex,
              length: 1
            });
            currentLength--;
          }
        }
      }
    }
  }
  
  console.log(`Extracted ${operations.length} operations from trace`);
  return operations;
}

// Measure memory usage
function measureMemory() {
  if (global.gc) {
    global.gc();
  }
  const memUsage = process.memoryUsage();
  return memUsage.heapUsed / 1024 / 1024; // Convert to MB
}

// Run benchmark at specific operation count
function runBenchmark(operations, targetOps) {
  const operationsToRun = operations.slice(0, targetOps);
  
  // Initialize plain JavaScript array
  const content = [];
  
  const startTime = Date.now();
  const startMemory = measureMemory();
  
  // Execute operations
  for (const op of operationsToRun) {
    if (op.type === 'insert') {
      // Insert character at position
      const pos = Math.min(op.position, content.length);
      content.splice(pos, 0, op.value);
    } else if (op.type === 'delete') {
      // Delete character at position
      if (content.length > 0) {
        const pos = Math.min(op.position, content.length - 1);
        content.splice(pos, 1);
      }
    }
  }
  
  const endTime = Date.now();
  const endMemory = measureMemory();
  
  const timeMs = endTime - startTime;
  const opsPerSec = Math.round((targetOps / timeMs) * 1000);
  const memoryMb = (endMemory - startMemory).toFixed(2);
  const finalLength = content.length;
  
  return {
    operations: targetOps,
    timeMs,
    opsPerSec,
    memoryMb: parseFloat(memoryMb),
    finalLength
  };
}

// Main benchmark function
async function runBenchmarks() {
  console.log('=== JavaScript Array Baseline Benchmark ===');
  console.log('Using plain JavaScript Array (no CRDT overhead)');
  console.log('Loading editing trace...');
  
  const trace = loadEditingTrace();
  const operations = extractOperations(trace, 50000);
  
  const testSizes = [1000, 5000, 10000, 20000, 30000, 40000, 50000];
  const results = [];
  
  console.log('\\nOperations,Time_ms,Ops_per_sec,Memory_MB,Final_Length');
  
  for (const size of testSizes) {
    const result = runBenchmark(operations, size);
    results.push(result);
    console.log(`${result.operations},${result.timeMs},${result.opsPerSec},${result.memoryMb},${result.finalLength}`);
  }
  
  // Save results to CSV
  const csvHeader = 'system,operations,time_ms,ops_per_sec,memory_mb,final_length';
  const csvRows = results.map(r => 
    `Baseline,${r.operations},${r.timeMs},${r.opsPerSec},${r.memoryMb},${r.finalLength}`
  );
  const csvContent = [csvHeader, ...csvRows].join('\\n');
  
  const outputPath = path.join(__dirname, 'baseline_results.csv');
  fs.writeFileSync(outputPath, csvContent);
  
  console.log('\\n✅ Results saved to baseline_results.csv');
  console.log('\\n🎯 JavaScript Array baseline benchmark completed!');
  console.log('\\nThis provides the theoretical minimum performance (no CRDT overhead)');
  console.log('Compare against CRDT implementations to measure coordination overhead.');
}

// Run if called directly
if (require.main === module) {
  runBenchmarks().catch(console.error);
}

module.exports = { runBenchmarks };