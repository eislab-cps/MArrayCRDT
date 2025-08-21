package marraycrdt

import (
	"fmt"
)

func ExampleShoppingList() {
	// Shopping list for two users
	user1 := New[string]("user1")
	user2 := New[string]("user2")

	// User1 creates initial list
	_ = user1.Push("Milk")
	eggsID := user1.Push("Eggs")
	breadID := user1.Push("Bread")

	// Sync to user2
	user2.Merge(user1)

	fmt.Println("Initial list:", user1.ToSlice())

	// Concurrent operations:
	// User1: Realizes they need 2 dozen eggs, edits the item
	user1.Set(eggsID, "Eggs (2 dozen)")

	// User2: Reorganizes list, moves bread to top (more important)
	user2.Move(breadID, 0)

	// Both sync
	user1.Merge(user2)
	user2.Merge(user1)

	fmt.Println("After sync:", user1.ToSlice())
	// Both operations applied: Bread moved to top, Eggs edited

	// Output:
	// Initial list: [Milk Eggs Bread]
	// After sync: [Bread Milk Eggs (2 dozen)]
}

func ExampleTaskList() {
	type Task struct {
		Title    string
		Priority int
	}

	// Two team members managing tasks
	alice := New[Task]("alice")
	bob := New[Task]("bob")

	// Alice creates tasks
	task1 := alice.Push(Task{"Write documentation", 2})
	task2 := alice.Push(Task{"Fix bug #123", 1})
	_ = alice.Push(Task{"Review PR", 3})

	// Sync to Bob
	bob.Merge(alice)

	// Concurrent actions:
	// Alice: Updates priority of documentation
	alice.Set(task1, Task{"Write documentation", 1})

	// Bob: Moves bug fix to end (completed)
	bob.Move(task2, 2)

	// Sync
	alice.Merge(bob)
	bob.Merge(alice)

	fmt.Println("Alice's view:", alice.ToSlice())
	fmt.Println("Bob's view:", bob.ToSlice())

	// Both see same result with both changes applied
}
