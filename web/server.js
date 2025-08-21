const express = require('express');
const path = require('path');
const fs = require('fs');
const csv2json = require('csvtojson');

const app = express();
const port = 3000;

// Serve static files
app.use(express.static('public'));
app.use('/data', express.static('../'));

// API endpoint to list available benchmark versions
app.get('/api/versions', (req, res) => {
  const versionsPath = path.join(__dirname, '../data/available_versions.json');
  
  if (!fs.existsSync(versionsPath)) {
    return res.json([]);
  }
  
  try {
    const versions = JSON.parse(fs.readFileSync(versionsPath, 'utf8'));
    res.json(versions);
  } catch (error) {
    console.error('Error reading versions:', error);
    res.status(500).json({ error: 'Failed to load versions' });
  }
});

// API endpoint to get performance data (with optional version)
app.get('/api/performance-data', async (req, res) => {
  try {
    const version = req.query.version;
    let dataSources;
    
    if (version) {
      // Load versioned data
      const versionDir = path.join(__dirname, '../data/benchmark_runs', version);
      if (!fs.existsSync(versionDir)) {
        return res.status(404).json({ error: `Version ${version} not found` });
      }
      
      dataSources = {
        marraycrdt: path.join(versionDir, 'marraycrdt_results.csv'),
        competitors: path.join(versionDir, 'competitors_comparison.csv'),
        fallback: path.join(__dirname, '../data/comprehensive_performance_comparison.csv')
      };
    } else {
      // Use latest data (existing behavior)
      dataSources = {
        marraycrdt: path.join(__dirname, '../simulation/marraycrdt_results.csv'),
        competitors: path.join(__dirname, '../data/competitors_comparison.csv'),
        fallback: path.join(__dirname, '../data/comprehensive_performance_comparison.csv')
      };
    }
    
    let allData = [];
    
    // Load MArrayCRDT simulation results
    if (fs.existsSync(dataSources.marraycrdt)) {
      const marraycrdtData = await csv2json().fromFile(dataSources.marraycrdt);
      allData.push(...marraycrdtData);
      console.log(`Loaded ${marraycrdtData.length} MArrayCRDT results`);
    }
    
    // Load competitor results
    if (fs.existsSync(dataSources.competitors)) {
      const competitorData = await csv2json().fromFile(dataSources.competitors);
      allData.push(...competitorData);
      console.log(`Loaded ${competitorData.length} competitor results`);
    }
    
    // Fallback to existing comprehensive data if no simulation results
    if (allData.length === 0 && fs.existsSync(dataSources.fallback)) {
      const fallbackData = await csv2json().fromFile(dataSources.fallback);
      allData.push(...fallbackData);
      console.log(`Loaded ${fallbackData.length} fallback results`);
    }
    
    if (allData.length === 0) {
      return res.status(404).json({ 
        error: 'No performance data found. Please run benchmarks first.' 
      });
    }
    
    // Process and group the data
    const systems = {
      MArrayCRDT: [],
      Automerge: [],
      Yjs: [],
      Loro: [],
      LoroArray: [],
      Baseline: []
    };
    
    // Process all combined data
    allData.forEach(row => {
      if (systems[row.system] !== undefined) {
        systems[row.system].push({
          operations: parseInt(row.operations),
          timeMs: parseFloat(row.time_ms),
          opsPerSec: parseFloat(row.ops_per_sec),
          memoryMb: parseFloat(row.memory_mb),
          insertOps: parseInt(row.insert_ops) || 0,
          deleteOps: parseInt(row.delete_ops) || 0,
          finalLength: parseInt(row.final_length) || 0
        });
      }
    });
    
    // Sort by operations count
    Object.keys(systems).forEach(system => {
      systems[system].sort((a, b) => a.operations - b.operations);
    });
    
    res.json(systems);
  } catch (error) {
    console.error('Error reading performance data:', error);
    res.status(500).json({ error: 'Failed to load performance data' });
  }
});

app.listen(port, () => {
  console.log(`🚀 MArrayCRDT Performance Visualization Server running at:`);
  console.log(`   http://localhost:${port}`);
  console.log('');
  console.log('📊 Make sure to run the Go benchmark first:');
  console.log('   go run run_comprehensive_comparison.go');
  console.log('');
  console.log('🎯 Then open your browser to view the performance comparison!');
});