#!/usr/bin/env node
// Simplified process isolation for reliable CRDT benchmarks
// Focus on process isolation for consistency, use existing memory estimation

const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

// Run single competitor in isolated process
async function runIsolatedCompetitor(competitorName, competitorScript) {
  console.log(`\\n${'='.repeat(50)}`);
  console.log(`🔬 Running ${competitorName} in isolated process...`);
  
  return new Promise((resolve, reject) => {
    const startTime = Date.now();
    
    // Spawn completely isolated Node.js process
    const childProcess = spawn('node', [
      '--expose-gc',
      '--max-old-space-size=4096',
      '--no-warnings',
      competitorScript
    ], {
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: path.dirname(competitorScript),
      env: { ...process.env, NODE_ENV: 'benchmark' }
    });
    
    let output = '';
    let errorOutput = '';
    
    childProcess.stdout.on('data', (data) => {
      const chunk = data.toString();
      output += chunk;
      if (chunk.includes('Progress:')) {
        process.stdout.write('.');
      }
    });
    
    childProcess.stderr.on('data', (data) => {
      errorOutput += data.toString();
    });
    
    childProcess.on('exit', (code) => {
      const totalTime = Date.now() - startTime;
      console.log(`\\n⏱️  ${competitorName} completed in ${(totalTime/1000).toFixed(1)}s`);
      
      if (code !== 0) {
        console.error(`❌ ${competitorName} failed with code ${code}`);
        if (errorOutput) console.error(`Error: ${errorOutput}`);
        reject(new Error(`${competitorName} benchmark failed`));
        return;
      }
      
      try {
        const results = parseResults(output, competitorName);
        console.log(`✅ ${competitorName}: ${results.length} data points`);
        resolve(results);
      } catch (error) {
        console.error(`❌ Failed to parse ${competitorName} results:`, error.message);
        reject(error);
      }
    });
    
    // Timeout after 5 minutes
    setTimeout(() => {
      if (!childProcess.killed) {
        console.log(`⏰ ${competitorName} timeout - terminating`);
        childProcess.kill('SIGKILL');
        reject(new Error('Timeout'));
      }
    }, 300000);
  });
}

// Parse CSV results from competitor output
function parseResults(output, systemName) {
  // Find CSV data between header and success message
  const lines = output.split('\\n');
  const results = [];
  let foundHeader = false;
  
  for (const line of lines) {
    if (line.includes('Operations,Time_ms,Ops_per_sec,Memory_MB,Final_Length')) {
      foundHeader = true;
      continue;
    }
    
    if (foundHeader && line.includes(',')) {
      const parts = line.trim().split(',');
      if (parts.length >= 5 && !isNaN(parseInt(parts[0]))) {
        results.push({
          system: systemName,
          operations: parseInt(parts[0]),
          time_ms: parseFloat(parts[1]),
          ops_per_sec: parseFloat(parts[2]),
          memory_mb: parseFloat(parts[3]),
          final_length: parseInt(parts[4])
        });
      }
    }
    
    // Stop parsing after success message
    if (line.includes('✅') || line.includes('🎯')) {
      break;
    }
  }
  
  return results;
}

// Main execution
async function main() {
  console.log('🚀 ISOLATED CRDT BENCHMARK SUITE');
  console.log('Each competitor runs in a separate Node.js process');
  console.log('This ensures consistent memory and timing measurements\\n');
  
  const competitors = [
    { name: 'Baseline', script: 'competitors/baseline/simulation.js' },
    { name: 'Automerge', script: 'competitors/automerge/simulation.js' },
    { name: 'Yjs', script: 'competitors/yjs/simulation.js' },
    { name: 'Loro', script: 'competitors/loro/simulation.js' },
    { name: 'LoroArray', script: 'competitors/loro/array_simulation.js' }
  ];
  
  const allResults = [];
  
  for (let i = 0; i < competitors.length; i++) {
    const { name, script } = competitors[i];
    const scriptPath = path.join(__dirname, script);
    
    if (!fs.existsSync(scriptPath)) {
      console.log(`⚠️  Skipping ${name} - script not found: ${script}`);
      continue;
    }
    
    try {
      // Memory cleanup between benchmarks
      if (i > 0) {
        console.log('\\n⏳ Waiting 3 seconds for system cleanup...');
        await new Promise(resolve => setTimeout(resolve, 3000));
      }
      
      const results = await runIsolatedCompetitor(name, scriptPath);
      allResults.push(...results);
      
    } catch (error) {
      console.error(`❌ Failed ${name}: ${error.message}`);
      // Continue with other competitors
    }
  }
  
  if (allResults.length === 0) {
    console.error('❌ No benchmarks completed successfully');
    process.exit(1);
  }
  
  // Create consolidated CSV
  const csvHeader = 'system,operations,time_ms,ops_per_sec,memory_mb,final_length';
  const csvRows = allResults.map(r => 
    `${r.system},${r.operations},${r.time_ms},${r.ops_per_sec},${r.memory_mb},${r.final_length}`
  );
  const csvContent = [csvHeader, ...csvRows].join('\\n');
  
  const outputPath = path.join(__dirname, 'data/isolated_competitors_comparison.csv');
  fs.writeFileSync(outputPath, csvContent);
  
  console.log(`\\n📊 Results saved to: ${outputPath}`);
  
  // Summary
  console.log('\\n' + '='.repeat(60));
  console.log('🎯 BENCHMARK SUMMARY');
  console.log('='.repeat(60));
  
  const systemStats = {};
  for (const result of allResults) {
    if (!systemStats[result.system]) systemStats[result.system] = [];
    systemStats[result.system].push(result);
  }
  
  for (const [system, results] of Object.entries(systemStats)) {
    const maxOps = Math.max(...results.map(r => r.ops_per_sec));
    const lastResult = results[results.length - 1];
    console.log(`\\n🏆 ${system}`);
    console.log(`   Peak: ${Math.round(maxOps).toLocaleString()} ops/sec`);
    console.log(`   At 50k ops: ${Math.round(lastResult.ops_per_sec).toLocaleString()} ops/sec, ${lastResult.memory_mb.toFixed(2)} MB`);
  }
  
  console.log('\\n✅ All isolated benchmarks completed successfully!');
  console.log('\\n🌐 To view results:');
  console.log('   1. Update web server to use isolated_competitors_comparison.csv');
  console.log('   2. cd web && npm start');
}

if (require.main === module) {
  main().catch(error => {
    console.error('❌ Benchmark suite failed:', error);
    process.exit(1);
  });
}

module.exports = { runIsolatedCompetitor, main };