#!/usr/bin/env node
// Complete process-isolated benchmark system for reliable CRDT comparison

const { execSync, spawn } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

// Get system memory usage for a process (Linux/macOS compatible)
function getProcessMemory(pid) {
  try {
    if (os.platform() === 'linux') {
      const status = fs.readFileSync(`/proc/${pid}/status`, 'utf8');
      const vmRSSMatch = status.match(/VmRSS:\\s+(\\d+)\\s+kB/);
      return vmRSSMatch ? parseInt(vmRSSMatch[1]) / 1024 : null; // KB to MB
    } else if (os.platform() === 'darwin') {
      const output = execSync(`ps -o rss= -p ${pid}`, {encoding: 'utf8', stdio: ['pipe', 'pipe', 'ignore']});
      return parseInt(output.trim()) / 1024; // KB to MB  
    } else {
      // Windows fallback - less accurate
      const output = execSync(`tasklist /fi "pid eq ${pid}" /fo csv`, {encoding: 'utf8', stdio: ['pipe', 'pipe', 'ignore']});
      const lines = output.split('\\n');
      if (lines.length > 1) {
        const memStr = lines[1].split(',')[4].replace(/"/g, '').replace(/,/g, '');
        return parseInt(memStr) / 1024; // KB to MB
      }
    }
  } catch (error) {
    // Silent fail - return null if monitoring unavailable
    return null;
  }
  return null;
}

// Monitor memory usage throughout process execution
async function monitorProcessMemory(childProcess) {
  const memories = [];
  let peakMemory = 0;
  const startTime = Date.now();
  
  return new Promise((resolve) => {
    const monitor = setInterval(() => {
      if (childProcess.killed || childProcess.exitCode !== null) {
        clearInterval(monitor);
        return;
      }
      
      const memory = getProcessMemory(childProcess.pid);
      if (memory !== null && memory > 0) {
        memories.push({
          timestamp: Date.now() - startTime,
          memoryMB: memory
        });
        peakMemory = Math.max(peakMemory, memory);
      }
    }, 50); // Sample every 50ms for good resolution
    
    childProcess.on('exit', () => {
      clearInterval(monitor);
      
      // Calculate statistics
      const avgMemory = memories.length > 0 ? 
        memories.reduce((sum, m) => sum + m.memoryMB, 0) / memories.length : 0;
      
      resolve({
        peakMemoryMB: peakMemory,
        avgMemoryMB: avgMemory,
        samples: memories.length,
        memoryTrace: memories.slice(-20) // Keep last 20 samples for analysis
      });
    });
  });
}

// Run single competitor in completely isolated process
async function runIsolatedCompetitor(competitorName, competitorScript) {
  console.log(`\\n${'='.repeat(60)}`);
  console.log(`🔬 ISOLATED BENCHMARK: ${competitorName.toUpperCase()}`);
  console.log(`📝 Script: ${competitorScript}`);
  console.log(`🕒 Starting at: ${new Date().toLocaleTimeString()}`);
  
  return new Promise((resolve, reject) => {
    const startTime = Date.now();
    
    // Spawn completely isolated Node.js process
    const childProcess = spawn('node', [
      '--expose-gc',
      '--max-old-space-size=4096', // 4GB - generous limit
      '--no-warnings',
      competitorScript
    ], {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: path.dirname(competitorScript),
      env: { ...process.env, NODE_ENV: 'benchmark' } // Clean environment
    });
    
    let output = '';
    let errorOutput = '';
    
    childProcess.stdout.on('data', (data) => {
      const chunk = data.toString();
      output += chunk;
      // Show real-time progress for long-running benchmarks
      if (chunk.includes('Progress:')) {
        process.stdout.write('.');
      }
    });
    
    childProcess.stderr.on('data', (data) => {
      errorOutput += data.toString();
    });
    
    // Start memory monitoring
    const memoryMonitoring = monitorProcessMemory(childProcess);
    
    childProcess.on('exit', async (code, signal) => {
      const endTime = Date.now();
      const totalTime = endTime - startTime;
      
      console.log(`\\n⏱️  Total benchmark time: ${(totalTime/1000).toFixed(1)}s`);
      
      if (code !== 0) {
        console.error(`❌ Process failed with code ${code} signal ${signal}`);
        if (errorOutput) console.error(`Error output: ${errorOutput}`);
        reject(new Error(`Benchmark failed: ${errorOutput || 'Unknown error'}`));
        return;
      }
      
      const memoryStats = await memoryMonitoring;
      console.log(`📊 Peak Memory: ${memoryStats.peakMemoryMB.toFixed(2)} MB (${memoryStats.samples} samples)`);
      
      try {
        // Parse CSV results from output
        const results = parseCompetitorResults(output, competitorName, memoryStats);
        console.log(`✅ Successfully parsed ${results.length} data points`);
        
        resolve({
          competitor: competitorName,
          results: results,
          memoryStats: memoryStats,
          totalTimeMs: totalTime,
          processOutput: output
        });
        
      } catch (parseError) {
        console.error(`❌ Failed to parse results: ${parseError.message}`);
        console.log('Raw output length:', output.length);
        console.log('First 500 chars:', output.substring(0, 500));
        reject(parseError);
      }
    });
    
    // Timeout after 10 minutes
    setTimeout(() => {
      if (!childProcess.killed) {
        console.log('⏰ Benchmark timeout - terminating process');
        childProcess.kill('SIGKILL');
        reject(new Error('Benchmark timeout'));
      }
    }, 600000);
  });
}

// Parse competitor results from process output
function parseCompetitorResults(output, competitorName, memoryStats) {
  // Look for CSV data in the output - be more flexible with the regex
  const csvMatch = output.match(/Operations,Time_ms,Ops_per_sec,Memory_MB,Final_Length\\s*([\\s\\S]*?)(?=\\n✅|\\n🎯|$)/);
  if (!csvMatch) {
    // Try alternative pattern
    const altMatch = output.match(/Operations,Time_ms,Ops_per_sec,Memory_MB,Final_Length([^]+?)(?=✅|🎯|$)/);
    if (!altMatch) {
      throw new Error('Could not find CSV data in output');
    }
    csvMatch[1] = altMatch[1];
  }
  
  const csvLines = csvMatch[1].trim().split('\\n').filter(line => line.includes(','));
  const results = [];
  
  for (const line of csvLines) {
    const parts = line.split(',');
    if (parts.length >= 5) {
      const [ops, timeMs, opsPerSec, internalMemMB, finalLength] = parts;
      results.push({
        system: competitorName,
        operations: parseInt(ops),
        time_ms: parseFloat(timeMs),
        ops_per_sec: parseFloat(opsPerSec),
        memory_mb_internal: parseFloat(internalMemMB), // JS process.memoryUsage()
        memory_mb_external: memoryStats.peakMemoryMB,   // System monitoring  
        memory_mb_reliable: memoryStats.peakMemoryMB,   // Use external as reliable
        final_length: parseInt(finalLength)
      });
    }
  }
  
  return results;
}

// Create consolidated CSV with reliable memory measurements
function createConsolidatedResults(allResults, outputPath) {
  const csvHeader = 'system,operations,time_ms,ops_per_sec,memory_mb,final_length';
  const csvRows = [];
  
  for (const competitorResult of allResults) {
    for (const dataPoint of competitorResult.results) {
      csvRows.push([
        dataPoint.system,
        dataPoint.operations,
        dataPoint.time_ms,
        dataPoint.ops_per_sec,
        dataPoint.memory_mb_reliable.toFixed(2), // Use reliable external measurement
        dataPoint.final_length
      ].join(','));
    }
  }
  
  const csvContent = [csvHeader, ...csvRows].join('\\n');
  fs.writeFileSync(outputPath, csvContent);
  
  console.log(`\\n📊 Consolidated results saved to: ${outputPath}`);
  return csvContent;
}

// Generate summary report
function generateSummaryReport(allResults) {
  console.log('\\n' + '='.repeat(80));
  console.log('🎯 RELIABLE BENCHMARK SUMMARY');
  console.log('='.repeat(80));
  
  for (const competitorResult of allResults) {
    const results = competitorResult.results;
    const lastResult = results[results.length - 1];
    
    console.log(`\\n🏆 ${competitorResult.competitor.toUpperCase()}`);
    console.log(`   Peak Performance: ${Math.round(Math.max(...results.map(r => r.ops_per_sec))).toLocaleString()} ops/sec`);
    console.log(`   Peak Memory: ${competitorResult.memoryStats.peakMemoryMB.toFixed(2)} MB`);
    console.log(`   At 50k ops: ${Math.round(lastResult.ops_per_sec).toLocaleString()} ops/sec, ${lastResult.memory_mb_reliable.toFixed(2)} MB`);
    console.log(`   Memory Efficiency: ${Math.round(50000 / lastResult.memory_mb_reliable).toLocaleString()} ops/MB`);
    console.log(`   Benchmark Time: ${(competitorResult.totalTimeMs/1000).toFixed(1)}s`);
  }
  
  console.log('\\n' + '='.repeat(80));
  console.log('✅ All benchmarks completed with reliable memory measurement!');
}

// Main execution
async function main() {
  console.log('🚀 RELIABLE CRDT BENCHMARK SUITE');
  console.log('Using Process Isolation + External Memory Monitoring');
  console.log(`🖥️  Platform: ${os.platform()} ${os.arch()}`);
  console.log(`💾 System Memory: ${(os.totalmem() / 1024 / 1024 / 1024).toFixed(1)} GB`);
  console.log(`🕐 Started: ${new Date().toLocaleString()}\\n`);
  
  const competitors = [
    { name: 'Baseline', script: path.join(__dirname, 'competitors/baseline/simulation.js') },
    { name: 'Automerge', script: path.join(__dirname, 'competitors/automerge/simulation.js') },
    // Add others one by one to test
  ];
  
  const allResults = [];
  
  for (let i = 0; i < competitors.length; i++) {
    const { name, script } = competitors[i];
    
    try {
      // Memory cleanup between benchmarks
      if (i > 0) {
        console.log('\\n⏳ Waiting 5 seconds for system memory cleanup...');
        await new Promise(resolve => setTimeout(resolve, 5000));
        
        // Force garbage collection in main process
        if (global.gc) global.gc();
      }
      
      const result = await runIsolatedCompetitor(name, script);
      allResults.push(result);
      
    } catch (error) {
      console.error(`❌ Failed ${name}: ${error.message}`);
      // Continue with other competitors
    }
  }
  
  if (allResults.length === 0) {
    console.error('❌ No benchmarks completed successfully');
    process.exit(1);
  }
  
  // Create consolidated results
  const outputPath = path.join(__dirname, 'data/reliable_competitors_results.csv');
  createConsolidatedResults(allResults, outputPath);
  
  // Generate summary
  generateSummaryReport(allResults);
  
  console.log('\\n🌐 To view results:');
  console.log('   1. Copy reliable results to web server data sources');
  console.log('   2. cd web && npm start');
  console.log('   3. Open http://localhost:3000');
}

if (require.main === module) {
  main().catch(error => {
    console.error('❌ Benchmark suite failed:', error);
    process.exit(1);
  });
}

module.exports = { runIsolatedCompetitor, main };