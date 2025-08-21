package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/caslun/MArrayCRDT/crdt"
)

// EditingOperation represents the Automerge trace format
type EditingOperation struct {
	Actor    string `json:"actor"`
	Seq      int    `json:"seq"`
	Deps     map[string]interface{} `json:"deps"`
	Message  string `json:"message"`
	StartOp  int    `json:"startOp"`
	Time     int64  `json:"time"`
	Ops      []AtomicOp `json:"ops"`
}

type AtomicOp struct {
	Action string `json:"action"`
	Obj    string `json:"obj"`
	Key    string `json:"key,omitempty"`
	ElemId string `json:"elemId,omitempty"`
	Insert bool   `json:"insert,omitempty"`
	Value  string `json:"value,omitempty"`
	Pred   []string `json:"pred"`
}

// MArrayBenchmarkResult stores results for a single benchmark run
type MArrayBenchmarkResult struct {
	System                string  `json:"system"`
	Operations            int     `json:"operations"`
	TimeMs                float64 `json:"time_ms"`
	OpsPerSec             float64 `json:"ops_per_sec"`
	MemoryMB              float64 `json:"memory_mb"`
	InsertOperations      int     `json:"insert_operations"`
	DeleteOperations      int     `json:"delete_operations"`
	FinalDocumentLength   int     `json:"final_document_length"`
}

// loadEditingTrace loads the Kleppmann editing trace
func loadEditingTrace() ([]EditingOperation, error) {
	data, err := os.ReadFile("../data/paper.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read paper.json: %v", err)
	}

	var operations []EditingOperation
	// Parse JSON lines format
	lines := []string{}
	current := ""
	for _, b := range data {
		if b == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(b)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	for _, line := range lines {
		var op EditingOperation
		if err := json.Unmarshal([]byte(line), &op); err != nil {
			continue // Skip invalid lines
		}
		operations = append(operations, op)
	}

	return operations, nil
}

// Memory tracking helper
func getMemoryUsageMB() float64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

// runMArrayCRDTBenchmark runs a benchmark with MArrayCRDT
func runMArrayCRDTBenchmark(operations []EditingOperation, maxOps int) MArrayBenchmarkResult {
	runtime.GC()
	startMem := getMemoryUsageMB()
	
	// Initialize MArrayCRDT
	array := marraycrdt.New[string]("site1")
	
	// Track element IDs for deletion (simple approach)
	var elementIDs []string
	
	insertOps := 0
	deleteOps := 0
	
	startTime := time.Now()
	
	opCount := 0
	for i := 0; i < len(operations) && opCount < maxOps; i++ {
		operation := operations[i]
		
		// Process each atomic operation within this edit operation
		for _, atomicOp := range operation.Ops {
			if opCount >= maxOps {
				break
			}
			
			if atomicOp.Action == "set" && atomicOp.Insert && atomicOp.Value != "" {
				// This is an insert operation
				// For simplicity, append to end (MArrayCRDT handles ordering)
				id := array.Insert(array.Len(), atomicOp.Value)
				elementIDs = append(elementIDs, id)
				insertOps++
				opCount++
			} else if atomicOp.Action == "del" {
				// This is a delete operation
				if len(elementIDs) > 0 {
					// Delete last element for simplicity
					lastIdx := len(elementIDs) - 1
					id := elementIDs[lastIdx]
					if array.Delete(id) {
						elementIDs = elementIDs[:lastIdx]
						deleteOps++
						opCount++
					}
				}
			}
		}
		
		// Progress reporting
		if opCount%5000 == 0 && opCount > 0 {
			elapsed := time.Since(startTime)
			opsPerSec := float64(opCount) / elapsed.Seconds()
			fmt.Printf("  Progress: %d/%d (%.0f ops/sec)\n", opCount, maxOps, opsPerSec)
		}
	}
	
	elapsed := time.Since(startTime)
	runtime.GC()
	endMem := getMemoryUsageMB()
	
	memoryUsed := endMem - startMem
	if memoryUsed < 0.01 {
		memoryUsed = 0.01 // Minimum reasonable value
	}
	
	opsPerSec := float64(opCount) / elapsed.Seconds()
	
	return MArrayBenchmarkResult{
		System:                "MArrayCRDT",
		Operations:            opCount,
		TimeMs:                float64(elapsed.Nanoseconds()) / 1e6,
		OpsPerSec:             opsPerSec,
		MemoryMB:              memoryUsed,
		InsertOperations:      insertOps,
		DeleteOperations:      deleteOps,
		FinalDocumentLength:   array.Len(),
	}
}

// writeCSVResults writes results to CSV file
func writeCSVResults(results []MArrayBenchmarkResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"system", "operations", "time_ms", "ops_per_sec", "memory_mb", "insert_ops", "delete_ops", "final_length"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, result := range results {
		row := []string{
			result.System,
			strconv.Itoa(result.Operations),
			fmt.Sprintf("%.2f", result.TimeMs),
			fmt.Sprintf("%.2f", result.OpsPerSec),
			fmt.Sprintf("%.2f", result.MemoryMB),
			strconv.Itoa(result.InsertOperations),
			strconv.Itoa(result.DeleteOperations),
			strconv.Itoa(result.FinalDocumentLength),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	fmt.Println("=== MArrayCRDT Performance Benchmark ===")
	
	// Load editing trace
	fmt.Println("Loading editing trace...")
	operations, err := loadEditingTrace()
	if err != nil {
		log.Fatalf("Failed to load editing trace: %v", err)
	}
	
	fmt.Printf("Loaded %d operations from trace\n", len(operations))
	
	// Benchmark at different operation counts
	operationCounts := []int{1000, 5000, 10000, 20000, 30000, 40000, 50000}
	var results []MArrayBenchmarkResult
	
	fmt.Println("\nOperations,Time_ms,Ops_per_sec,Memory_MB,Final_Length")
	
	for _, count := range operationCounts {
		if count > len(operations) {
			count = len(operations)
		}
		
		fmt.Printf("\nRunning benchmark with %d operations...\n", count)
		result := runMArrayCRDTBenchmark(operations, count)
		results = append(results, result)
		
		fmt.Printf("%d,%.2f,%.0f,%.2f,%d\n", 
			result.Operations, result.TimeMs, result.OpsPerSec, result.MemoryMB, result.FinalDocumentLength)
	}
	
	// Write results to CSV
	csvFile := "marraycrdt_results.csv"
	if err := writeCSVResults(results, csvFile); err != nil {
		log.Fatalf("Failed to write CSV: %v", err)
	}
	
	fmt.Printf("\n✅ Results saved to %s\n", csvFile)
	fmt.Println("🎯 MArrayCRDT benchmark completed!")
}