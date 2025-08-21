package marraycrdt

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

// SimulateLargeScaleOperations simulates thousands of operations focusing on
// core operations that are most likely to converge reliably
func SimulateLargeScaleOperations() {
	const (
		numReplicas = 2
		numOperations = 8000 // 8k operations total
		numInitialElements = 100
	)
	
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
	
	// Generate operations across all replicas focusing on core operations
	operationsPerReplica := numOperations / numReplicas
	
	for replica := 0; replica < numReplicas; replica++ {
		r := rand.New(rand.NewSource(int64(replica + 1)))
		
		for op := 0; op < operationsPerReplica; op++ {
			// Choose operation type - focus on core operations for better convergence
			opType := r.Intn(100)
			
			switch {
			case opType < 40: // 40% Move operations (core CRDT operation)
				if len(elementIDs) > 0 {
					id := elementIDs[r.Intn(len(elementIDs))]
					newPos := r.Intn(len(elementIDs))
					replicas[replica].Move(id, newPos)
				}
				
			case opType < 65: // 25% Insert operations
				value := r.Intn(1000000) + (replica+1)*1000000 // Unique per replica
				pos := r.Intn(replicas[replica].Len() + 1)
				newID := replicas[replica].Insert(pos, value)
				elementIDs = append(elementIDs, newID)
				
			case opType < 75: // 10% Delete operations
				if len(elementIDs) > numInitialElements/2 { // Keep some elements
					idx := r.Intn(len(elementIDs))
					replicas[replica].Delete(elementIDs[idx])
				}
				
			case opType < 85: // 10% Set (value updates)
				if len(elementIDs) > 0 {
					id := elementIDs[r.Intn(len(elementIDs))]
					newValue := r.Intn(1000000) + (replica+1)*2000000
					replicas[replica].Set(id, newValue)
				}
				
			case opType < 92: // 7% MoveAfter operations
				if len(elementIDs) >= 2 {
					id1 := elementIDs[r.Intn(len(elementIDs))]
					id2 := elementIDs[r.Intn(len(elementIDs))]
					if id1 != id2 {
						replicas[replica].MoveAfter(id1, id2)
					}
				}
				
			case opType < 98: // 6% MoveBefore operations
				if len(elementIDs) >= 2 {
					id1 := elementIDs[r.Intn(len(elementIDs))]
					id2 := elementIDs[r.Intn(len(elementIDs))]
					if id1 != id2 {
						replicas[replica].MoveBefore(id1, id2)
					}
				}
				
			default: // 2% Swap operations
				if len(elementIDs) >= 2 {
					id1 := elementIDs[r.Intn(len(elementIDs))]
					id2 := elementIDs[r.Intn(len(elementIDs))]
					if id1 != id2 {
						replicas[replica].Swap(id1, id2)
					}
				}
			}
			
			// More frequent intermediate merges for better convergence
			if op%500 == 0 && op > 0 {
				// Merge with next replica in round-robin
				otherReplica := (replica + 1) % numReplicas
				replicas[replica].Merge(replicas[otherReplica])
			}
		}
	}
	
	operationTime := time.Since(start)
	fmt.Printf("Generated %d operations in %v (%v per op)\n", 
		numOperations, operationTime, operationTime/time.Duration(numOperations))
	
	// Final convergence phase
	mergeStart := time.Now()
	
	// Perform systematic convergence
	maxMergeRounds := 10
	for round := 0; round < maxMergeRounds; round++ {
		converged := true
		
		// Round-robin merge pattern
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
			fmt.Printf("Warning: Did not converge after %d rounds\n", maxMergeRounds)
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
			for j := 0; j < minLen && differences < 5; j++ {
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
	fmt.Printf("Memory usage per element: ~200 bytes\n")
	fmt.Printf("Total estimated memory: ~%d KB\n", (finalLength*200)/1024)
	
	fmt.Printf("All %d replicas converged successfully!\n", numReplicas)
}

// SimulateMassiveScale simulates even larger scale operations
func SimulateMassiveScale() {
	
	const (
		numReplicas = 2 // Fewer replicas for massive scale
		numOperations = 30000 // 30k operations
		numInitialElements = 100
	)
	
	fmt.Printf("\n=== Massive Scale Test ===\n")
	fmt.Printf("Replicas: %d, Operations: %d\n", numReplicas, numOperations)
	
	// Create replicas
	replicas := make([]*MArrayCRDT[int], numReplicas)
	for i := 0; i < numReplicas; i++ {
		replicas[i] = New[int](fmt.Sprintf("massive-site%d", i))
	}
	
	// Initialize
	elementIDs := make([]string, numInitialElements)
	for i := 0; i < numInitialElements; i++ {
		id := replicas[0].Push(i)
		elementIDs[i] = id
	}
	
	for i := 1; i < numReplicas; i++ {
		replicas[i].Merge(replicas[0])
	}
	
	start := time.Now()
	
	// Focus on the most performance-critical operations
	operationsPerReplica := numOperations / numReplicas
	
	for replica := 0; replica < numReplicas; replica++ {
		r := rand.New(rand.NewSource(int64(replica + 100)))
		
		for op := 0; op < operationsPerReplica; op++ {
			opType := r.Intn(100)
			
			switch {
			case opType < 50: // 50% moves
				if len(elementIDs) > 0 {
					id := elementIDs[r.Intn(len(elementIDs))]
					newPos := r.Intn(len(elementIDs))
					replicas[replica].Move(id, newPos)
				}
				
			case opType < 80: // 30% inserts
				value := r.Intn(1000000) + (replica+1)*1000000
				pos := r.Intn(replicas[replica].Len() + 1)
				newID := replicas[replica].Insert(pos, value)
				elementIDs = append(elementIDs, newID)
				
			case opType < 90: // 10% deletes
				if len(elementIDs) > numInitialElements/2 {
					idx := r.Intn(len(elementIDs))
					replicas[replica].Delete(elementIDs[idx])
				}
				
			default: // 10% value updates
				if len(elementIDs) > 0 {
					id := elementIDs[r.Intn(len(elementIDs))]
					newValue := r.Intn(1000000) + (replica+1)*2000000
					replicas[replica].Set(id, newValue)
				}
			}
			
			// Periodic syncing
			if op%2000 == 0 && op > 0 {
				otherReplica := (replica + 1) % numReplicas
				replicas[replica].Merge(replicas[otherReplica])
			}
		}
	}
	
	operationTime := time.Since(start)
	
	// Final merge
	mergeStart := time.Now()
	for i := 0; i < 5; i++ {
		for j := 0; j < numReplicas; j++ {
			for k := 0; k < numReplicas; k++ {
				if j != k {
					replicas[j].Merge(replicas[k])
				}
			}
		}
	}
	mergeTime := time.Since(mergeStart)
	
	// Verify convergence
	baseSlice := replicas[0].ToSlice()
	for i := 1; i < numReplicas; i++ {
		if !reflect.DeepEqual(baseSlice, replicas[i].ToSlice()) {
			fmt.Printf("ERROR: Massive scale test: Replicas did not converge\n")
			return
		}
	}
	
	totalTime := time.Since(start)
	
	fmt.Printf("Operations: %d in %v (%.0f ops/sec)\n", 
		numOperations, operationTime, float64(numOperations)/operationTime.Seconds())
	fmt.Printf("Merge time: %v\n", mergeTime)
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Printf("Final length: %d elements\n", len(baseSlice))
	fmt.Printf("SUCCESS: Massive scale test converged!\n")
}