package marraycrdt

import (
	"fmt"
	"reflect"
	"testing"
)

// TestConcurrentMoves tests that concurrent moves converge correctly
func TestConcurrentMoves(t *testing.T) {
	// Create two replicas
	replica1 := New[string]("replica1")
	replica2 := New[string]("site2")

	// Add initial items to replica1
	_ = replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")
	_ = replica1.Push("D")

	// Sync to replica2
	replica2.Merge(replica1)

	// Verify both have same initial state
	if !reflect.DeepEqual(replica1.ToSlice(), []string{"A", "B", "C", "D"}) {
		t.Errorf("Replica1 initial state wrong: %v", replica1.ToSlice())
	}
	if !reflect.DeepEqual(replica2.ToSlice(), []string{"A", "B", "C", "D"}) {
		t.Errorf("Replica2 initial state wrong: %v", replica2.ToSlice())
	}

	// Concurrent moves:
	// Replica1: Move B to position 3 (after C)
	// Replica2: Move C to position 1 (after A)
	replica1.Move(idB, 3)
	replica2.Move(idC, 1)

	fmt.Printf("After concurrent moves:\n")
	fmt.Printf("Replica1 (moved B to pos 3): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (moved C to pos 1): %v\n", replica2.ToSlice())

	// Merge both ways
	replica1.Merge(replica2)
	replica2.Merge(replica1)

	// Both should converge to same state
	fmt.Printf("\nAfter merge:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!\nReplica1: %v\nReplica2: %v",
			replica1.ToSlice(), replica2.ToSlice())
	}
}

// TestConcurrentMovesSameElement tests moving the same element concurrently
func TestConcurrentMovesSameElement(t *testing.T) {
	// Create two replicas
	replica1 := New[string]("replica1")
	replica2 := New[string]("site2")

	// Add initial items
	_ = replica1.Push("A")
	idB := replica1.Push("B")
	_ = replica1.Push("C")
	_ = replica1.Push("D")

	// Sync to replica2
	replica2.Merge(replica1)

	fmt.Printf("Initial state: %v\n", replica1.ToSlice())

	// Both replicas move B to different positions concurrently
	// Replica1: Move B to end (position 3)
	// Replica2: Move B to beginning (position 0)
	replica1.Move(idB, 3)
	replica2.Move(idB, 0)

	fmt.Printf("\nAfter concurrent moves of same element:\n")
	fmt.Printf("Replica1 (moved B to end): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (moved B to start): %v\n", replica2.ToSlice())

	// Merge both ways
	replica1.Merge(replica2)
	replica2.Merge(replica1)

	// Both should converge (one move wins deterministically)
	fmt.Printf("\nAfter merge:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!\nReplica1: %v\nReplica2: %v",
			replica1.ToSlice(), replica2.ToSlice())
	}
}

// TestConcurrentMoveAndEdit tests that moves and edits don't interfere
func TestConcurrentMoveAndEdit(t *testing.T) {
	// Create two replicas
	replica1 := New[string]("replica1")
	replica2 := New[string]("site2")

	// Add initial items
	_ = replica1.Push("Apple")
	idB := replica1.Push("Banana")
	_ = replica1.Push("Cherry")

	// Sync to replica2
	replica2.Merge(replica1)

	fmt.Printf("Initial state: %v\n", replica1.ToSlice())

	// Concurrent operations:
	// Replica1: Edit Banana to "Blueberry"
	// Replica2: Move Banana to position 0
	replica1.Set(idB, "Blueberry")
	replica2.Move(idB, 0)

	fmt.Printf("\nAfter concurrent move and edit:\n")
	fmt.Printf("Replica1 (edited B to Blueberry): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (moved B to start): %v\n", replica2.ToSlice())

	// Merge both ways
	replica1.Merge(replica2)
	replica2.Merge(replica1)

	// Both operations should apply: Blueberry at position 0
	fmt.Printf("\nAfter merge (both operations should apply):\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!\nReplica1: %v\nReplica2: %v",
			replica1.ToSlice(), replica2.ToSlice())
	}

	// Verify both operations applied
	expected := replica1.ToSlice()
	if expected[0] != "Blueberry" {
		t.Errorf("Edit didn't apply correctly. First element is %s, expected Blueberry",
			expected[0])
	}
}

// TestComplexConcurrentOperations tests multiple concurrent operations
func TestComplexConcurrentOperations(t *testing.T) {
	// Create three replicas for more complex scenario
	replica1 := New[string]("replica1")
	replica2 := New[string]("site2")
	replica3 := New[string]("site3")

	// Initial setup on replica1
	_ = replica1.Push("Item A")
	idB := replica1.Push("Item B")
	idC := replica1.Push("Item C")
	idD := replica1.Push("Item D")
	idE := replica1.Push("Item E")

	// Sync to all replicas
	replica2.Merge(replica1)
	replica3.Merge(replica1)

	fmt.Printf("Initial state: %v\n", replica1.ToSlice())

	// Concurrent operations:
	// Replica1: Move D to position 1 and edit B
	// Replica2: Move B to position 4 and edit D
	// Replica3: Move C to position 0 and edit E

	replica1.Move(idD, 1)
	replica1.Set(idB, "Item B (edited by replica1)")

	replica2.Move(idB, 4)
	replica2.Set(idD, "Item D (edited by site2)")

	replica3.Move(idC, 0)
	replica3.Set(idE, "Item E (edited by site3)")

	fmt.Printf("\nAfter concurrent operations:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())
	fmt.Printf("Replica3: %v\n", replica3.ToSlice())

	// Merge all replicas in a ring
	replica1.Merge(replica2)
	replica2.Merge(replica3)
	replica3.Merge(replica1)
	replica1.Merge(replica3)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter full merge:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())
	fmt.Printf("Replica3: %v\n", replica3.ToSlice())

	// All should converge
	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) ||
		!reflect.DeepEqual(replica2.ToSlice(), replica3.ToSlice()) {
		t.Errorf("Replicas did not converge!\nReplica1: %v\nReplica2: %v\nReplica3: %v",
			replica1.ToSlice(), replica2.ToSlice(), replica3.ToSlice())
	}
}

// TestMoveAfterAndBefore tests the MoveAfter and MoveBefore operations
func TestMoveAfterAndBefore(t *testing.T) {
	replica1 := New[string]("replica1")
	replica2 := New[string]("site2")

	// Setup
	idA := replica1.Push("A")
	idB := replica1.Push("B")
	_ = replica1.Push("C")
	idD := replica1.Push("D")

	replica2.Merge(replica1)

	// Concurrent operations using MoveAfter and MoveBefore
	// Replica1: Move D after A
	// Replica2: Move B before D
	replica1.MoveAfter(idD, idA)
	replica2.MoveBefore(idB, idD)

	fmt.Printf("After concurrent MoveAfter/MoveBefore:\n")
	fmt.Printf("Replica1 (D after A): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (B before D): %v\n", replica2.ToSlice())

	// Merge
	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter merge:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!\nReplica1: %v\nReplica2: %v",
			replica1.ToSlice(), replica2.ToSlice())
	}
}

// TestStressTestMoves performs many concurrent moves
func TestStressTestMoves(t *testing.T) {
	replica1 := New[int]("replica1")
	replica2 := New[int]("site2")

	// Add more items
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		ids[i] = replica1.Push(i)
	}

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Perform multiple concurrent moves
	// Replica1: Move even numbers to beginning
	// Replica2: Move odd numbers to end
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			replica1.Move(ids[i], i/2)
		} else {
			replica2.Move(ids[i], 9-i/2)
		}
	}

	fmt.Printf("\nAfter concurrent moves:\n")
	fmt.Printf("Replica1 (evens to start): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (odds to end): %v\n", replica2.ToSlice())

	// Merge
	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter merge:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}

// TestSwapOperation tests the swap functionality
func TestSwapOperation(t *testing.T) {
	replica1 := New[string]("replica1")
	replica2 := New[string]("site2")

	idA := replica1.Push("A")
	idB := replica1.Push("B")
	idC := replica1.Push("C")
	idD := replica1.Push("D")

	replica2.Merge(replica1)

	fmt.Printf("Initial: %v\n", replica1.ToSlice())

	// Concurrent swaps
	// Replica1: Swap A and D
	// Replica2: Swap B and C
	replica1.Swap(idA, idD)
	replica2.Swap(idB, idC)

	fmt.Printf("\nAfter concurrent swaps:\n")
	fmt.Printf("Replica1 (swapped A-D): %v\n", replica1.ToSlice())
	fmt.Printf("Replica2 (swapped B-C): %v\n", replica2.ToSlice())

	// Merge
	replica1.Merge(replica2)
	replica2.Merge(replica1)

	fmt.Printf("\nAfter merge:\n")
	fmt.Printf("Replica1: %v\n", replica1.ToSlice())
	fmt.Printf("Replica2: %v\n", replica2.ToSlice())

	if !reflect.DeepEqual(replica1.ToSlice(), replica2.ToSlice()) {
		t.Errorf("Replicas did not converge!")
	}
}
