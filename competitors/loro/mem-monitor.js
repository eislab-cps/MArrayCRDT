#!/usr/bin/env node
// External memory monitoring for reliable measurement
const { exec } = require('child_process');

const PID = process.argv[2];
if (!PID) {
  console.error('Usage: node mem-monitor.js <PID>');
  process.exit(1);
}

let maxMemoryMB = 0;
let measurements = [];
let monitoring = true;

async function getMemoryUsage(pid) {
  return new Promise((resolve) => {
    exec(`ps -p ${pid} -o rss= 2>/dev/null`, (error, stdout) => {
      if (error || !stdout.trim()) {
        resolve(null); // Process ended
        return;
      }
      
      const rssKB = parseInt(stdout.trim());
      if (!isNaN(rssKB)) {
        const memoryMB = rssKB / 1024;
        maxMemoryMB = Math.max(maxMemoryMB, memoryMB);
        measurements.push({ timestamp: Date.now(), memoryMB });
        resolve(memoryMB);
      } else {
        resolve(null);
      }
    });
  });
}

// Monitor every 50ms for better accuracy
async function monitorLoop() {
  while (monitoring) {
    const mem = await getMemoryUsage(PID);
    if (mem === null) {
      // Process ended
      break;
    }
    await new Promise(resolve => setTimeout(resolve, 50));
  }
  
  // Output final results
  console.log(JSON.stringify({
    maxMemoryMB: Math.round(maxMemoryMB * 100) / 100,
    measurements: measurements.length,
    avgMemoryMB: measurements.length > 0 ? 
      Math.round((measurements.reduce((sum, m) => sum + m.memoryMB, 0) / measurements.length) * 100) / 100 : 0
  }));
}

// Start monitoring
monitorLoop();

// Handle process end signals
process.on('SIGTERM', () => {
  monitoring = false;
});

process.on('SIGINT', () => {
  monitoring = false;
});