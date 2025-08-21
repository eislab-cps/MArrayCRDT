#!/usr/bin/env node
// Reliable benchmark runner using process isolation for accurate memory measurement

const { execSync, spawn } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

// System memory monitoring (Linux/macOS)
function getProcessMemory(pid) {
  try {
    if (os.platform() === 'linux') {
      // Linux: Read from /proc/PID/status
      const status = fs.readFileSync(`/proc/${pid}/status`, 'utf8');
      const vmRSSMatch = status.match(/VmRSS:\\s+(\\d+)\\s+kB/);
      return vmRSSMatch ? parseInt(vmRSSMatch[1]) / 1024 : null; // Convert KB to MB
    } else if (os.platform() === 'darwin') {
      // macOS: Use ps command
      const output = execSync(`ps -o rss= -p ${pid}`, {encoding: 'utf8'});
      return parseInt(output.trim()) / 1024; // Convert KB to MB
    }
  } catch (error) {
    return null;
  }
  return null;
}

// Monitor memory usage of a spawned process
async function monitorProcessMemory(childProcess, samples = 10, intervalMs = 100) {
  const memories = [];
  const startTime = Date.now();
  
  return new Promise((resolve) => {
    const monitor = setInterval(() => {
      const memory = getProcessMemory(childProcess.pid);
      if (memory !== null) {
        memories.push({
          timestamp: Date.now() - startTime,
          memoryMB: memory
        });
      }
      
      if (memories.length >= samples || !childProcess.killed) {
        clearInterval(monitor);
      }
    }, intervalMs);
    
    childProcess.on('exit', () => {
      clearInterval(monitor);
      const maxMemory = memories.length > 0 ? Math.max(...memories.map(m => m.memoryMB)) : 0;
      const avgMemory = memories.length > 0 ? memories.reduce((sum, m) => sum + m.memoryMB, 0) / memories.length : 0;
      
      resolve({
        peakMemoryMB: maxMemory,
        avgMemoryMB: avgMemory,
        samples: memories.length,
        memoryTrace: memories
      });
    });
  });
}

// Run a single competitor benchmark in isolated process with memory monitoring
async function runIsolatedBenchmark(competitorPath, maxOps = 50000) {
  return new Promise((resolve, reject) => {
    console.log(`🔍 Running isolated benchmark: ${competitorPath}`);
    
    // Spawn isolated Node.js process with reasonable memory limits
    const childProcess = spawn('node', [
      '--expose-gc',
      '--max-old-space-size=2048', // 2GB limit - enough for any reasonable CRDT
      '--no-warnings',
      competitorPath
    ], {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: path.dirname(competitorPath)
    });
    
    let output = '';
    let errorOutput = '';
    
    childProcess.stdout.on('data', (data) => {
      output += data.toString();
    });
    
    childProcess.stderr.on('data', (data) => {
      errorOutput += data.toString();
    });
    
    // Monitor memory usage
    const memoryMonitoring = monitorProcessMemory(childProcess);
    
    childProcess.on('exit', async (code) => {
      const memoryStats = await memoryMonitoring;
      
      if (code !== 0) {
        console.error(`❌ Benchmark failed with code ${code}`);
        console.error(`Error: ${errorOutput}`);
        reject(new Error(`Benchmark failed: ${errorOutput}`));
        return;
      }
      
      try {
        // Parse the CSV output from the competitor benchmark
        const csvMatch = output.match(/Operations,Time_ms,Ops_per_sec,Memory_MB,Final_Length([\\s\\S]*?)✅/);
        if (!csvMatch) {
          throw new Error('Could not parse benchmark output');
        }
        
        const csvLines = csvMatch[1].trim().split('\\n');
        const results = [];
        
        for (const line of csvLines) {
          if (line.includes(',')) {
            const [ops, timeMs, opsPerSec, memoryMB, finalLength] = line.split(',');
            results.push({
              operations: parseInt(ops),
              timeMs: parseFloat(timeMs),
              opsPerSec: parseFloat(opsPerSec),
              memoryMB_internal: parseFloat(memoryMB), // Internal JS measurement
              memoryMB_external: memoryStats.peakMemoryMB, // External system measurement
              finalLength: parseInt(finalLength),
              memoryTrace: memoryStats.memoryTrace
            });
          }
        }
        
        console.log(`✅ Completed: ${results.length} data points`);
        console.log(`📊 Memory: Internal=${results[results.length-1]?.memoryMB_internal?.toFixed(2)}MB, External=${memoryStats.peakMemoryMB?.toFixed(2)}MB`);
        
        resolve({
          results,
          memoryStats,
          competitorName: path.basename(path.dirname(competitorPath))
        });
        
      } catch (parseError) {
        console.error(`❌ Failed to parse results: ${parseError.message}`);
        console.log('Raw output:', output);
        reject(parseError);
      }
    });
  });
}

// Run all competitors with reliable memory monitoring
async function runReliableBenchmarks() {
  console.log('🔬 Starting Reliable CRDT Benchmark Suite');
  console.log('Using process isolation and external memory monitoring\\n');
  
  const competitors = [
    '/home/caslun/github/MArrayCRDT/competitors/automerge/simulation.js',
    '/home/caslun/github/MArrayCRDT/competitors/baseline/simulation.js'
    // Add others after testing
  ];
  
  const allResults = [];
  
  for (const competitor of competitors) {
    try {
      console.log(`\\n${'='.repeat(50)}`);
      
      // Wait between benchmarks to let system clear memory
      if (allResults.length > 0) {
        console.log('⏳ Waiting 3 seconds for system cleanup...');
        await new Promise(resolve => setTimeout(resolve, 3000));
      }
      
      const result = await runIsolatedBenchmark(competitor);
      allResults.push(result);
      
    } catch (error) {
      console.error(`❌ Failed to benchmark ${competitor}: ${error.message}`);
    }
  }
  
  // Generate consolidated results with both internal and external memory measurements
  const consolidatedCSV = ['system,operations,time_ms,ops_per_sec,memory_mb_internal,memory_mb_external,final_length'];
  
  for (const competitorResult of allResults) {
    for (const dataPoint of competitorResult.results) {
      consolidatedCSV.push(
        `${competitorResult.competitorName},${dataPoint.operations},${dataPoint.timeMs},${dataPoint.opsPerSec},${dataPoint.memoryMB_internal},${dataPoint.memoryMB_external},${dataPoint.finalLength}`
      );
    }
  }
  
  const outputPath = path.join(__dirname, '../data/reliable_competitors_comparison.csv');
  fs.writeFileSync(outputPath, consolidatedCSV.join('\\n'));
  
  console.log(`\\n🎯 Reliable benchmarks completed!`);
  console.log(`📊 Results saved to: ${outputPath}`);
  console.log('\\nComparison: Internal JS vs External System Memory Monitoring');
  
  return allResults;
}

if (require.main === module) {
  runReliableBenchmarks().catch(console.error);
}

module.exports = { runReliableBenchmarks, runIsolatedBenchmark };