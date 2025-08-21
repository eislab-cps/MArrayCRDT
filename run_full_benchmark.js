#!/usr/bin/env node
// Unified benchmark runner with versioning system
// Runs both MArrayCRDT and all competitor benchmarks, then archives results with timestamps

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

// Create timestamped version directory
function createVersionDir() {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
  const versionDir = path.join(__dirname, 'data', 'benchmark_runs', timestamp);
  
  if (!fs.existsSync(path.dirname(versionDir))) {
    fs.mkdirSync(path.dirname(versionDir), { recursive: true });
  }
  fs.mkdirSync(versionDir);
  
  console.log(`📂 Created benchmark version: ${timestamp}`);
  return { versionDir, timestamp };
}

// Copy results to versioned directory
function archiveResults(versionDir, timestamp) {
  const resultFiles = [
    // MArrayCRDT results
    { src: 'simulation/marraycrdt_results.csv', dst: 'marraycrdt_results.csv' },
    { src: 'simulation/marraycrdt_comprehensive_benchmark.json', dst: 'marraycrdt_detailed.json' },
    
    // Competitor consolidated results  
    { src: 'data/competitors_comparison.csv', dst: 'competitors_comparison.csv' },
    
    // Individual competitor results
    { src: 'competitors/automerge/automerge_results.csv', dst: 'automerge_results.csv' },
    { src: 'competitors/yjs/yjs_results.csv', dst: 'yjs_results.csv' },
    { src: 'competitors/loro/loro_results.csv', dst: 'loro_results.csv' },
    { src: 'competitors/loro/loro_array_results.csv', dst: 'loro_array_results.csv' },
    { src: 'competitors/baseline/baseline_results.csv', dst: 'baseline_results.csv' }
  ];
  
  let copiedFiles = 0;
  const manifest = {
    version: timestamp,
    created: new Date().toISOString(),
    files: [],
    summary: {}
  };
  
  for (const { src, dst } of resultFiles) {
    const srcPath = path.join(__dirname, src);
    const dstPath = path.join(versionDir, dst);
    
    if (fs.existsSync(srcPath)) {
      try {
        fs.copyFileSync(srcPath, dstPath);
        const stats = fs.statSync(dstPath);
        manifest.files.push({
          name: dst,
          size: stats.size,
          originalPath: src
        });
        copiedFiles++;
        console.log(`  ✅ Archived: ${src} -> ${dst}`);
      } catch (error) {
        console.log(`  ⚠️  Failed to copy ${src}: ${error.message}`);
      }
    } else {
      console.log(`  ⚠️  Missing: ${src}`);
    }
  }
  
  // Generate summary from results
  try {
    if (fs.existsSync(path.join(versionDir, 'competitors_comparison.csv'))) {
      const csvContent = fs.readFileSync(path.join(versionDir, 'competitors_comparison.csv'), 'utf8');
      const lines = csvContent.trim().split('\\n').slice(1); // Skip header
      
      const systems = {};
      for (const line of lines) {
        const [system, ops, timeMs, opsPerSec, memMb] = line.split(',');
        if (!systems[system]) systems[system] = [];
        systems[system].push({
          operations: parseInt(ops),
          opsPerSec: parseFloat(opsPerSec),
          memoryMb: parseFloat(memMb)
        });
      }
      
      // Calculate peak performance for each system
      for (const [system, data] of Object.entries(systems)) {
        const maxOps = Math.max(...data.map(d => d.opsPerSec));
        const at50k = data.find(d => d.operations === 50000);
        manifest.summary[system] = {
          peakOpsPerSec: Math.round(maxOps),
          opsPerSecAt50k: at50k ? Math.round(at50k.opsPerSec) : null,
          memoryAt50k: at50k ? at50k.memoryMb : null
        };
      }
    }
  } catch (error) {
    console.log(`  ⚠️  Failed to generate summary: ${error.message}`);
  }
  
  // Save manifest
  fs.writeFileSync(path.join(versionDir, 'manifest.json'), JSON.stringify(manifest, null, 2));
  console.log(`  📋 Generated manifest with ${copiedFiles} files`);
  
  return manifest;
}

// Update available versions list
function updateVersionsList(versionDir, timestamp, manifest) {
  const versionsPath = path.join(__dirname, 'data', 'available_versions.json');
  
  let versions = [];
  if (fs.existsSync(versionsPath)) {
    try {
      versions = JSON.parse(fs.readFileSync(versionsPath, 'utf8'));
    } catch (error) {
      console.log('Creating new versions list...');
    }
  }
  
  // Add new version at the beginning (most recent first)
  versions.unshift({
    version: timestamp,
    path: path.relative(__dirname, versionDir),
    created: new Date().toISOString(),
    fileCount: manifest.files.length,
    summary: manifest.summary
  });
  
  // Keep only last 10 versions in the list
  versions = versions.slice(0, 10);
  
  fs.writeFileSync(versionsPath, JSON.stringify(versions, null, 2));
  console.log(`📝 Updated versions list (${versions.length} versions available)`);
}

// Main benchmark execution
async function runFullBenchmark() {
  console.log('🚀 Starting Full CRDT Benchmark Suite');
  console.log('=====================================\\n');
  
  const { versionDir, timestamp } = createVersionDir();
  
  try {
    // Step 1: Run competitor benchmarks
    console.log('📊 Step 1: Running competitor benchmarks...');
    execSync('node run_all.js', { 
      cwd: path.join(__dirname, 'competitors'),
      stdio: 'inherit'
    });
    
    console.log('\\n⏱️  Step 2: Running MArrayCRDT benchmark...');
    execSync('go run .', {
      cwd: path.join(__dirname, 'benchmarks'),
      stdio: 'inherit'
    });
    
    console.log('\\n📦 Step 3: Archiving results...');
    const manifest = archiveResults(versionDir, timestamp);
    
    console.log('\\n📚 Step 4: Updating versions registry...');
    updateVersionsList(versionDir, timestamp, manifest);
    
    console.log('\\n🎯 Full benchmark suite completed successfully!');
    console.log(`\\n📂 Results archived in: ${path.relative(__dirname, versionDir)}`);
    console.log(`\\n🌐 Start web server to view results:`);
    console.log(`   cd web && npm start`);
    
    // Print performance summary
    if (manifest.summary && Object.keys(manifest.summary).length > 0) {
      console.log('\\n🏆 Performance Summary (Peak ops/sec):');
      console.log('----------------------------------------');
      
      const sorted = Object.entries(manifest.summary)
        .sort(([,a], [,b]) => b.peakOpsPerSec - a.peakOpsPerSec);
      
      for (const [system, data] of sorted) {
        const peak = data.peakOpsPerSec?.toLocaleString() || 'N/A';
        const at50k = data.opsPerSecAt50k?.toLocaleString() || 'N/A';
        const memory = data.memoryAt50k !== null ? `${data.memoryAt50k} MB` : 'N/A';
        console.log(`${system.padEnd(12)} | ${peak.padStart(10)} | ${at50k.padStart(10)} @ 50k | ${memory}`);
      }
    }
    
    return versionDir;
    
  } catch (error) {
    console.error('❌ Benchmark suite failed:', error.message);
    
    // Still try to archive partial results
    try {
      console.log('🔄 Attempting to archive partial results...');
      const manifest = archiveResults(versionDir, timestamp);
      updateVersionsList(versionDir, timestamp, manifest);
      console.log('📦 Partial results archived');
    } catch (archiveError) {
      console.error('Failed to archive partial results:', archiveError.message);
    }
    
    process.exit(1);
  }
}

// Run if called directly
if (require.main === module) {
  runFullBenchmark().catch(console.error);
}

module.exports = { runFullBenchmark };