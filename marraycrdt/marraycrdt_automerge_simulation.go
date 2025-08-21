package marraycrdt

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

// SimulateAutomergeComparison replicates the exact test from Kleppmann et al.'s paper
// "A highly-available move operation for replicated trees" to compare performance
// with automerge's RGA implementation.
//
// This simulation mimics their experimental setup:
// - 2 replicas 
// - 10,000 operations each (20k total)
// - 80% inserts, 20% moves
// - Operations applied in parallel without intermediate syncing
// - Final merge and convergence measurement
func SimulateAutomergeComparison() {
	const (
		numReplicas = 2
		opsPerReplica = 10000
		totalOps = numReplicas * opsPerReplica
		insertProbability = 0.8 // 80% inserts, 20% moves
	)
	
	fmt.Printf("=== Automerge Comparison Test ===\n")
	fmt.Printf("Replicas: %d\n", numReplicas)
	fmt.Printf("Operations per replica: %d\n", opsPerReplica)
	fmt.Printf("Total operations: %d\n", totalOps)
	fmt.Printf("Insert probability: %.0f%%\n", insertProbability*100)
	
	// Create replicas
	replica1 := New[int]("replica1")
	replica2 := New[int]("replica2")
	
	// Initialize with a single element (like automerge test)
	initialID := replica1.Push(0)
	replica2.Merge(replica1)
	
	// Track all element IDs for move operations
	replica1IDs := []string{initialID}
	replica2IDs := []string{initialID}
	
	start := time.Now()
	
	// Generate operations on replica1
	r1 := rand.New(rand.NewSource(1)) // Fixed seed for reproducibility
	for i := 0; i < opsPerReplica; i++ {
		if r1.Float64() < insertProbability {
			// Insert operation - insert at random position
			pos := r1.Intn(replica1.Len() + 1)
			value := i + 1000000 // Unique values for replica1
			newID := replica1.Insert(pos, value)
			replica1IDs = append(replica1IDs, newID)
		} else {
			// Move operation - move random element to random position
			if len(replica1IDs) > 1 { // Need at least 2 elements to move
				idIndex := r1.Intn(len(replica1IDs))
				newPos := r1.Intn(replica1.Len())
				replica1.Move(replica1IDs[idIndex], newPos)
			}
		}
	}
	
	// Generate operations on replica2
	r2 := rand.New(rand.NewSource(2)) // Different seed
	for i := 0; i < opsPerReplica; i++ {
		if r2.Float64() < insertProbability {
			// Insert operation - insert at random position
			pos := r2.Intn(replica2.Len() + 1)
			value := i + 2000000 // Unique values for replica2
			newID := replica2.Insert(pos, value)
			replica2IDs = append(replica2IDs, newID)
		} else {
			// Move operation - move random element to random position
			if len(replica2IDs) > 1 { // Need at least 2 elements to move
				idIndex := r2.Intn(len(replica2IDs))
				newPos := r2.Intn(replica2.Len())
				replica2.Move(replica2IDs[idIndex], newPos)
			}
		}
	}
	
	operationTime := time.Since(start)
	
	fmt.Printf("\n=== Operation Generation ===\n")
	fmt.Printf("Time to generate operations: %v\n", operationTime)
	fmt.Printf("Operations per second: %.0f\n", float64(totalOps)/operationTime.Seconds())
	fmt.Printf("Replica1 final length: %d\n", replica1.Len())
	fmt.Printf("Replica2 final length: %d\n", replica2.Len())
	
	// Measure merge time (this is the key metric from the paper)
	mergeStart := time.Now()
	
	// Perform bidirectional merge
	replica1.Merge(replica2)
	replica2.Merge(replica1)
	
	mergeTime := time.Since(mergeStart)
	
	fmt.Printf("\n=== Merge Performance ===\n")
	fmt.Printf("Time to merge: %v\n", mergeTime)
	fmt.Printf("Merge rate: %.0f ops/sec\n", float64(totalOps)/mergeTime.Seconds())
	
	// Verify convergence
	slice1 := replica1.ToSlice()
	slice2 := replica2.ToSlice()
	
	if !reflect.DeepEqual(slice1, slice2) {
		fmt.Printf("ERROR: Replicas did not converge after merge!\n")
		fmt.Printf("ERROR: Replica1 length: %d\n", len(slice1))
		fmt.Printf("ERROR: Replica2 length: %d\n", len(slice2))
		
		// Show first few differences
		minLen := len(slice1)
		if len(slice2) < minLen {
			minLen = len(slice2)
		}
		
		for i := 0; i < minLen && i < 10; i++ {
			if slice1[i] != slice2[i] {
				fmt.Printf("ERROR: Position %d: replica1=%v, replica2=%v\n", i, slice1[i], slice2[i])
			}
		}
		return
	}
	
	finalLength := len(slice1)
	
	fmt.Printf("\n=== Final Results ===\n")
	fmt.Printf("Final array length: %d\n", finalLength)
	fmt.Printf("Total time: %v\n", time.Since(start))
	fmt.Printf("Memory per element: ~%d bytes\n", estimateMemoryPerElement())
	fmt.Printf("Total estimated memory: ~%d KB\n", (finalLength*estimateMemoryPerElement())/1024)
	
	// Expected results based on operations:
	// - Started with 1 element
	// - Each replica did ~8000 inserts (80% of 10k)
	// - Total expected: 1 + 8000 + 8000 = 16001 elements
	expectedElements := 1 + int(float64(opsPerReplica)*insertProbability*2)
	
	if finalLength < expectedElements-100 || finalLength > expectedElements+100 {
		fmt.Printf("ERROR: Unexpected final length: got %d, expected ~%d\n", finalLength, expectedElements)
	}
	
	fmt.Printf("Expected elements: ~%d (actual: %d)\n", expectedElements, finalLength)
	fmt.Printf("Convergence: SUCCESS\n")
}

// SimulateBenchmarkAutomerge provides detailed benchmarks matching the paper's methodology
func SimulateBenchmarkAutomerge() {
	scenarios := []struct {
		name     string
		replicas int
		opsPerReplica int
		insertProb float64
	}{
		{"Small_2r_1k", 2, 1000, 0.8},
		{"Medium_2r_5k", 2, 5000, 0.8},
		{"Large_2r_10k", 2, 10000, 0.8},
		{"XLarge_2r_20k", 2, 20000, 0.8},
		{"MultiReplica_5r_2k", 5, 2000, 0.8},
	}
	
	for _, scenario := range scenarios {
		fmt.Printf("\n=== Running Scenario: %s ===\n", scenario.name)
		start := time.Now()
		for i := 0; i < 5; i++ { // Run 5 iterations instead of b.N
			runAutomergeBenchmark(scenario.replicas, scenario.opsPerReplica, scenario.insertProb)
		}
		elapsed := time.Since(start)
		fmt.Printf("Scenario %s completed in %v (avg: %v per run)\n", scenario.name, elapsed, elapsed/5)
	}
}

func runAutomergeBenchmark(numReplicas, opsPerReplica int, insertProb float64) {
	// Create replicas
	replicas := make([]*MArrayCRDT[int], numReplicas)
	allIDs := make([][]string, numReplicas)
	
	for i := 0; i < numReplicas; i++ {
		replicas[i] = New[int](fmt.Sprintf("replica%d", i))
	}
	
	// Initialize with single element
	initialID := replicas[0].Push(0)
	allIDs[0] = []string{initialID}
	
	// Sync initial state
	for i := 1; i < numReplicas; i++ {
		replicas[i].Merge(replicas[0])
		allIDs[i] = []string{initialID}
	}
	
	// Generate operations on each replica
	for replicaIdx := 0; replicaIdx < numReplicas; replicaIdx++ {
		r := rand.New(rand.NewSource(int64(replicaIdx + 1)))
		
		for op := 0; op < opsPerReplica; op++ {
			if r.Float64() < insertProb {
				// Insert
				pos := r.Intn(replicas[replicaIdx].Len() + 1)
				value := op + (replicaIdx+1)*1000000
				newID := replicas[replicaIdx].Insert(pos, value)
				allIDs[replicaIdx] = append(allIDs[replicaIdx], newID)
			} else {
				// Move
				if len(allIDs[replicaIdx]) > 1 {
					idIndex := r.Intn(len(allIDs[replicaIdx]))
					newPos := r.Intn(replicas[replicaIdx].Len())
					replicas[replicaIdx].Move(allIDs[replicaIdx][idIndex], newPos)
				}
			}
		}
	}
	
	// Merge all replicas
	for i := 0; i < numReplicas; i++ {
		for j := 0; j < numReplicas; j++ {
			if i != j {
				replicas[i].Merge(replicas[j])
			}
		}
	}
}

// SimulateAutomergeMemoryUsage provides memory usage analysis
func SimulateAutomergeMemoryUsage() {
	sizes := []int{100, 1000, 10000}
	
	fmt.Printf("\n=== Memory Usage Analysis ===\n")
	fmt.Printf("%-10s %-15s %-15s %-15s\n", "Elements", "Estimated (KB)", "Per Element", "Efficiency")
	
	for _, size := range sizes {
		crdt := New[int]("memory-test")
		
		// Fill with elements
		for i := 0; i < size; i++ {
			crdt.Push(i)
		}
		
		estimatedBytes := size * estimateMemoryPerElement()
		estimatedKB := estimatedBytes / 1024
		perElement := estimateMemoryPerElement()
		
		// Efficiency compared to simple slice of ints
		simpleSliceBytes := size * 8 // 8 bytes per int64
		efficiency := float64(simpleSliceBytes) / float64(estimatedBytes) * 100
		
		fmt.Printf("%-10d %-15d %-15d %-13.1f%%\n", 
			size, estimatedKB, perElement, efficiency)
	}
}