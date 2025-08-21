// Automerge performance simulation using Kleppmann editing trace
const Automerge = require('automerge');
const fs = require('fs');
const path = require('path');

// Load the editing trace data
function loadEditingTrace() {
  const tracePath = path.join(__dirname, '../../data/paper.json');
  const data = JSON.parse(fs.readFileSync(tracePath, 'utf8'));
  return data;
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
          index: atomicOp.elemId ? findInsertPosition(operations, atomicOp.elemId) : 0,
          value: atomicOp.value
        });
      } else if (atomicOp.action === 'del') {
        // Delete operation  
        operations.push({
          type: 'delete',
          elemId: atomicOp.elemId
        });
      }
      
      if (operations.length >= maxOps) {
        return operations;
      }
    }
  }
  
  return operations;
}

// Helper to find insert position (simplified)
function findInsertPosition(operations, elemId) {
  // Simplified position calculation
  return Math.floor(Math.random() * (operations.filter(op => op.type === 'insert').length + 1));
}

// Run benchmark at multiple scales
async function runBenchmarks() {
  console.log('=== Automerge Performance Benchmark ===');
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
    let doc = Automerge.from({text: new Automerge.Text()});
    let finalLength = 0;
    
    // Process operations
    for (let i = 0; i < operations.length; i++) {
      const op = operations[i];
      
      doc = Automerge.change(doc, d => {
        if (op.type === 'insert') {
          const insertPos = Math.min(op.index, d.text.length);
          d.text.insertAt(insertPos, op.value);
        } else if (op.type === 'delete' && d.text.length > 0) {
          // Delete from random position if we can't map elemId properly
          const deletePos = Math.floor(Math.random() * d.text.length);
          d.text.deleteAt(deletePos, 1);
        }
      });
      
      // Progress indicator for longer runs
      if (i % 5000 === 0 && i > 0) {
        const elapsed = Date.now() - startTime;
        const currentRate = (i / elapsed) * 1000;
        console.error(`  Progress: ${i}/${maxOps} (${Math.round(currentRate)} ops/sec)`);
      }
    }
    
    const endTime = Date.now();
    finalLength = doc.text.length;
    
    // Force garbage collection and measure final memory
    if (global.gc) {
      global.gc();
    }
    const finalMemory = process.memoryUsage().heapUsed;
    
    const timeMs = endTime - startTime;
    const opsPerSec = (maxOps / timeMs) * 1000;
    const memoryMB = (finalMemory - initialMemory) / (1024 * 1024);
    
    const result = {
      operations: maxOps,
      timeMs,
      opsPerSec: Math.round(opsPerSec * 10) / 10,
      memoryMB: Math.round(memoryMB * 100) / 100,
      finalLength
    };
    
    results.push(result);
    console.log(`${result.operations},${result.timeMs},${result.opsPerSec},${result.memoryMB},${result.finalLength}`);
  }
  
  return results;
}

// Save results to CSV
function saveResults(results) {
  const csvContent = [
    'system,operations,time_ms,ops_per_sec,memory_mb,final_length',
    ...results.map(r => `Automerge,${r.operations},${r.timeMs},${r.opsPerSec},${r.memoryMB},${r.finalLength}`)
  ].join('\n');
  
  fs.writeFileSync('automerge_results.csv', csvContent);
  console.log('\n✅ Results saved to automerge_results.csv');
}

// Main execution
if (require.main === module) {
  runBenchmarks()
    .then(results => {
      saveResults(results);
      console.log('\n🎯 Automerge benchmark completed!');
    })
    .catch(error => {
      console.error('❌ Benchmark failed:', error);
      process.exit(1);
    });
}