#!/usr/bin/env node
// Archive the latest benchmark results for web UI

const fs = require('fs');
const path = require('path');

// Create latest version directory for web UI  
function archiveLatestResults() {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
  const versionDir = path.join(__dirname, 'data', 'benchmark_runs', `isolated-${timestamp}`);
  
  if (!fs.existsSync(path.dirname(versionDir))) {
    fs.mkdirSync(path.dirname(versionDir), { recursive: true });
  }
  fs.mkdirSync(versionDir);
  
  console.log(`📂 Archiving to: isolated-${timestamp}`);
  
  const resultFiles = [
    // Individual competitor results  
    { src: 'competitors/automerge/automerge_results.csv', dst: 'automerge_results.csv' },
    { src: 'competitors/yjs/yjs_results.csv', dst: 'yjs_results.csv' },
    { src: 'competitors/loro/loro_results.csv', dst: 'loro_results.csv' },
    { src: 'competitors/loro/loro_array_results.csv', dst: 'loro_array_results.csv' },
    { src: 'competitors/baseline/baseline_results.csv', dst: 'baseline_results.csv' },
    // MArrayCRDT results (from previous run)
    { src: 'simulation/marraycrdt_results.csv', dst: 'marraycrdt_results.csv' }
  ];
  
  let copiedFiles = 0;
  
  for (const { src, dst } of resultFiles) {
    const srcPath = path.join(__dirname, src);
    const dstPath = path.join(versionDir, dst);
    
    if (fs.existsSync(srcPath)) {
      fs.copyFileSync(srcPath, dstPath);
      copiedFiles++;
      console.log(`  ✅ Copied: ${src}`);
    } else {
      console.log(`  ⚠️  Missing: ${src}`);
    }
  }
  
  // Create consolidated competitors CSV 
  const competitorFiles = [
    { name: 'Automerge', file: path.join(versionDir, 'automerge_results.csv') },
    { name: 'Yjs', file: path.join(versionDir, 'yjs_results.csv') },
    { name: 'Loro', file: path.join(versionDir, 'loro_results.csv') },
    { name: 'LoroArray', file: path.join(versionDir, 'loro_array_results.csv') },
    { name: 'Baseline', file: path.join(versionDir, 'baseline_results.csv') }
  ];
  
  const consolidatedRows = ['system,operations,time_ms,ops_per_sec,memory_mb,final_length'];
  
  for (const { name, file } of competitorFiles) {
    if (fs.existsSync(file)) {
      const content = fs.readFileSync(file, 'utf8');
      const lines = content.trim().split('\n').slice(1); // Skip header
      for (const line of lines) {
        if (line.includes(',') && !line.includes('system,operations') && line.length > 10) {
          const parts = line.split(',');
          if (parts.length >= 5) {
            // Replace system name for consistency
            parts[0] = name;
            consolidatedRows.push(parts.join(','));
          }
        }
      }
    }
  }
  
  fs.writeFileSync(path.join(versionDir, 'competitors_comparison.csv'), consolidatedRows.join('\n'));
  console.log(`  📊 Created: competitors_comparison.csv (${consolidatedRows.length-1} data points)`);
  
  // Update versions list
  const versionsPath = path.join(__dirname, 'data', 'available_versions.json');
  let versions = [];
  if (fs.existsSync(versionsPath)) {
    try {
      versions = JSON.parse(fs.readFileSync(versionsPath, 'utf8'));
    } catch (error) {
      versions = [];
    }
  }
  
  versions.unshift({
    version: `isolated-${timestamp}`,
    path: path.relative(__dirname, versionDir),
    created: new Date().toISOString(),
    method: 'process_isolated',
    fileCount: copiedFiles,
    description: 'Process-isolated benchmarks with reliable measurements'
  });
  
  versions = versions.slice(0, 10);
  fs.writeFileSync(versionsPath, JSON.stringify(versions, null, 2));
  
  console.log('\\n🎯 Latest process-isolated results archived!');
  console.log(`📊 Consolidated ${consolidatedRows.length - 1} competitor data points`);  
  console.log('🌐 Results available in web UI version selector');
  
  return versionDir;
}

if (require.main === module) {
  archiveLatestResults();
}