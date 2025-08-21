package marraycrdt

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GeneratePerformanceGraphs creates various performance visualizations
func GeneratePerformanceGraphs(metricsFile string) error {
	// Load metrics from file
	data, err := os.ReadFile(metricsFile)
	if err != nil {
		return fmt.Errorf("failed to read metrics file: %v", err)
	}

	var metrics PerformanceMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("failed to parse metrics: %v", err)
	}

	fmt.Printf("=== Performance Visualization Generator ===\n")
	fmt.Printf("Loaded metrics from: %s\n", metricsFile)
	fmt.Printf("Total operations: %d\n", metrics.TotalOperations)
	fmt.Printf("Progressive samples: %d\n", len(metrics.ProgressiveMetrics))

	// Generate different types of graphs
	if err := generateThroughputGraph(metrics); err != nil {
		return fmt.Errorf("failed to generate throughput graph: %v", err)
	}

	if err := generateMemoryGraph(metrics); err != nil {
		return fmt.Errorf("failed to generate memory graph: %v", err)
	}

	if err := generateComparisonReport(metrics); err != nil {
		return fmt.Errorf("failed to generate comparison report: %v", err)
	}

	return nil
}

// generateThroughputGraph creates a simple ASCII graph of throughput over time
func generateThroughputGraph(metrics PerformanceMetrics) error {
	fmt.Printf("\n=== Throughput Over Time ===\n")
	
	if len(metrics.ProgressiveMetrics) == 0 {
		return fmt.Errorf("no progressive metrics available")
	}

	// Find max throughput for scaling
	maxThroughput := 0.0
	for _, pm := range metrics.ProgressiveMetrics {
		if pm.OpsPerSecond > maxThroughput {
			maxThroughput = pm.OpsPerSecond
		}
	}

	// Create ASCII graph
	graphWidth := 60
	fmt.Printf("Operations/Second (max: %.0f)\n", maxThroughput)
	fmt.Printf("Timeline: %d operations\n", metrics.TotalOperations)
	fmt.Printf("%s\n", strings.Repeat("-", graphWidth+10))

	for i, pm := range metrics.ProgressiveMetrics {
		barLength := int((pm.OpsPerSecond / maxThroughput) * float64(graphWidth))
		bar := strings.Repeat("█", barLength)
		percentage := float64(pm.OperationIndex) * 100 / float64(metrics.TotalOperations)
		
		fmt.Printf("%3.0f%% |%-60s| %6.0f ops/sec (%d elements)\n", 
			percentage, bar, pm.OpsPerSecond, pm.DocumentLength)
		
		// Only show every few entries for readability
		if len(metrics.ProgressiveMetrics) > 20 && i%(len(metrics.ProgressiveMetrics)/20) != 0 {
			continue
		}
	}

	// Save throughput data to CSV for external plotting tools
	csvFile := "throughput_data.csv"
	csvData := "operation_index,ops_per_second,document_length,elapsed_ms,insert_count,delete_count\n"
	for _, pm := range metrics.ProgressiveMetrics {
		csvData += fmt.Sprintf("%d,%.2f,%d,%.2f,%d,%d\n",
			pm.OperationIndex, pm.OpsPerSecond, pm.DocumentLength,
			pm.ElapsedTimeMs, pm.InsertCount, pm.DeleteCount)
	}
	
	if err := os.WriteFile(csvFile, []byte(csvData), 0644); err != nil {
		return fmt.Errorf("failed to write CSV file: %v", err)
	}
	
	fmt.Printf("\nThroughput data saved to: %s\n", csvFile)
	return nil
}

// generateMemoryGraph creates a visualization of memory usage growth
func generateMemoryGraph(metrics PerformanceMetrics) error {
	fmt.Printf("\n=== Memory Usage Growth ===\n")
	
	if len(metrics.ProgressiveMetrics) == 0 {
		return fmt.Errorf("no progressive metrics available")
	}

	// Calculate memory usage at each point
	graphWidth := 60
	maxMemoryMB := 0.0
	
	for _, pm := range metrics.ProgressiveMetrics {
		memoryMB := float64(pm.DocumentLength * metrics.MemoryPerElement) / (1024 * 1024)
		if memoryMB > maxMemoryMB {
			maxMemoryMB = memoryMB
		}
	}

	fmt.Printf("Memory Usage in MB (max: %.1f MB)\n", maxMemoryMB)
	fmt.Printf("Memory per element: %d bytes (%.1fx overhead)\n", 
		metrics.MemoryPerElement, metrics.MemoryOverhead)
	fmt.Printf("%s\n", strings.Repeat("-", graphWidth+10))

	for i, pm := range metrics.ProgressiveMetrics {
		memoryMB := float64(pm.DocumentLength * metrics.MemoryPerElement) / (1024 * 1024)
		barLength := int((memoryMB / maxMemoryMB) * float64(graphWidth))
		bar := strings.Repeat("█", barLength)
		percentage := float64(pm.OperationIndex) * 100 / float64(metrics.TotalOperations)
		
		fmt.Printf("%3.0f%% |%-60s| %6.1f MB (%d elements)\n", 
			percentage, bar, memoryMB, pm.DocumentLength)
		
		// Only show every few entries for readability
		if len(metrics.ProgressiveMetrics) > 20 && i%(len(metrics.ProgressiveMetrics)/20) != 0 {
			continue
		}
	}

	// Save memory data to CSV
	csvFile := "memory_data.csv"
	csvData := "operation_index,memory_mb,document_length,memory_per_element\n"
	for _, pm := range metrics.ProgressiveMetrics {
		memoryMB := float64(pm.DocumentLength * metrics.MemoryPerElement) / (1024 * 1024)
		csvData += fmt.Sprintf("%d,%.2f,%d,%d\n",
			pm.OperationIndex, memoryMB, pm.DocumentLength, metrics.MemoryPerElement)
	}
	
	if err := os.WriteFile(csvFile, []byte(csvData), 0644); err != nil {
		return fmt.Errorf("failed to write memory CSV file: %v", err)
	}
	
	fmt.Printf("\nMemory data saved to: %s\n", csvFile)
	return nil
}

// generateComparisonReport creates a comparison with expected Automerge performance
func generateComparisonReport(metrics PerformanceMetrics) error {
	fmt.Printf("\n=== MArrayCRDT vs Automerge Comparison ===\n")

	// Based on Kleppmann's paper, typical automerge performance:
	// - Text editing: ~1000-10000 ops/sec (depending on document size)
	// - Memory: ~100-300 bytes per element for RGA
	// - These are rough estimates from the paper

	automergeEstimatedOpsPerSec := 5000.0 // Conservative estimate
	automergeEstimatedMemoryPerElement := 150 // bytes

	fmt.Printf("Performance Comparison:\n")
	fmt.Printf("  MArrayCRDT throughput: %.0f ops/sec\n", metrics.OperationsPerSecond)
	fmt.Printf("  Automerge estimated:   %.0f ops/sec\n", automergeEstimatedOpsPerSec)
	fmt.Printf("  Performance ratio:     %.1fx %s\n", 
		metrics.OperationsPerSecond/automergeEstimatedOpsPerSec,
		getPerformanceIndicator(metrics.OperationsPerSecond/automergeEstimatedOpsPerSec))

	fmt.Printf("\nMemory Comparison:\n")
	fmt.Printf("  MArrayCRDT memory/element: %d bytes\n", metrics.MemoryPerElement)
	fmt.Printf("  Automerge estimated:       %d bytes\n", automergeEstimatedMemoryPerElement)
	fmt.Printf("  Memory ratio:              %.1fx %s\n", 
		float64(metrics.MemoryPerElement)/float64(automergeEstimatedMemoryPerElement),
		getMemoryIndicator(float64(metrics.MemoryPerElement)/float64(automergeEstimatedMemoryPerElement)))

	fmt.Printf("\nWorkload Analysis:\n")
	fmt.Printf("  Insert operations: %d (%.1f%%)\n", 
		metrics.InsertOperations, 
		float64(metrics.InsertOperations)*100/float64(metrics.TotalOperations))
	fmt.Printf("  Delete operations: %d (%.1f%%)\n", 
		metrics.DeleteOperations,
		float64(metrics.DeleteOperations)*100/float64(metrics.TotalOperations))
	fmt.Printf("  Insert throughput: %.0f inserts/sec\n", metrics.InsertThroughput)
	fmt.Printf("  Delete throughput: %.0f deletes/sec\n", metrics.DeleteThroughput)
	
	fmt.Printf("\nOverhead Analysis:\n")
	fmt.Printf("  Time per operation: %.1f μs\n", metrics.TimePerOperationUs)
	fmt.Printf("  Time per insert:    %.1f μs\n", metrics.AvgTimePerInsertUs)
	fmt.Printf("  Time per delete:    %.1f μs\n", metrics.AvgTimePerDeleteUs)
	fmt.Printf("  Final document:     %d characters\n", metrics.FinalDocumentLength)
	fmt.Printf("  Total time:         %.1f ms\n", metrics.TotalTimeMs)
	fmt.Printf("  Memory overhead:    %.1fx vs raw text\n", metrics.MemoryOverhead)

	// Save comparison report
	report := generateTextReport(metrics, automergeEstimatedOpsPerSec, float64(automergeEstimatedMemoryPerElement))
	if err := os.WriteFile("performance_comparison.txt", []byte(report), 0644); err != nil {
		return fmt.Errorf("failed to write comparison report: %v", err)
	}
	
	fmt.Printf("\nDetailed comparison saved to: performance_comparison.txt\n")
	return nil
}

func getPerformanceIndicator(ratio float64) string {
	if ratio > 1.2 {
		return "(faster ✓)"
	} else if ratio > 0.8 {
		return "(similar ~)"
	} else {
		return "(slower ✗)"
	}
}

func getMemoryIndicator(ratio float64) string {
	if ratio < 0.8 {
		return "(lower ✓)"
	} else if ratio < 1.2 {
		return "(similar ~)"
	} else {
		return "(higher ✗)"
	}
}

func generateTextReport(metrics PerformanceMetrics, automergeOps, automergeMemory float64) string {
	report := fmt.Sprintf(`MArrayCRDT Performance Analysis Report
Generated: %s

=== WORKLOAD CHARACTERISTICS ===
Total Operations: %d
- Insert Operations: %d (%.1f%%)
- Delete Operations: %d (%.1f%%)
- Final Document Length: %d characters
- Execution Time: %.2f seconds

=== PERFORMANCE METRICS ===
Throughput:
- Overall: %.0f operations/second
- Insert: %.0f inserts/second  
- Delete: %.0f deletes/second

Latency:
- Average per operation: %.1f μs
- Average per insert: %.1f μs
- Average per delete: %.1f μs

=== MEMORY USAGE ===
- Memory per element: %d bytes
- Total estimated memory: %.1f MB
- Memory overhead vs raw text: %.1fx

=== COMPARISON WITH AUTOMERGE ===
Performance:
- MArrayCRDT: %.0f ops/sec
- Automerge (estimated): %.0f ops/sec
- Ratio: %.2fx %s

Memory:
- MArrayCRDT: %d bytes/element
- Automerge (estimated): %.0f bytes/element
- Ratio: %.2fx %s

=== OVERHEAD FACTORS ===
MArrayCRDT design adds overhead through:
1. Vector clocks for causality tracking
2. Fractional indexing for position management
3. UUID-based element identification
4. Last-Writer-Wins conflict resolution metadata
5. Element tombstones for deleted items

The overhead is expected for a full-featured CRDT that supports:
- Concurrent editing across multiple replicas
- Deterministic conflict resolution
- Move operations with semantic preservation
- Rich operation history tracking
`,
		metrics.Timestamp.Format("2006-01-02 15:04:05"),
		metrics.TotalOperations,
		metrics.InsertOperations, float64(metrics.InsertOperations)*100/float64(metrics.TotalOperations),
		metrics.DeleteOperations, float64(metrics.DeleteOperations)*100/float64(metrics.TotalOperations),
		metrics.FinalDocumentLength,
		metrics.TotalTimeMs/1000,
		metrics.OperationsPerSecond,
		metrics.InsertThroughput,
		metrics.DeleteThroughput,
		metrics.TimePerOperationUs,
		metrics.AvgTimePerInsertUs,
		metrics.AvgTimePerDeleteUs,
		metrics.MemoryPerElement,
		metrics.EstimatedMemoryMB,
		metrics.MemoryOverhead,
		metrics.OperationsPerSecond,
		automergeOps,
		metrics.OperationsPerSecond/automergeOps,
		getPerformanceIndicator(metrics.OperationsPerSecond/automergeOps),
		metrics.MemoryPerElement,
		automergeMemory,
		float64(metrics.MemoryPerElement)/automergeMemory,
		getMemoryIndicator(float64(metrics.MemoryPerElement)/automergeMemory))

	return report
}

// GenerateVisualizationsFromFile loads metrics and generates all visualizations
func GenerateVisualizationsFromFile(filename string) {
	fmt.Printf("Generating performance visualizations...\n")
	if err := GeneratePerformanceGraphs(filename); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("Visualization generation completed!\n")
}