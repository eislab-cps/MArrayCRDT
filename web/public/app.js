// Performance Visualization App
class PerformanceVisualization {
    constructor() {
        this.data = null;
        this.charts = {};
        this.currentVersion = '';
        this.init();
    }
    
    async init() {
        try {
            await this.loadVersions();
            this.setupVersionSelector();
            // Load the first available version instead of latest
            const firstVersion = this.versions.length > 0 ? this.versions[0].version : '';
            await this.loadData(firstVersion);
            this.updateStats();
            this.createCharts();
            this.showContent();
        } catch (error) {
            this.showError(error.message);
        }
    }
    
    async loadVersions() {
        try {
            const response = await fetch('/api/versions');
            if (response.ok) {
                this.versions = await response.json();
            } else {
                this.versions = [];
            }
        } catch (error) {
            console.warn('Failed to load versions:', error);
            this.versions = [];
        }
    }
    
    async loadData(version = '') {
        const url = version ? `/api/performance-data?version=${version}` : '/api/performance-data';
        const response = await fetch(url);
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to load data');
        }
        this.data = await response.json();
        this.currentVersion = version;
        console.log('Loaded performance data:', this.data);
    }
    
    updateStats() {
        const { MArrayCRDT, Automerge, Yjs, Loro, LoroArray, Baseline } = this.data;
        
        // Calculate stats from all available systems
        const allSystems = [MArrayCRDT, Automerge, Yjs || [], Loro || [], LoroArray || [], Baseline];
        const maxOperations = Math.max(
            ...allSystems.flatMap(system => system.map(d => d.operations))
        );
        
        const marraycrdtBest = MArrayCRDT.length > 0 ? Math.max(...MArrayCRDT.map(d => d.opsPerSec)) : 0;
        const automergeBest = Automerge.length > 0 ? Math.max(...Automerge.map(d => d.opsPerSec)) : 0;
        const yjsBest = Yjs && Yjs.length > 0 ? Math.max(...Yjs.map(d => d.opsPerSec)) : 0;
        const loroBest = Loro && Loro.length > 0 ? Math.max(...Loro.map(d => d.opsPerSec)) : 0;
        const loroArrayBest = LoroArray && LoroArray.length > 0 ? Math.max(...LoroArray.map(d => d.opsPerSec)) : 0;
        const baselineOps = Baseline.length > 0 ? Baseline[0].opsPerSec : 0;
        
        // Update DOM
        document.getElementById('total-operations').textContent = maxOperations.toLocaleString();
        document.getElementById('marraycrdt-best').textContent = Math.round(marraycrdtBest).toLocaleString();
        document.getElementById('automerge-best').textContent = Math.round(automergeBest).toLocaleString();
        document.getElementById('yjs-best').textContent = Math.round(yjsBest).toLocaleString();
        document.getElementById('loro-best').textContent = Math.round(loroBest).toLocaleString();
        document.getElementById('loro-array-best').textContent = Math.round(loroArrayBest).toLocaleString();
        document.getElementById('baseline-ops').textContent = Math.round(baselineOps).toLocaleString();
    }
    
    createCharts() {
        this.createThroughputChart();
        this.createMemoryChart();
    }
    
    createThroughputChart() {
        const ctx = document.getElementById('throughputChart').getContext('2d');
        const { MArrayCRDT, Automerge, Yjs, Loro, LoroArray, Baseline } = this.data;
        
        // Color scheme for all competitors
        const colors = {
            MArrayCRDT: { border: '#e74c3c', bg: 'rgba(231, 76, 60, 0.1)' },
            Automerge: { border: '#3498db', bg: 'rgba(52, 152, 219, 0.1)' },
            Yjs: { border: '#f39c12', bg: 'rgba(243, 156, 18, 0.1)' },
            Loro: { border: '#9b59b6', bg: 'rgba(155, 89, 182, 0.1)' },
            LoroArray: { border: '#e91e63', bg: 'rgba(233, 30, 99, 0.1)' }, // Changed to pink
            Baseline: { border: '#2ecc71', bg: 'rgba(46, 204, 113, 0.1)' }
        };
        
        // Prepare datasets
        const datasets = [];
        
        const systems = { MArrayCRDT, Automerge, Yjs, Loro, LoroArray };
        
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
        const { MArrayCRDT, Automerge, Yjs, Loro, LoroArray, Baseline } = this.data;
        
        // Color scheme for memory chart (reuse from throughput)
        const colors = {
            MArrayCRDT: { border: '#e74c3c', bg: 'rgba(231, 76, 60, 0.2)' },
            Automerge: { border: '#3498db', bg: 'rgba(52, 152, 219, 0.2)' },
            Yjs: { border: '#f39c12', bg: 'rgba(243, 156, 18, 0.2)' },
            Loro: { border: '#9b59b6', bg: 'rgba(155, 89, 182, 0.2)' },
            LoroArray: { border: '#e91e63', bg: 'rgba(233, 30, 99, 0.2)' },
            Baseline: { border: '#2ecc71', bg: 'rgba(46, 204, 113, 0.2)' }
        };
        
        const datasets = [];
        const systems = { MArrayCRDT, Automerge, Yjs, Loro, LoroArray, Baseline };
        
        Object.entries(systems).forEach(([name, data]) => {
            if (data && data.length > 0) {
                datasets.push({
                    label: name,
                    data: data.map(d => ({x: d.operations, y: d.memoryMb})),
                    borderColor: colors[name].border,
                    backgroundColor: colors[name].bg,
                    borderWidth: 3,
                    fill: true,
                    tension: 0.2
                });
            }
        });
        
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
    
    
    showContent() {
        document.getElementById('loading').style.display = 'none';
        document.getElementById('content').style.display = 'block';
    }
    
    setupVersionSelector() {
        const select = document.getElementById('version-select');
        
        // Clear all existing options
        select.innerHTML = '';
        
        // Add version options (most recent first, with clearer timestamps)
        for (const version of this.versions) {
            const option = document.createElement('option');
            option.value = version.version;
            
            // Format timestamp more clearly
            const date = new Date(version.created);
            const timeStr = date.toLocaleString('en-US', {
                month: 'short',
                day: 'numeric', 
                hour: '2-digit',
                minute: '2-digit'
            });
            
            // Create user-friendly name
            const isIsolated = version.version.startsWith('isolated-');
            const typeLabel = isIsolated ? 'Isolated' : 'Standard';
            option.textContent = `${timeStr} - ${typeLabel} Benchmark`;
            select.appendChild(option);
        }
        
        // Set current selection to first version
        if (this.versions.length > 0) {
            select.value = this.currentVersion || this.versions[0].version;
        }
        
        // Handle version changes
        select.addEventListener('change', async (e) => {
            const selectedVersion = e.target.value;
            document.getElementById('loading').style.display = 'block';
            document.getElementById('content').style.display = 'none';
            document.getElementById('error').style.display = 'none';
            
            try {
                await this.loadData(selectedVersion);
                this.updateStats();
                this.destroyCharts();
                this.createCharts();
                this.showContent();
            } catch (error) {
                this.showError(error.message);
            }
        });
    }
    
    destroyCharts() {
        Object.values(this.charts).forEach(chart => {
            if (chart && typeof chart.destroy === 'function') {
                chart.destroy();
            }
        });
        this.charts = {};
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