// Master script to run all competitor CRDT benchmarks
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const COMPETITORS = ['automerge', 'yjs', 'loro', 'baseline'];

async function installDependencies() {
  console.log('🔧 Installing dependencies for all competitors...\n');
  
  for (const competitor of COMPETITORS) {
    console.log(`Installing ${competitor} dependencies...`);
    try {
      execSync('npm install', { 
        cwd: path.join(__dirname, competitor),
        stdio: 'inherit'
      });
    } catch (error) {
      console.error(`❌ Failed to install ${competitor} dependencies:`, error.message);
      process.exit(1);
    }
  }
  
  console.log('\n✅ All dependencies installed!\n');
}

async function runBenchmarks() {
  console.log('🚀 Running benchmarks for all competitors...\n');
  
  const allResults = [];
  
  for (const competitor of COMPETITORS) {
    console.log(`\n=== Running ${competitor.toUpperCase()} benchmark ===`);
    
    try {
      // Run standard benchmark
      const output = execSync('node --expose-gc simulation.js', {
        cwd: path.join(__dirname, competitor),
        encoding: 'utf8'
      });
      
      console.log(output);
      
      // Read the results CSV
      const resultsPath = path.join(__dirname, competitor, `${competitor}_results.csv`);
      if (fs.existsSync(resultsPath)) {
        const csvContent = fs.readFileSync(resultsPath, 'utf8');
        const lines = csvContent.trim().split('\n').slice(1); // Skip header
        allResults.push(...lines);
      }
      
      // Special case: Loro also has array benchmark
      if (competitor === 'loro') {
        console.log(`\n=== Running LORO ARRAY benchmark (MovableList) ===`);
        try {
          const arrayOutput = execSync('node --expose-gc array_simulation.js', {
            cwd: path.join(__dirname, competitor),
            encoding: 'utf8'
          });
          
          console.log(arrayOutput);
          
          // Read the array results CSV
          const arrayResultsPath = path.join(__dirname, competitor, 'loro_array_results.csv');
          if (fs.existsSync(arrayResultsPath)) {
            const csvContent = fs.readFileSync(arrayResultsPath, 'utf8');
            const lines = csvContent.trim().split('\n').slice(1); // Skip header
            allResults.push(...lines);
          }
        } catch (arrayError) {
          console.error(`❌ Loro array benchmark failed:`, arrayError.message);
        }
      }
      
    } catch (error) {
      console.error(`❌ ${competitor} benchmark failed:`, error.message);
      // Continue with other benchmarks
    }
  }
  
  return allResults;
}

async function consolidateResults(results) {
  const header = 'system,operations,time_ms,ops_per_sec,memory_mb,final_length';
  const consolidated = [header, ...results].join('\n');
  
  const outputPath = path.join(__dirname, '../data/competitors_comparison.csv');
  fs.writeFileSync(outputPath, consolidated);
  
  console.log(`\n📊 Consolidated results saved to: ${outputPath}`);
}

async function generateSummary(results) {
  console.log('\n=== BENCHMARK SUMMARY ===');
  
  const systems = {};
  
  for (const line of results) {
    const [system, operations, timeMs, opsPerSec, memoryMB, finalLength] = line.split(',');
    
    if (!systems[system]) {
      systems[system] = [];
    }
    
    systems[system].push({
      operations: parseInt(operations),
      opsPerSec: parseFloat(opsPerSec),
      memoryMB: parseFloat(memoryMB)
    });
  }
  
  console.log('\nPerformance at 50k operations:');
  console.log('System     | Ops/sec | Memory (MB) | Memory Efficiency');
  console.log('-----------|---------|-------------|------------------');
  
  for (const [system, data] of Object.entries(systems)) {
    const maxScale = data.find(d => d.operations === 50000);
    if (maxScale) {
      const efficiency = (50000 / maxScale.memoryMB).toFixed(0);
      console.log(`${system.padEnd(10)} | ${Math.round(maxScale.opsPerSec).toString().padStart(7)} | ${maxScale.memoryMB.toString().padStart(11)} | ${efficiency} ops/MB`);
    }
  }
}

// Main execution
async function main() {
  try {
    await installDependencies();
    const results = await runBenchmarks();
    await consolidateResults(results);
    await generateSummary(results);
    
    console.log('\n🎯 All competitor benchmarks completed!');
    console.log('\nNext steps:');
    console.log('1. Run MArrayCRDT benchmark: cd ../benchmarks && go run .');
    console.log('2. View results in web UI: cd ../web && npm start');
    
  } catch (error) {
    console.error('❌ Benchmark suite failed:', error);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}