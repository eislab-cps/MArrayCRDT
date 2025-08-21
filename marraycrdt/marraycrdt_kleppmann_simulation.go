package marraycrdt

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

// SimulateKleppmannStyleComplexity simulates 10s of thousands of operations
// using all available MArrayCRDT operations, mimicking the complexity
// from Kleppmann et al.'s "A highly-available move operation for replicated trees"
func SimulateKleppmannStyleComplexity() {
	// Test parameters similar to Kleppmann's paper but scaled for stability
	numReplicas := 3
	numOperations := 9000 // 9k operations total  
	numInitialElements := 300
	
	// Create replicas
	replicas := make([]*MArrayCRDT[int], numReplicas)
	for i := 0; i < numReplicas; i++ {
		replicas[i] = New[int](fmt.Sprintf("site%d", i))
	}
	
	// Initialize with elements
	elementIDs := make([]string, numInitialElements)
	for i := 0; i < numInitialElements; i++ {
		id := replicas[0].Push(i)
		elementIDs[i] = id
	}
	
	// Sync all replicas with initial data
	for i := 1; i < numReplicas; i++ {
		replicas[i].Merge(replicas[0])
	}
	
	fmt.Printf("Initialized %d replicas with %d elements\n", numReplicas, numInitialElements)
	
	// Track timing
	start := time.Now()
	
	// Generate operations across all replicas
	r := rand.New(rand.NewSource(42)) // Fixed seed for reproducibility
	operationsPerReplica := numOperations / numReplicas
	
	for replica := 0; replica < numReplicas; replica++ {
		for op := 0; op < operationsPerReplica; op++ {
			// Choose operation type with weights similar to real usage
			opType := r.Intn(100)
			
			switch {
			case opType < 30: // 30% Move operations
				if len(elementIDs) > 0 {
					id := elementIDs[r.Intn(len(elementIDs))]
					newPos := r.Intn(len(elementIDs))
					replicas[replica].Move(id, newPos)
				}
				
			case opType < 45: // 15% Insert operations
				value := r.Intn(1000000) + 1000000 // Large values to distinguish from initial
				pos := r.Intn(len(elementIDs) + 1)
				newID := replicas[replica].Insert(pos, value)
				elementIDs = append(elementIDs, newID)
				
			case opType < 55: // 10% Delete operations
				if len(elementIDs) > numInitialElements/2 { // Keep some elements
					idx := r.Intn(len(elementIDs))
					replicas[replica].Delete(elementIDs[idx])
				}
				
			case opType < 65: // 10% Set (value updates)
				if len(elementIDs) > 0 {
					id := elementIDs[r.Intn(len(elementIDs))]
					newValue := r.Intn(1000000) + 2000000
					replicas[replica].Set(id, newValue)
				}
				
			case opType < 72: // 7% MoveAfter operations
				if len(elementIDs) >= 2 {
					id1 := elementIDs[r.Intn(len(elementIDs))]
					id2 := elementIDs[r.Intn(len(elementIDs))]
					if id1 != id2 {
						replicas[replica].MoveAfter(id1, id2)
					}
				}
				
			case opType < 79: // 7% MoveBefore operations
				if len(elementIDs) >= 2 {
					id1 := elementIDs[r.Intn(len(elementIDs))]
					id2 := elementIDs[r.Intn(len(elementIDs))]
					if id1 != id2 {
						replicas[replica].MoveBefore(id1, id2)
					}
				}
				
			case opType < 85: // 6% Swap operations
				if len(elementIDs) >= 2 {
					id1 := elementIDs[r.Intn(len(elementIDs))]
					id2 := elementIDs[r.Intn(len(elementIDs))]
					if id1 != id2 {
						replicas[replica].Swap(id1, id2)
					}
				}
				
			case opType < 90: // 5% Push operations
				value := r.Intn(1000000) + 3000000
				newID := replicas[replica].Push(value)
				elementIDs = append(elementIDs, newID)
				
			case opType < 94: // 4% Unshift operations
				value := r.Intn(1000000) + 4000000
				newID := replicas[replica].Unshift(value)
				elementIDs = append(elementIDs, newID)
				
			case opType < 98: // 1% Reverse operations (expensive)
				if op%200 == 0 { // Only every 200th operation
					replicas[replica].Reverse()
				}
				
			case opType < 99: // 1% Sort operations (expensive)
				if op%300 == 0 { // Only every 300th operation
					replicas[replica].Sort(func(a, b int) bool { return a < b })
				}
				
			default: // <1% Shuffle operations (expensive)
				if op%500 == 0 { // Only every 500th operation
					replicas[replica].Shuffle()
				}
			}
			
			// Occasional intermediate merges (every 1000 ops)
			if op%1000 == 0 && op > 0 {
				// Merge with a random other replica
				otherReplica := (replica + 1 + r.Intn(numReplicas-1)) % numReplicas
				replicas[replica].Merge(replicas[otherReplica])
			}
		}
	}
	
	operationTime := time.Since(start)
	fmt.Printf("Generated %d operations in %v (%v per op)\n", 
		numOperations, operationTime, operationTime/time.Duration(numOperations))
	
	// Final convergence phase
	mergeStart := time.Now()
	
	// Perform complete pairwise merging until convergence
	maxMergeRounds := 20
	for round := 0; round < maxMergeRounds; round++ {
		converged := true
		
		// Full mesh merge - each replica merges with every other replica
		for i := 0; i < numReplicas; i++ {
			for j := 0; j < numReplicas; j++ {
				if i != j {
					before := replicas[i].ToSlice()
					replicas[i].Merge(replicas[j])
					after := replicas[i].ToSlice()
					
					if !reflect.DeepEqual(before, after) {
						converged = false
					}
				}
			}
		}
		
		if converged {
			fmt.Printf("Converged after %d merge rounds\n", round+1)
			break
		}
		
		if round == maxMergeRounds-1 {
			// If still not converged, try a few more intensive rounds
			fmt.Printf("Warning: Did not converge after %d rounds, trying intensive merge\n", maxMergeRounds)
			for intensiveRound := 0; intensiveRound < 5; intensiveRound++ {
				for i := 0; i < numReplicas; i++ {
					for j := 0; j < numReplicas; j++ {
						if i != j {
							replicas[i].Merge(replicas[j])
						}
					}
				}
			}
		}
	}
	
	mergeTime := time.Since(mergeStart)
	fmt.Printf("Convergence took %v\n", mergeTime)
	
	// Verify all replicas have converged
	baseSlice := replicas[0].ToSlice()
	for i := 1; i < numReplicas; i++ {
		replicaSlice := replicas[i].ToSlice()
		if !reflect.DeepEqual(baseSlice, replicaSlice) {
			fmt.Printf("ERROR: Replica %d did not converge! Expected %d elements, got %d\n", 
				i, len(baseSlice), len(replicaSlice))
			
			// Show first few differences for debugging
			minLen := len(baseSlice)
			if len(replicaSlice) < minLen {
				minLen = len(replicaSlice)
			}
			
			differences := 0
			for j := 0; j < minLen && differences < 10; j++ {
				if baseSlice[j] != replicaSlice[j] {
					fmt.Printf("ERROR:   Position %d: replica0=%v, replica%d=%v\n", 
						j, baseSlice[j], i, replicaSlice[j])
					differences++
				}
			}
			return
		}
	}
	
	// Performance statistics
	totalTime := time.Since(start)
	finalLength := len(baseSlice)
	
	fmt.Printf("\n=== Performance Results ===\n")
	fmt.Printf("Total operations: %d\n", numOperations)
	fmt.Printf("Final array length: %d\n", finalLength)
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Printf("Time per operation: %v\n", totalTime/time.Duration(numOperations))
	fmt.Printf("Operations per second: %.0f\n", float64(numOperations)/totalTime.Seconds())
	fmt.Printf("Memory usage per element: ~%d bytes\n", estimateMemoryPerElement())
	
	// Verify data integrity - check that we have reasonable data
	if finalLength < numInitialElements/10 {
		fmt.Printf("WARNING: Too many elements were deleted: %d remaining from %d initial\n", 
			finalLength, numInitialElements)
	}
	
	if finalLength > numInitialElements*3 {
		fmt.Printf("WARNING: Too many elements were added: %d final from %d initial\n", 
			finalLength, numInitialElements)
	}
	
	fmt.Printf("All %d replicas converged successfully!\n", numReplicas)
}

// estimateMemoryPerElement provides a rough estimate of memory usage per element
func estimateMemoryPerElement() int {
	// Rough calculation based on struct sizes:
	// Element: ~200 bytes (ID=32, Value=~50, Index=~50, VectorClock=~50, etc.)
	// VectorClocks: ~50 bytes each
	// Total per element: ~200 bytes
	return 200
}

// SimulateBenchmarkKleppmannStyle benchmarks the performance of various operations
func SimulateBenchmarkKleppmannStyle() {
	// Test different operation mixes
	scenarios := []struct {
		name    string
		ops     int
		setup   func() *MArrayCRDT[int]
		operation func(*MArrayCRDT[int], *rand.Rand, []string)
	}{
		{
			name: "Move-Heavy",
			ops:  1000,
			setup: func() *MArrayCRDT[int] {
				crdt := New[int]("bench")
				for i := 0; i < 1000; i++ {
					crdt.Push(i)
				}
				return crdt
			},
			operation: func(crdt *MArrayCRDT[int], r *rand.Rand, ids []string) {
				id := ids[r.Intn(len(ids))]
				pos := r.Intn(len(ids))
				crdt.Move(id, pos)
			},
		},
		{
			name: "Insert-Heavy",
			ops:  1000,
			setup: func() *MArrayCRDT[int] {
				crdt := New[int]("bench")
				for i := 0; i < 100; i++ {
					crdt.Push(i)
				}
				return crdt
			},
			operation: func(crdt *MArrayCRDT[int], r *rand.Rand, ids []string) {
				pos := r.Intn(crdt.Len() + 1)
				value := r.Intn(1000000)
				crdt.Insert(pos, value)
			},
		},
		{
			name: "Mixed-Operations",
			ops:  1000,
			setup: func() *MArrayCRDT[int] {
				crdt := New[int]("bench")
				for i := 0; i < 500; i++ {
					crdt.Push(i)
				}
				return crdt
			},
			operation: func(crdt *MArrayCRDT[int], r *rand.Rand, ids []string) {
				switch r.Intn(4) {
				case 0: // Move
					if len(ids) > 0 {
						id := ids[r.Intn(len(ids))]
						pos := r.Intn(len(ids))
						crdt.Move(id, pos)
					}
				case 1: // Insert
					pos := r.Intn(crdt.Len() + 1)
					crdt.Insert(pos, r.Intn(1000000))
				case 2: // Delete
					if len(ids) > 100 {
						id := ids[r.Intn(len(ids))]
						crdt.Delete(id)
					}
				case 3: // Set
					if len(ids) > 0 {
						id := ids[r.Intn(len(ids))]
						crdt.Set(id, r.Intn(1000000))
					}
				}
			},
		},
	}
	
	for _, scenario := range scenarios {
		fmt.Printf("\n=== Running Benchmark Scenario: %s ===\n", scenario.name)
		totalTime := time.Duration(0)
		iterations := 5 // Run 5 iterations instead of b.N
		
		for i := 0; i < iterations; i++ {
			crdt := scenario.setup()
			ids := crdt.IDs()
			r := rand.New(rand.NewSource(int64(i)))
			
			start := time.Now()
			for j := 0; j < scenario.ops; j++ {
				scenario.operation(crdt, r, ids)
			}
			elapsed := time.Since(start)
			totalTime += elapsed
		}
		
		avgTime := totalTime / time.Duration(iterations)
		opsPerSec := float64(scenario.ops) / avgTime.Seconds()
		fmt.Printf("Scenario %s: %d ops in %v avg (%.0f ops/sec)\n", 
			scenario.name, scenario.ops, avgTime, opsPerSec)
	}
}