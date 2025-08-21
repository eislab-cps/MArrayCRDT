package marraycrdt

import (
	"fmt"
	"reflect"
	"testing"
)

// TestConcurrentMoveSameItemMultipleReplicas tests when 3+ replicas all move the same item
func TestConcurrentMoveSameItemMultipleReplicas(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")
	replica3 := New[string]("site3")
	replica4 := New[string]("site4")

	// Setup
	_ = replica1.Push("A")
	idB := replica1.Push("B")
	_ = replica1.Push("C")
	_ = replica1.Push("D")

	// Sync to all
	replica2.Merge(replica1)
	replica3.Merge(replica1)
	replica4.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// All replicas move B to different positions concurrently
	replica1.Move(idB, 0) // B to start
	replica2.Move(idB, 3) // B to end
	replica3.Move(idB, 1) // B to middle-front
	replica4.Move(idB, 2) // B to middle-back

	fmt.Printf("\nAfter each replica moves B:\n")
	fmt.Printf("Replica1 (B->0): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (B->3): %v\n", replica2.ToSlice())
	fmt.Printf("Replica3 (B->1): %v\n", replica3.ToSlice())
	fmt.Printf("Replica4 (B->2): %v\n", replica4.ToSlice())

	// Full merge
	for i := 0; i < 3; i++ { // Multiple rounds to ensure convergence
		replica1.Merge(replica2)
		replica2.Merge(replica3)
		replica3.Merge(replica4)
		replica4.Merge(replica1)
	}

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("All replicas: %v\n", replica1.ToSlice())

	// Verify convergence
	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) ||
		!reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) ||
		!reflect.DeepEqual(replica3.ToSlice(), replica4.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestMultipleItemsToSamePosition tests when multiple items move to the same position
func TestMultipleItemsToSamePosition(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")
	replica3 := New[string]("site3")

	_ = replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")
	idD := replica1.Push("D")
	_ = replica1.Push("E")

	replica2.Merge(replica1)
	replica3.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Three different items all move to position 0 concurrently
	replica1.Move(idB, 0)
	replica2.Move(idC, 0)
	replica3.Move(idD, 0)

	fmt.Printf("\nAfter concurrent moves to position 0:\n")
	fmt.Printf("Replica1 (B->0): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (C->0): %v\n", replica2.ToSlice())
	fmt.Printf("Replica3 (D->0): %v\n", replica3.ToSlice())

	// Merge all
	replica1.Merge(replica2)
	replica1.Merge(replica3)
	replica2.Merge(replica1)
	replica3.Merge(replica1)

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("All replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) ||
		!reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestOverlappingSwaps tests concurrent swaps that share elements
func TestOverlappingSwaps(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")

	idA := replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")
	_ = replica1.Push("D")

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Overlapping swaps: A<->B and B<->C
	// This is tricky because B is involved in both swaps
	replica1.Swap(idA, idB)
	replica2.Swap(idB, idC)

	fmt.Printf("\nAfter overlapping swaps:\n")
	fmt.Printf("Replica1 (A<->B): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (B<->C): %v\n", replica2.ToSlice())

	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("Both replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestCircularMoves tests A->B's position, B->C's position, C->A's position
func TestCircularMoves(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")
	replica3 := New[string]("site3")

	idA := replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")
	_ = replica1.Push("D")

	replica2.Merge(replica1)
	replica3.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Circular moves - each item moves to another's position
	replica1.Move(idA, 1) // A moves to B's position
	replica2.Move(idB, 2) // B moves to C's position
	replica3.Move(idC, 0) // C moves to A's position

	fmt.Printf("\nAfter circular moves:\n")
	fmt.Printf("Replica1 (A->pos1): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (B->pos2): %v\n", replica2.ToSlice())
	fmt.Printf("Replica3 (C->pos0): %v\n", replica3.ToSlice())

	// Full merge
	for i := 0; i < 3; i++ {
		replica1.Merge(replica2)
		replica2.Merge(replica3)
		replica3.Merge(replica1)
	}

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("All replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) ||
		!reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestMoveDeletedItem tests concurrent move and delete
func TestMoveDeletedItem(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")

	_ = replica1.Push("A")
	idB := replica1.Push("B")
	_ = replica1.Push("C")

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Replica1 deletes B while Replica2 moves it
	replica1.Delete(idB)
	replica2.Move(idB, 0)

	fmt.Printf("\nAfter concurrent delete and move:\n")
	fmt.Printf("Replica1 (deleted B): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (moved B): %v\n", replica2.ToSlice())

	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge! R1=%v, R2=%v", replica1.ToSlice(), replica2.ToSlice())
	}
}

// TestInsertWhileMoving tests inserting at positions while items are moving there
func TestInsertWhileMoving(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")
	replica3 := New[string]("site3")

	_ = replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")

	replica2.Merge(replica1)
	replica3.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Concurrent operations at position 1:
	// Replica1: Insert "X" at position 1
	// Replica2: Move C to position 1
	// Replica3: Move B to position 1
	replica1.Insert(1, "X")
	replica2.Move(idC, 1)
	replica3.Move(idB, 1)

	fmt.Printf("\nAfter concurrent insert and moves to position 1:\n")
	fmt.Printf("Replica1 (insert X): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (move C): %v\n", replica2.ToSlice())
	fmt.Printf("Replica3 (move B): %v\n", replica3.ToSlice())

	// Merge all
	for i := 0; i < 3; i++ {
		replica1.Merge(replica2)
		replica2.Merge(replica3)
		replica3.Merge(replica1)
	}

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("All replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) ||
		!reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestRapidSequentialMoves tests multiple moves of same item in quick succession
func TestRapidSequentialMoves(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")

	_ = replica1.Push("A")
	idB := replica1.Push("B")
	_ = replica1.Push("C")
	_ = replica1.Push("D")

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Replica1 does multiple moves of B
	replica1.Move(idB, 3)
	replica1.Move(idB, 0)
	replica1.Move(idB, 2)

	// Replica2 does different moves of B (concurrent with all of replica1's)
	replica2.Move(idB, 1)

	fmt.Printf("\nAfter rapid moves:\n")
	fmt.Printf("Replica1 (B: 1->3->0->2): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (B->1): %v\n", replica2.ToSlice())

	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("Both replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestMoveAfterChain tests A moves after B, B moves after C, C moves after D
func TestMoveAfterChain(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")
	replica3 := New[string]("site3")

	idA := replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")
	idD := replica1.Push("D")

	replica2.Merge(replica1)
	replica3.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Chain of MoveAfter operations
	replica1.MoveAfter(idA, idB)
	replica2.MoveAfter(idB, idC)
	replica3.MoveAfter(idC, idD)

	fmt.Printf("\nAfter chained MoveAfter:\n")
	fmt.Printf("Replica1 (A after B): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (B after C): %v\n", replica2.ToSlice())
	fmt.Printf("Replica3 (C after D): %v\n", replica3.ToSlice())

	// Merge
	for i := 0; i < 3; i++ {
		replica1.Merge(replica2)
		replica2.Merge(replica3)
		replica3.Merge(replica1)
	}

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("All replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) ||
		!reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestTripleSwap tests A<->B, B<->C, C<->A concurrently
func TestTripleSwap(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")
	replica3 := New[string]("site3")

	idA := replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")
	_ = replica1.Push("D")

	replica2.Merge(replica1)
	replica3.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Circular swaps
	replica1.Swap(idA, idB)
	replica2.Swap(idB, idC)
	replica3.Swap(idC, idA)

	fmt.Printf("\nAfter triple swap:\n")
	fmt.Printf("Replica1 (A<->B): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (B<->C): %v\n", replica2.ToSlice())
	fmt.Printf("Replica3 (C<->A): %v\n", replica3.ToSlice())

	// Merge
	for i := 0; i < 3; i++ {
		replica1.Merge(replica2)
		replica2.Merge(replica3)
		replica3.Merge(replica1)
	}

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("All replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) ||
		!reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestFractionalIndexStress tests many moves to same position to stress fractional indices
func TestFractionalIndexStress(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")

	idA := replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")
	idD := replica1.Push("D")
	idE := replica1.Push("E")

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Repeatedly move items between A and B to stress fractional indices
	for i := 0; i < 20; i++ {
		if i%2 == 0 {
			replica1.MoveAfter(idC, idA)
			replica2.MoveAfter(idD, idA)
		} else {
			replica1.MoveAfter(idE, idA)
			replica2.MoveAfter(idC, idA)
		}

		// Merge to propagate changes
		replica1.Merge(replica2)
		replica2.Merge(replica1)
	}

	fmt.Printf("\nAfter %d moves between same positions:\n", 20)
	fmt.Printf("Both replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge after fractional index stress!")
	}

	// Check if reindexing would help (indices getting too close)
	elem1, _ := replica1.GetElement(idA)
	elem2, _ := replica1.GetElement(idB)
	if elem1 != nil && elem2 != nil {
		diff := elem2.Index.Position - elem1.Index.Position
		fmt.Printf("Index difference between A and B: %v\n", diff)
		if diff < 0.0001 {
			t.Logf("Warning: Indices getting very close: %v", diff)
		}
	}
}

// TestInvalidPositionMoves tests moves to invalid positions
func TestInvalidPositionMoves(t *testing.T) {
	replica1 := New[string]("site1")

	_ = replica1.Push("A")
	idB := replica1.Push("B")
	_ = replica1.Push("C")

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Try invalid positions
	replica1.Move(idB, -1)   // Negative position
	replica1.Move(idB, 1000) // Beyond array bounds

	fmt.Printf("After moves to invalid positions: %v\n", replica1.ToSlice())

	// Should handle gracefully (clamp to valid range)
	if len(replica1.ToSlice()) != 3 {
		t.Errorf("Array corrupted by invalid position moves")
	}
}

// TestConcurrentSortAndMove tests sort operation with concurrent moves
func TestConcurrentSortAndMove(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")

	_ = replica1.Push("Charlie")
	_ = replica1.Push("Alice")
	idD := replica1.Push("David")
	_ = replica1.Push("Bob")

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Replica1 sorts alphabetically while Replica2 moves items
	replica1.Sort(func(a, b string) bool { return a < b })
	replica2.Move(idD, 0) // Move David to start

	fmt.Printf("\nAfter concurrent sort and move:\n")
	fmt.Printf("Replica1 (sorted): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (moved David): %v\n", replica2.ToSlice())

	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("Both replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestConcurrentReverseAndSwap tests reverse with concurrent swaps
func TestConcurrentReverseAndSwap(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")

	idA := replica1.Push("A")
	_ = replica1.Push("B")
	_ = replica1.Push("C")
	idD := replica1.Push("D")

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Replica1 reverses while Replica2 swaps
	replica1.Reverse()
	replica2.Swap(idA, idD)

	fmt.Printf("\nAfter concurrent reverse and swap:\n")
	fmt.Printf("Replica1 (reversed): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (swapped A-D): %v\n", replica2.ToSlice())

	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("Both replicas: %v\n", replica1.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestDelayedMergeWithMultipleMoves tests merging after many unsynced operations
func TestDelayedMergeWithMultipleMoves(t *testing.T) {
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")

	// Initial setup
	ids := make([]string, 6)
	for i := 0; i < 6; i++ {
		ids[i] = replica1.Push(fmt.Sprintf("%c", 'A'+i))
	}

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Many operations on replica1 without merging
	replica1.Move(ids[1], 0)      // B to start
	replica1.Swap(ids[2], ids[3]) // C <-> D
	replica1.Move(ids[4], 2)      // E to middle
	replica1.Delete(ids[5])       // Delete F
	replica1.Insert(3, "X")       // Insert X

	// Many operations on replica2 without merging
	replica2.Move(ids[0], 5)      // A to end
	replica2.Swap(ids[1], ids[4]) // B <-> E
	replica2.Move(ids[3], 0)      // D to start
	replica2.Set(ids[2], "C-modified")
	replica2.Insert(2, "Y") // Insert Y

	fmt.Printf("\nAfter many unsynced operations:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	// Now merge
	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter convergence:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge after delayed merge!")
	}
}

// TestClockDriftSimulation simulates clock drift effects
func TestClockDriftSimulation(t *testing.T) {
	// Create replicas with simulated clock drift
	replica1 := New[string]("site1")
	replica2 := New[string]("site2")

	idA := replica1.Push("A")
	idB := replica1.Push("B")
	_ = replica1.Push("C")

	replica2.Merge(replica1)

	// Simulate operations happening "at the same time" but with clock drift
	// In reality, vector clocks handle this, but let's test rapid operations
	done1 := make(chan bool)
	done2 := make(chan bool)

	go func() {
		replica1.Move(idA, 2)
		replica1.Move(idB, 0)
		done1 <- true
	}()

	go func() {
		replica2.Move(idB, 2)
		replica2.Move(idA, 0)
		done2 <- true
	}()

	<-done1
	<-done2

	// Merge
	replica1.Merge(replica2)
	replica2.Merge(replica1)

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge with concurrent operations!")
	}
}

// TestExtremeStressTest performs many random operations
func TestExtremeStressTest(t *testing.T) {
	replica1 := New[int]("site1")
	replica2 := New[int]("site2")
	replica3 := New[int]("site3")

	// Add initial elements
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		ids[i] = replica1.Push(i)
	}

	fmt.Printf("Initial state: %v\n", replica1.ToSlice())
	fmt.Printf("Initial IDs: %v\n", ids)

	replica2.Merge(replica1)
	replica3.Merge(replica1)

	// Perform many random operations on each replica
	for i := 0; i < 50; i++ {
		fmt.Printf("\n=== Operation %d ===\n", i)
		switch i % 7 {
		case 0:
			fmt.Printf("Replica1: Move %s (value %d) to position %d\n", ids[i%10][:8], i%10, (i*3)%10)
			replica1.Move(ids[i%10], (i*3)%10)
		case 1:
			fmt.Printf("Replica2: Swap %s (value %d) with %s (value %d)\n", 
				ids[i%10][:8], i%10, ids[(i+1)%10][:8], (i+1)%10)
			replica2.Swap(ids[i%10], ids[(i+1)%10])
		case 2:
			fmt.Printf("Replica3: Move %s (value %d) to position %d\n", ids[i%10][:8], i%10, (i*7)%10)
			replica3.Move(ids[i%10], (i*7)%10)
		case 3:
			if i%3 == 0 {
				fmt.Printf("Replica1: Reverse\n")
				replica1.Reverse()
			}
		case 4:
			fmt.Printf("Replica2: MoveAfter %s (value %d) after %s (value %d)\n", 
				ids[i%10][:8], i%10, ids[(i+3)%10][:8], (i+3)%10)
			replica2.MoveAfter(ids[i%10], ids[(i+3)%10])
		case 5:
			fmt.Printf("Replica3: MoveBefore %s (value %d) before %s (value %d)\n", 
				ids[i%10][:8], i%10, ids[(i+5)%10][:8], (i+5)%10)
			replica3.MoveBefore(ids[i%10], ids[(i+5)%10])
		case 6:
			fmt.Printf("Replica1: Set %s to value %d\n", ids[i%10][:8], i*100)
			replica1.Set(ids[i%10], i*100)
		}

		// Print state after operation
		fmt.Printf("After operation:\n")
		fmt.Printf("  Replica1: %v\n", replica1.ToSlice())
		fmt.Printf("  Replica2: %v\n", replica2.ToSlice())
		fmt.Printf("  Replica3: %v\n", replica3.ToSlice())

		// Occasional merges
		if i%10 == 0 {
			fmt.Printf("\nPerforming intermediate merge at operation %d\n", i)
			replica1.Merge(replica2)
			replica2.Merge(replica3)
			replica3.Merge(replica1)
			
			fmt.Printf("After intermediate merge:\n")
			fmt.Printf("  Replica1: %v\n", replica1.ToSlice())
			fmt.Printf("  Replica2: %v\n", replica2.ToSlice())
			fmt.Printf("  Replica3: %v\n", replica3.ToSlice())
		}
	}

	fmt.Printf("\n=== Starting final merge rounds ===\n")
	// Final merge
	for round := 0; round < 3; round++ {
		fmt.Printf("\nFinal merge round %d:\n", round+1)
		replica1.Merge(replica2)
		fmt.Printf("  After R1.Merge(R2): R1=%v\n", replica1.ToSlice())
		
		replica2.Merge(replica3)
		fmt.Printf("  After R2.Merge(R3): R2=%v\n", replica2.ToSlice())
		
		replica3.Merge(replica1)
		fmt.Printf("  After R3.Merge(R1): R3=%v\n", replica3.ToSlice())
		
		replica1.Merge(replica3)
		fmt.Printf("  After R1.Merge(R3): R1=%v\n", replica1.ToSlice())
		
		replica2.Merge(replica1)
		fmt.Printf("  After R2.Merge(R1): R2=%v\n", replica2.ToSlice())
	}

	fmt.Printf("\n=== Final states ===\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())
	fmt.Printf("Replica3: %v\n", replica3.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		fmt.Printf("\nERROR: Replica1 and Replica2 differ!\n")
		fmt.Printf("Replica1: %v\n", replica1.ToSlice())
		fmt.Printf("Replica2: %v\n", replica2.ToSlice())
	}
	
	if !reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) {
		fmt.Printf("\nERROR: Replica2 and Replica3 differ!\n")
		fmt.Printf("Replica2: %v\n", replica2.ToSlice())
		fmt.Printf("Replica3: %v\n", replica3.ToSlice())
	}

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) ||
		!reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) {
		t.Errorf("Replicas did not converge after extreme stress test!")
	}
}
