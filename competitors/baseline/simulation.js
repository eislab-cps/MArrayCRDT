// Fixed Baseline performance simulation  
const fs = require('fs');
const path = require('path');

// Load data from root data directory
const DATA_PATH = path.join(__dirname, '../../data/paper.json');

// Load the editing trace data
function loadEditingTrace() {
  const tracePath = DATA_PATH;
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
          const insertIndex = Math.floor(Math.random() * (currentLength + 1));
          operations.push({ type: 'insert', position: insertIndex, value: op.value });
          currentLength++;
        } else if (op.action === 'del') {
          if (currentLength > 0) {
            const deleteIndex = Math.floor(Math.random() * currentLength);
            operations.push({ type: 'delete', position: deleteIndex, length: 1 });
            currentLength--;
          }
        }
      }
    }
  }
  
  console.log(`Extracted ${operations.length} operations from trace`);
  return operations;
}

// Run benchmark at specific operation count
function runBenchmark(operations, targetOps) {
  const operationsToRun = operations.slice(0, targetOps);
  const content = [];
  
  const startTime = Date.now();
  
  for (const op of operationsToRun) {
    if (op.type === 'insert') {
      const pos = Math.min(op.position, content.length);
      content.splice(pos, 0, op.value);
    } else if (op.type === 'delete') {
      if (content.length > 0) {
        const pos = Math.min(op.position, content.length - 1);
        content.splice(pos, 1);
      }
    }
  }
  
  const endTime = Date.now();
  const timeMs = endTime - startTime;
  const opsPerSec = Math.round((targetOps / timeMs) * 1000);
  
  // Memory estimation for JavaScript Array
  const memoryMb = (content.length * 2) / 1024 / 1024; // 2 bytes per char
  
  return {
    operations: targetOps,
    timeMs,
    opsPerSec,
    memoryMb: parseFloat(memoryMb.toFixed(2)),
    finalLength: content.length
  };
}

async function runBenchmarks() {
  console.log('=== JavaScript Array Baseline Benchmark ===');
  console.log('Using plain JavaScript Array (no CRDT overhead)');
  console.log('Loading editing trace...');
  
  const trace = loadEditingTrace();
  const operations = extractOperations(trace, 50000);
  
  const testSizes = [1000, 5000, 10000, 20000, 30000, 40000, 50000];
  const results = [];
  
  console.log('\nOperations,Time_ms,Ops_per_sec,Memory_MB,Final_Length');
  
  for (const size of testSizes) {
    const result = runBenchmark(operations, size);
    results.push(result);
    console.log(`${result.operations},${result.timeMs},${result.opsPerSec},${result.memoryMb},${result.finalLength}`);
  }
  
  const csvHeader = 'system,operations,time_ms,ops_per_sec,memory_mb,final_length';
  const csvRows = results.map(r => 
    `Baseline,${r.operations},${r.timeMs},${r.opsPerSec},${r.memoryMb},${r.finalLength}`
  );
  const csvContent = [csvHeader, ...csvRows].join('\n');
  
  fs.writeFileSync(path.join(__dirname, 'baseline_results.csv'), csvContent);
  console.log('\n✅ Results saved to baseline_results.csv\n🎯 JavaScript Array baseline benchmark completed!');
}

if (require.main === module) {
  runBenchmarks().catch(console.error);
}

module.exports = { runBenchmarks };