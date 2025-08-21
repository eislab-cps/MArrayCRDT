const express = require('express');
const path = require('path');
const fs = require('fs');
const csv2json = require('csvtojson');

const app = express();
const port = 3000;

// Serve static files
app.use(express.static('public'));
app.use('/data', express.static('../'));

// API endpoint to get performance data
app.get('/api/performance-data', async (req, res) => {
  try {
    const csvFilePath = path.join(__dirname, '../comprehensive_performance_comparison.csv');
    
    // Check if the comprehensive data exists, otherwise use the smaller comparison
    let dataFile = csvFilePath;
    if (!fs.existsSync(csvFilePath)) {
      dataFile = path.join(__dirname, '../performance_comparison.csv');
    }
    
    if (!fs.existsSync(dataFile)) {
      return res.status(404).json({ 
        error: 'Performance data not found. Please run the benchmark first.' 
      });
    }
    
    const jsonData = await csv2json().fromFile(dataFile);
    
    // Process and group the data
    const systems = {
      MArrayCRDT: [],
      Automerge: [],
      Baseline: []
    };
    
    jsonData.forEach(row => {
      if (systems[row.system]) {
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