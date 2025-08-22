// Yjs performance simulation using Kleppmann editing trace
const Y = require('yjs');
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
  
  for (const op of trace) {
    for (const atomicOp of op.ops || []) {
      if (atomicOp.action === 'set' && atomicOp.insert) {
        operations.push({ 
          type: 'insert', 
          value: atomicOp.value 
        });
      } else if (atomicOp.action === 'del') {
        operations.push({ type: 'delete' });
      }
      
      if (operations.length >= maxOps) {
        return operations;
      }
    }
  }
  
  return operations;
}

// Run benchmark with snapshots at milestone operations (single run)
function runBenchmarkWithSnapshots(operations) {
  const snapshotPoints = [1000, 5000, 10000, 20000, 30000, 40000, 50000];
  const results = [];
  const memorySamples = [];
  
  // Initialize Y.Doc
  const doc = new Y.Doc();
  const ytext = doc.getText('content');
  
  const startTime = Date.now();
  let nextSnapshotIdx = 0;
  let opCount = 0;
  
  console.log('Running single benchmark with snapshots at milestone operations...');
  console.log('Operations,Time_ms,Ops_per_sec,Avg_Memory_MB,Final_Length');
  
  for (let i = 0; i < operations.length && nextSnapshotIdx < snapshotPoints.length; i++) {
    const op = operations[i];
    
    // Apply operation
    if (op.type === 'insert') {
      ytext.insert(ytext.length, op.value);
    } else if (op.type === 'delete' && ytext.length > 0) {
      const deletePos = Math.floor(Math.random() * ytext.length);
      ytext.delete(deletePos, 1);
    }
    
    opCount++;
    
    // Sample memory every 100 operations
    if (opCount % 100 === 0) {
      if (global.gc) global.gc();
      memorySamples.push(process.memoryUsage().heapUsed);
    }
    
    // Check if we've reached a snapshot point
    if (opCount === snapshotPoints[nextSnapshotIdx]) {
      const elapsed = Date.now() - startTime;
      
      // Calculate average memory from samples
      const avgMemoryBytes = memorySamples.length > 0 
        ? memorySamples.reduce((a, b) => a + b, 0) / memorySamples.length
        : process.memoryUsage().heapUsed;
      const avgMemoryMB = avgMemoryBytes / (1024 * 1024);
      
      const opsPerSec = parseFloat(((opCount / elapsed) * 1000).toFixed(1));
      const finalLength = ytext.length;
      
      const result = {
        operations: opCount,
        timeMs: elapsed,
        opsPerSec,
        memoryMb: parseFloat(avgMemoryMB.toFixed(2)),
        finalLength
      };
      
      results.push(result);
      console.log(`${result.operations},${result.timeMs},${result.opsPerSec},${result.memoryMb},${result.finalLength}`);
      
      nextSnapshotIdx++;
    }
    
    // Progress reporting
    if (opCount % 5000 === 0 && opCount > 0) {
      const elapsed = Date.now() - startTime;
      const currentOpsPerSec = Math.round((opCount / elapsed) * 1000);
      console.error(`  Progress: ${opCount} operations (${currentOpsPerSec} ops/sec)`);
    }
  }
  
  return results;
}

async function runBenchmarks() {
  console.log('=== Yjs Performance Benchmark ===');
  console.log('Loading editing trace...');
  
  const trace = loadEditingTrace();
  const allOps = extractOperations(trace, 50000);
  console.log(`Extracted ${allOps.length} operations from trace\n`);
  
  // Run single benchmark with snapshots
  const results = runBenchmarkWithSnapshots(allOps);
  
  // Save results to CSV
  const csvHeader = 'system,operations,time_ms,ops_per_sec,memory_mb,final_length';
  const csvRows = results.map(r => 
    `Yjs,${r.operations},${r.timeMs},${r.opsPerSec},${r.memoryMb},${r.finalLength}`
  );
  const csvContent = [csvHeader, ...csvRows].join('\n');
  
  fs.writeFileSync(path.join(__dirname, 'yjs_results.csv'), csvContent);
  console.log('\n✅ Results saved to yjs_results.csv');
  console.log('🎯 Yjs benchmark completed!');
}

if (require.main === module) {
  runBenchmarks().catch(console.error);
}

module.exports = { runBenchmarks };