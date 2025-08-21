// Performance Visualization App
class PerformanceVisualization {
    constructor() {
        this.data = null;
        this.charts = {};
        this.init();
    }
    
    async init() {
        try {
            await this.loadData();
            this.updateStats();
            this.createCharts();
            this.showContent();
        } catch (error) {
            this.showError(error.message);
        }
    }
    
    async loadData() {
        const response = await fetch('/api/performance-data');
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to load data');
        }
        this.data = await response.json();
        console.log('Loaded performance data:', this.data);
    }
    
    updateStats() {
        const { MArrayCRDT, Automerge, Yjs, Loro, Baseline } = this.data;
        
        // Calculate stats from all available systems
        const allSystems = [MArrayCRDT, Automerge, Yjs || [], Loro || [], Baseline];
        const maxOperations = Math.max(
            ...allSystems.flatMap(system => system.map(d => d.operations))
        );
        
        const marraycrdtBest = MArrayCRDT.length > 0 ? Math.max(...MArrayCRDT.map(d => d.opsPerSec)) : 0;
        const automergeBest = Automerge.length > 0 ? Math.max(...Automerge.map(d => d.opsPerSec)) : 0;
        const yjsBest = Yjs && Yjs.length > 0 ? Math.max(...Yjs.map(d => d.opsPerSec)) : 0;
        const loroBest = Loro && Loro.length > 0 ? Math.max(...Loro.map(d => d.opsPerSec)) : 0;
        const baselineOps = Baseline.length > 0 ? Baseline[0].opsPerSec : 0;
        
        // Update DOM
        document.getElementById('total-operations').textContent = maxOperations.toLocaleString();
        document.getElementById('marraycrdt-best').textContent = Math.round(marraycrdtBest).toLocaleString();
        document.getElementById('automerge-best').textContent = Math.round(automergeBest).toLocaleString();
        document.getElementById('baseline-ops').textContent = Math.round(baselineOps).toLocaleString();
    }
    
    createCharts() {
        this.createThroughputChart();
        this.createMemoryChart();
        this.createRatioChart();
    }
    
    createThroughputChart() {
        const ctx = document.getElementById('throughputChart').getContext('2d');
        const { MArrayCRDT, Automerge, Yjs, Loro, Baseline } = this.data;
        
        // Color scheme for all competitors
        const colors = {
            MArrayCRDT: { border: '#e74c3c', bg: 'rgba(231, 76, 60, 0.1)' },
            Automerge: { border: '#3498db', bg: 'rgba(52, 152, 219, 0.1)' },
            Yjs: { border: '#f39c12', bg: 'rgba(243, 156, 18, 0.1)' },
            Loro: { border: '#9b59b6', bg: 'rgba(155, 89, 182, 0.1)' },
            Baseline: { border: '#2ecc71', bg: 'rgba(46, 204, 113, 0.1)' }
        };
        
        // Prepare datasets
        const datasets = [];
        
        const systems = { MArrayCRDT, Automerge, Yjs, Loro };
        
        Object.entries(systems).forEach(([name, data]) => {
            if (data && data.length > 0) {
                datasets.push({
                    label: name,
                    data: data.map(d => ({x: d.operations, y: d.opsPerSec})),
                    borderColor: colors[name].border,
                    backgroundColor: colors[name].bg,
                    borderWidth: 3,
                    fill: false,
                    tension: 0.2
                });
            }
        });
        
        if (Baseline.length > 0) {
            datasets.push({
                label: 'JavaScript Array (Baseline)',
                data: Baseline.map(d => ({x: d.operations, y: d.opsPerSec})),
                borderColor: '#2ecc71',
                backgroundColor: 'rgba(46, 204, 113, 0.1)',
                borderWidth: 2,
                borderDash: [5, 5],
                fill: false,
                pointRadius: 6
            });
        }
        
        this.charts.throughput = new Chart(ctx, {
            type: 'line',
            data: { datasets },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        type: 'linear',
                        position: 'bottom',
                        title: {
                            display: true,
                            text: 'Operations',
                            font: { size: 14, weight: 'bold' }
                        },
                        ticks: {
                            callback: function(value) {
                                return (value / 1000) + 'k';
                            }
                        }
                    },
                    y: {
                        title: {
                            display: true,
                            text: 'Operations per Second',
                            font: { size: 14, weight: 'bold' }
                        },
                        ticks: {
                            callback: function(value) {
                                return value.toLocaleString();
                            }
                        }
                    }
                },
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            usePointStyle: true,
                            padding: 20
                        }
                    },
                    tooltip: {
                        callbacks: {
                            title: function(context) {
                                return context[0].parsed.x.toLocaleString() + ' operations';
                            },
                            label: function(context) {
                                return context.dataset.label + ': ' + 
                                       Math.round(context.parsed.y).toLocaleString() + ' ops/sec';
                            }
                        }
                    }
                }
            }
        });
    }
    
    createMemoryChart() {
        const ctx = document.getElementById('memoryChart').getContext('2d');
        const { MArrayCRDT, Automerge } = this.data;
        
        const datasets = [];
        
        if (MArrayCRDT.length > 0) {
            datasets.push({
                label: 'MArrayCRDT',
                data: MArrayCRDT.map(d => ({x: d.operations, y: d.memoryMb})),
                borderColor: '#e74c3c',
                backgroundColor: 'rgba(231, 76, 60, 0.2)',
                borderWidth: 3,
                fill: true,
                tension: 0.2
            });
        }
        
        if (Automerge.length > 0) {
            datasets.push({
                label: 'Automerge',
                data: Automerge.map(d => ({x: d.operations, y: d.memoryMb})),
                borderColor: '#3498db',
                backgroundColor: 'rgba(52, 152, 219, 0.2)',
                borderWidth: 3,
                fill: true,
                tension: 0.2
            });
        }
        
        this.charts.memory = new Chart(ctx, {
            type: 'line',
            data: { datasets },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        type: 'linear',
                        position: 'bottom',
                        title: {
                            display: true,
                            text: 'Operations',
                            font: { size: 14, weight: 'bold' }
                        },
                        ticks: {
                            callback: function(value) {
                                return (value / 1000) + 'k';
                            }
                        }
                    },
                    y: {
                        title: {
                            display: true,
                            text: 'Memory Usage (MB)',
                            font: { size: 14, weight: 'bold' }
                        },
                        ticks: {
                            callback: function(value) {
                                return value.toFixed(1) + ' MB';
                            }
                        }
                    }
                },
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            usePointStyle: true,
                            padding: 20
                        }
                    },
                    tooltip: {
                        callbacks: {
                            title: function(context) {
                                return context[0].parsed.x.toLocaleString() + ' operations';
                            },
                            label: function(context) {
                                return context.dataset.label + ': ' + 
                                       context.parsed.y.toFixed(1) + ' MB';
                            }
                        }
                    }
                }
            }
        });
    }
    
    createRatioChart() {
        const ctx = document.getElementById('ratioChart').getContext('2d');
        const { MArrayCRDT, Automerge } = this.data;
        
        // Calculate performance ratios
        const ratioData = [];
        MArrayCRDT.forEach(mData => {
            const automergeData = Automerge.find(a => a.operations === mData.operations);
            if (automergeData) {
                const ratio = mData.opsPerSec / automergeData.opsPerSec;
                ratioData.push({
                    x: mData.operations,
                    y: ratio,
                    marraycrdt: mData.opsPerSec,
                    automerge: automergeData.opsPerSec
                });
            }
        });
        
        this.charts.ratio = new Chart(ctx, {
            type: 'bar',
            data: {
                datasets: [{
                    label: 'Performance Ratio (MArrayCRDT / Automerge)',
                    data: ratioData,
                    backgroundColor: ratioData.map(d => d.y >= 1 ? 'rgba(46, 204, 113, 0.8)' : 'rgba(231, 76, 60, 0.8)'),
                    borderColor: ratioData.map(d => d.y >= 1 ? '#27ae60' : '#c0392b'),
                    borderWidth: 2
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        type: 'linear',
                        position: 'bottom',
                        title: {
                            display: true,
                            text: 'Operations',
                            font: { size: 14, weight: 'bold' }
                        },
                        ticks: {
                            callback: function(value) {
                                return (value / 1000) + 'k';
                            }
                        }
                    },
                    y: {
                        title: {
                            display: true,
                            text: 'Ratio (>1 = MArrayCRDT faster)',
                            font: { size: 14, weight: 'bold' }
                        },
                        ticks: {
                            callback: function(value) {
                                return value.toFixed(2) + 'x';
                            }
                        }
                    }
                },
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            usePointStyle: true,
                            padding: 20
                        }
                    },
                    tooltip: {
                        callbacks: {
                            title: function(context) {
                                return context[0].parsed.x.toLocaleString() + ' operations';
                            },
                            label: function(context) {
                                const data = context.raw;
                                return [
                                    `Ratio: ${data.y.toFixed(2)}x`,
                                    `MArrayCRDT: ${Math.round(data.marraycrdt).toLocaleString()} ops/sec`,
                                    `Automerge: ${Math.round(data.automerge).toLocaleString()} ops/sec`
                                ];
                            }
                        }
                    }
                }
            }
        });
    }
    
    showContent() {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('content').style.display = 'block';
    }
    
    showError(message) {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('error').style.display = 'block';
        document.getElementById('error-message').textContent = message;
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new PerformanceVisualization();
});