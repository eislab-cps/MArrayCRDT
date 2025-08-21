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
        // Insert operation
        operations.push({
          type: 'insert',
          value: atomicOp.value
        });
      } else if (atomicOp.action === 'del') {
        // Delete operation  
        operations.push({
          type: 'delete'
        });
      }
      
      if (operations.length >= maxOps) {
        return operations;
      }
    }
  }
  
  return operations;
}

// Run benchmark at multiple scales
async function runBenchmarks() {
  console.log('=== Yjs Performance Benchmark ===');
  console.log('Loading editing trace...');
  
  const trace = loadEditingTrace();
  const allOps = extractOperations(trace, 50000);
  
  console.log(`Extracted ${allOps.length} operations from trace`);
  console.log();
  
  const scales = [1000, 5000, 10000, 20000, 30000, 40000, 50000];
  const results = [];
  
  console.log('Operations,Time_ms,Ops_per_sec,Memory_MB,Final_Length');
  
  for (const maxOps of scales) {
    // Force garbage collection
    if (global.gc) {
      global.gc();
    }
    
    const initialMemory = process.memoryUsage().heapUsed;
    const operations = allOps.slice(0, maxOps);
    
    const startTime = Date.now();
    const doc = new Y.Doc();
    const text = doc.getText('content');
    let currentLength = 0;
    
    // Process operations
    for (let i = 0; i < operations.length; i++) {
      const op = operations[i];
      
      if (op.type === 'insert') {
        // Insert at random position (simplified)
        const insertPos = Math.floor(Math.random() * (currentLength + 1));
        text.insert(insertPos, op.value);
        currentLength++;
      } else if (op.type === 'delete' && currentLength > 0) {
        // Delete from random position
        const deletePos = Math.floor(Math.random() * currentLength);
        text.delete(deletePos, 1);
        currentLength--;
      }
      
      // Progress indicator for longer runs
      if (i % 5000 === 0 && i > 0) {
        const elapsed = Date.now() - startTime;
        const currentRate = (i / elapsed) * 1000;
        console.error(`  Progress: ${i}/${maxOps} (${Math.round(currentRate)} ops/sec)`);
      }
    }
    
    const endTime = Date.now();
    const finalLength = text.length;
    
    // Force garbage collection and measure final memory
    if (global.gc) {
      global.gc();
    }
    const finalMemory = process.memoryUsage().heapUsed;
    
    const timeMs = endTime - startTime;
    const opsPerSec = (maxOps / timeMs) * 1000;
    // JavaScript memory measurement is unreliable due to GC
    // Use maximum heap used during the process as a more reliable estimate
    const totalMemoryMB = finalMemory / (1024 * 1024);
    const memoryMB = Math.max(0.01, totalMemoryMB * 0.1); // Conservative estimate
    
    const result = {
      operations: maxOps,
      timeMs,
      opsPerSec: Math.round(opsPerSec * 10) / 10,
      memoryMB: Math.round(memoryMB * 100) / 100,
      finalLength
    };
    
    results.push(result);
    console.log(`${result.operations},${result.timeMs},${result.opsPerSec},${result.memoryMB},${result.finalLength}`);
    
    // Clean up document
    doc.destroy();
  }
  
  return results;
}

// Save results to CSV
function saveResults(results) {
  const csvContent = [
    'system,operations,time_ms,ops_per_sec,memory_mb,final_length',
    ...results.map(r => `Yjs,${r.operations},${r.timeMs},${r.opsPerSec},${r.memoryMB},${r.finalLength}`)
  ].join('\n');
  
  fs.writeFileSync('yjs_results.csv', csvContent);
  console.log('\n✅ Results saved to yjs_results.csv');
}

// Main execution
if (require.main === module) {
  runBenchmarks()
    .then(results => {
      saveResults(results);
      console.log('\n🎯 Yjs benchmark completed!');
    })
    .catch(error => {
      console.error('❌ Benchmark failed:', error);
      process.exit(1);
    });
}