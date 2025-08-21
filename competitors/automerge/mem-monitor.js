#!/usr/bin/env node
// Memory monitoring for Automerge competitor
const fs = require('fs');
const { spawn } = require('child_process');

const PID = process.argv[2];
if (!PID) {
  console.error('Usage: node mem-monitor.js <PID>');
  process.exit(1);
}

let maxMemoryMB = 0;
let measurements = [];

function getMemoryUsage(pid) {
  try {
    const result = spawn('ps', ['-p', pid, '-o', 'rss='], { stdio: 'pipe' });
    let output = '';
    
    result.stdout.on('data', (data) => {
      output += data.toString();
    });
    
    result.on('close', (code) => {
      if (code === 0 && output.trim()) {
        const rssKB = parseInt(output.trim());
        const memoryMB = rssKB / 1024;
        maxMemoryMB = Math.max(maxMemoryMB, memoryMB);
        measurements.push({ timestamp: Date.now(), memoryMB });
      }
    });
  } catch (error) {
    // Process might have ended
  }
}

// Monitor every 100ms
const interval = setInterval(() => {
  getMemoryUsage(PID);
}, 100);

// Stop monitoring when main process ends
process.stdin.on('data', () => {
  clearInterval(interval);
  console.log(JSON.stringify({
    maxMemoryMB,
    measurements: measurements.length,
    avgMemoryMB: measurements.length > 0 ? 
      measurements.reduce((sum, m) => sum + m.memoryMB, 0) / measurements.length : 0
  }));
  process.exit(0);
});

// Handle process end
process.on('SIGTERM', () => {
  clearInterval(interval);
  process.exit(0);
});