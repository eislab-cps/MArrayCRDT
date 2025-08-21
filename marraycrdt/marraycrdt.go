// Package marraycrdt provides a Movable Array CRDT with full array operation support
package marraycrdt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	mathrand "math/rand"
	"sort"
	"sync"
	"time"
)

// MArrayCRDT is a Movable Array CRDT that supports full array operations
// including move, sort, reverse, and more while maintaining convergence
type MArrayCRDT[T any] struct {
	mu     sync.RWMutex
	items  map[string]*Element[T]
	siteID string
	clock  *VectorClock
	config Config

	// Cache for performance
	sortedCache []*Element[T]
	cacheValid  bool
}

// Element represents a single element in the array
type Element[T any] struct {
	ID          string
	Value       *VersionedValue[T]
	Index       *VersionedIndex
	VectorClock *VectorClock
	Deleted     bool
	DeleteClock *VectorClock
}

// VersionedValue tracks value changes independently
type VersionedValue[T any] struct {
	Data        T
	VectorClock *VectorClock
}

// VersionedIndex tracks position changes independently
type VersionedIndex struct {
	Position    float64
	VectorClock *VectorClock
}

// Config holds configuration options
type Config struct {
	AutoReindex      bool
	ReindexThreshold float64
	InitialIndex     float64
	IndexSpacing     float64
	KeepSorted       bool
	LessFunc         func(a, b interface{}) bool
}

// VectorClock implementation for causality tracking
type VectorClock struct {
	mu     sync.RWMutex
	clocks map[string]uint64
}

// Option is a configuration option
type Option func(*Config)

// NewVectorClock creates a new vector clock
func NewVectorClock() *VectorClock {
	return &VectorClock{
		clocks: make(map[string]uint64),
	}
}

// Increment increments the clock for a site
func (vc *VectorClock) Increment(siteID string) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.clocks[siteID]++
}

// Merge merges another vector clock into this one
func (vc *VectorClock) Merge(other *VectorClock) {
	if other == nil {
		return
	}

	vc.mu.Lock()
	other.mu.RLock()
	defer vc.mu.Unlock()
	defer other.mu.RUnlock()

	for site, clock := range other.clocks {
		if clock > vc.clocks[site] {
			vc.clocks[site] = clock
		}
	}
}

// After returns true if this clock is causally after other
func (vc *VectorClock) After(other *VectorClock) bool {
	if other == nil {
		return true
	}

	vc.mu.RLock()
	other.mu.RLock()
	defer vc.mu.RUnlock()
	defer other.mu.RUnlock()

	hasGreater := false
	for site, clock := range vc.clocks {
		if clock < other.clocks[site] {
			return false
		}
		if clock > other.clocks[site] {
			hasGreater = true
		}
	}

	for site, clock := range other.clocks {
		if _, exists := vc.clocks[site]; !exists && clock > 0 {
			return false
		}
	}

	return hasGreater
}

// Concurrent returns true if clocks are concurrent
func (vc *VectorClock) Concurrent(other *VectorClock) bool {
	return !vc.After(other) && !other.After(vc)
}

// Clone creates a copy of the vector clock
func (vc *VectorClock) Clone() *VectorClock {
	if vc == nil {
		return nil
	}
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	newVC := NewVectorClock()
	for site, clock := range vc.clocks {
		newVC.clocks[site] = clock
	}
	return newVC
}

// Fork creates a copy for independent tracking
func (vc *VectorClock) Fork() *VectorClock {
	return vc.Clone()
}

// GetMaxSite returns the site with highest ID for tiebreaking
func (vc *VectorClock) GetMaxSite() string {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	maxSite := ""
	for site := range vc.clocks {
		if site > maxSite {
			maxSite = site
		}
	}
	return maxSite
}

// defaultConfig returns default configuration
func defaultConfig() Config {
	return Config{
		AutoReindex:      true,
		ReindexThreshold: 0.0001,
		InitialIndex:     1000.0,
		IndexSpacing:     1000.0,
		KeepSorted:       false,
	}
}

// WithAutoReindex enables automatic reindexing
func WithAutoReindex(threshold float64) Option {
	return func(c *Config) {
		c.AutoReindex = true
		c.ReindexThreshold = threshold
	}
}

// WithAutoSort keeps array automatically sorted
func WithAutoSort[T any](less func(a, b T) bool) Option {
	return func(c *Config) {
		c.KeepSorted = true
		c.LessFunc = func(a, b interface{}) bool {
			return less(a.(T), b.(T))
		}
	}
}

// New creates a new MArrayCRDT
func New[T any](siteID string, opts ...Option) *MArrayCRDT[T] {
	config := defaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return &MArrayCRDT[T]{
		items:  make(map[string]*Element[T]),
		siteID: siteID,
		clock:  NewVectorClock(),
		config: config,
	}
}

// generateUUID generates a unique identifier
func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Clone creates a deep copy of an element
func (e *Element[T]) Clone() *Element[T] {
	return &Element[T]{
		ID: e.ID,
		Value: &VersionedValue[T]{
			Data:        e.Value.Data,
			VectorClock: e.Value.VectorClock.Clone(),
		},
		Index: &VersionedIndex{
			Position:    e.Index.Position,
			VectorClock: e.Index.VectorClock.Clone(),
		},
		VectorClock: e.VectorClock.Clone(),
		Deleted:     e.Deleted,
		DeleteClock: e.DeleteClock.Clone(),
	}
}

// Push adds element to end
func (ma *MArrayCRDT[T]) Push(value T) string {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	id := generateUUID()
	maxIndex := ma.findMaxIndexLocked()

	elem := &Element[T]{
		ID: id,
		Value: &VersionedValue[T]{
			Data:        value,
			VectorClock: ma.clock.Fork(),
		},
		Index: &VersionedIndex{
			Position:    maxIndex + ma.config.IndexSpacing,
			VectorClock: ma.clock.Fork(),
		},
		VectorClock: ma.clock.Fork(),
	}

	ma.clock.Increment(ma.siteID)
	elem.Value.VectorClock.Increment(ma.siteID)
	elem.Index.VectorClock.Increment(ma.siteID)
	elem.VectorClock.Increment(ma.siteID)

	ma.items[id] = elem
	ma.invalidateCache()

	if ma.config.KeepSorted {
		ma.maintainSortLocked()
	}

	return id
}

// Pop removes and returns last element
func (ma *MArrayCRDT[T]) Pop() (T, bool) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	sorted := ma.getSortedElementsLocked()
	if len(sorted) == 0 {
		var zero T
		return zero, false
	}

	last := sorted[len(sorted)-1]
	ma.deleteElementLocked(last.ID)

	return last.Value.Data, true
}

// Shift removes and returns first element
func (ma *MArrayCRDT[T]) Shift() (T, bool) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	sorted := ma.getSortedElementsLocked()
	if len(sorted) == 0 {
		var zero T
		return zero, false
	}

	first := sorted[0]
	ma.deleteElementLocked(first.ID)

	return first.Value.Data, true
}

// Unshift adds element to beginning
func (ma *MArrayCRDT[T]) Unshift(value T) string {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	id := generateUUID()
	minIndex := ma.findMinIndexLocked()

	elem := &Element[T]{
		ID: id,
		Value: &VersionedValue[T]{
			Data:        value,
			VectorClock: ma.clock.Fork(),
		},
		Index: &VersionedIndex{
			Position:    minIndex - ma.config.IndexSpacing,
			VectorClock: ma.clock.Fork(),
		},
		VectorClock: ma.clock.Fork(),
	}

	ma.clock.Increment(ma.siteID)
	elem.Value.VectorClock.Increment(ma.siteID)
	elem.Index.VectorClock.Increment(ma.siteID)
	elem.VectorClock.Increment(ma.siteID)

	ma.items[id] = elem
	ma.invalidateCache()

	if ma.config.KeepSorted {
		ma.maintainSortLocked()
	}

	return id
}

// Get returns element at index
func (ma *MArrayCRDT[T]) Get(index int) (T, bool) {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	sorted := ma.getSortedElementsLocked()
	if index < 0 || index >= len(sorted) {
		var zero T
		return zero, false
	}

	return sorted[index].Value.Data, true
}

// Set updates value of element
func (ma *MArrayCRDT[T]) Set(id string, value T) bool {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elem, exists := ma.items[id]
	if !exists || elem.Deleted {
		return false
	}

	ma.clock.Increment(ma.siteID)
	elem.Value.Data = value
	elem.Value.VectorClock = ma.clock.Fork()
	elem.Value.VectorClock.Increment(ma.siteID)
	elem.VectorClock.Merge(elem.Value.VectorClock)

	return true
}

// Insert adds element at specific index
func (ma *MArrayCRDT[T]) Insert(index int, value T) string {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	sorted := ma.getSortedElementsLocked()
	id := generateUUID()

	var position float64
	if index <= 0 {
		minIndex := ma.findMinIndexLocked()
		position = minIndex - ma.config.IndexSpacing
	} else if index >= len(sorted) {
		maxIndex := ma.findMaxIndexLocked()
		position = maxIndex + ma.config.IndexSpacing
	} else {
		// Insert between elements
		if index == 0 {
			position = sorted[0].Index.Position - ma.config.IndexSpacing
		} else {
			prev := sorted[index-1]
			next := sorted[index]
			position = (prev.Index.Position + next.Index.Position) / 2
		}
	}

	elem := &Element[T]{
		ID: id,
		Value: &VersionedValue[T]{
			Data:        value,
			VectorClock: ma.clock.Fork(),
		},
		Index: &VersionedIndex{
			Position:    position,
			VectorClock: ma.clock.Fork(),
		},
		VectorClock: ma.clock.Fork(),
	}

	ma.clock.Increment(ma.siteID)
	elem.Value.VectorClock.Increment(ma.siteID)
	elem.Index.VectorClock.Increment(ma.siteID)
	elem.VectorClock.Increment(ma.siteID)

	ma.items[id] = elem
	ma.invalidateCache()

	if ma.config.AutoReindex {
		ma.checkReindexLocked()
	}

	if ma.config.KeepSorted {
		ma.maintainSortLocked()
	}

	return id
}

// Delete removes element by ID
func (ma *MArrayCRDT[T]) Delete(id string) bool {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	return ma.deleteElementLocked(id)
}

// deleteElementLocked deletes element (must hold lock)
func (ma *MArrayCRDT[T]) deleteElementLocked(id string) bool {
	elem, exists := ma.items[id]
	if !exists || elem.Deleted {
		return false
	}

	ma.clock.Increment(ma.siteID)
	elem.Deleted = true
	elem.DeleteClock = ma.clock.Fork()
	elem.DeleteClock.Increment(ma.siteID)
	elem.VectorClock.Merge(elem.DeleteClock)

	ma.invalidateCache()
	return true
}

// Move element to specific position
func (ma *MArrayCRDT[T]) Move(id string, toIndex int) bool {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elem, exists := ma.items[id]
	if !exists {
		return false
	}

	// IMPORTANT: Moving a deleted item resurrects it with LWW semantics
	if elem.Deleted {
		elem.Deleted = false
		elem.DeleteClock = nil
	}

	sorted := ma.getSortedElementsLocked()

	// Adjust index bounds
	if toIndex < 0 {
		toIndex = 0
	}
	if toIndex >= len(sorted) {
		toIndex = len(sorted) - 1
	}

	var newPos float64
	if toIndex == 0 {
		newPos = sorted[0].Index.Position - ma.config.IndexSpacing
	} else if toIndex >= len(sorted)-1 {
		newPos = sorted[len(sorted)-1].Index.Position + ma.config.IndexSpacing
	} else {
		// Find the target position between elements
		// Account for current element position
		targetElements := make([]*Element[T], 0)
		for _, e := range sorted {
			if e.ID != id {
				targetElements = append(targetElements, e)
			}
		}

		if toIndex > 0 && toIndex <= len(targetElements) {
			prev := targetElements[toIndex-1]
			next := targetElements[toIndex]
			newPos = (prev.Index.Position + next.Index.Position) / 2
		} else {
			newPos = targetElements[toIndex].Index.Position - ma.config.IndexSpacing
		}
	}

	ma.clock.Increment(ma.siteID)
	elem.Index.Position = newPos
	elem.Index.VectorClock = ma.clock.Fork()
	elem.Index.VectorClock.Increment(ma.siteID)
	elem.VectorClock.Merge(elem.Index.VectorClock)

	ma.invalidateCache()

	if ma.config.AutoReindex {
		ma.checkReindexLocked()
	}

	return true
}

// MoveAfter moves element after another element
func (ma *MArrayCRDT[T]) MoveAfter(id string, afterID string) bool {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elem, exists := ma.items[id]
	if !exists {
		return false
	}

	after, exists := ma.items[afterID]
	if !exists || after.Deleted {
		return false
	}

	// Resurrect if deleted
	if elem.Deleted {
		elem.Deleted = false
		elem.DeleteClock = nil
	}

	// Find next element after target
	sorted := ma.getSortedElementsLocked()
	var next *Element[T]
	foundAfter := false

	for _, e := range sorted {
		if foundAfter && e.ID != id {
			next = e
			break
		}
		if e.ID == afterID {
			foundAfter = true
		}
	}

	var newPos float64
	if next != nil {
		newPos = (after.Index.Position + next.Index.Position) / 2
	} else {
		newPos = after.Index.Position + ma.config.IndexSpacing
	}

	ma.clock.Increment(ma.siteID)
	elem.Index.Position = newPos
	elem.Index.VectorClock = ma.clock.Fork()
	elem.Index.VectorClock.Increment(ma.siteID)
	elem.VectorClock.Merge(elem.Index.VectorClock)

	ma.invalidateCache()

	if ma.config.AutoReindex {
		ma.checkReindexLocked()
	}

	return true
}

// MoveBefore moves element before another element
func (ma *MArrayCRDT[T]) MoveBefore(id string, beforeID string) bool {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elem, exists := ma.items[id]
	if !exists {
		return false
	}

	before, exists := ma.items[beforeID]
	if !exists || before.Deleted {
		return false
	}

	// Resurrect if deleted
	if elem.Deleted {
		elem.Deleted = false
		elem.DeleteClock = nil
	}

	// Find previous element before target
	sorted := ma.getSortedElementsLocked()
	var prev *Element[T]

	for _, e := range sorted {
		if e.ID == beforeID {
			break
		}
		if e.ID != id {
			prev = e
		}
	}

	var newPos float64
	if prev != nil {
		newPos = (prev.Index.Position + before.Index.Position) / 2
	} else {
		newPos = before.Index.Position - ma.config.IndexSpacing
	}

	ma.clock.Increment(ma.siteID)
	elem.Index.Position = newPos
	elem.Index.VectorClock = ma.clock.Fork()
	elem.Index.VectorClock.Increment(ma.siteID)
	elem.VectorClock.Merge(elem.Index.VectorClock)

	ma.invalidateCache()

	if ma.config.AutoReindex {
		ma.checkReindexLocked()
	}

	return true
}

// Sort array with custom comparison
func (ma *MArrayCRDT[T]) Sort(less func(a, b T) bool) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elements := ma.getSortedElementsLocked()
	if len(elements) == 0 {
		return
	}

	// Sort by value
	sort.Slice(elements, func(i, j int) bool {
		return less(elements[i].Value.Data, elements[j].Value.Data)
	})

	// Update indices
	ma.clock.Increment(ma.siteID)

	for i, elem := range elements {
		elem.Index.Position = float64(i+1) * ma.config.IndexSpacing
		// Give each element a unique clock
		elem.Index.VectorClock = ma.clock.Fork()
		elem.Index.VectorClock.Increment(ma.siteID)
		elem.VectorClock.Merge(elem.Index.VectorClock)
		ma.clock.Increment(ma.siteID)
	}

	ma.invalidateCache()
}

// Reverse reverses the array order
func (ma *MArrayCRDT[T]) Reverse() {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elements := ma.getSortedElementsLocked()
	n := len(elements)
	if n == 0 {
		return
	}

	ma.clock.Increment(ma.siteID)

	for i, elem := range elements {
		elem.Index.Position = float64(n-i) * ma.config.IndexSpacing
		// Give each element a unique clock by incrementing for each one
		elem.Index.VectorClock = ma.clock.Fork()
		elem.Index.VectorClock.Increment(ma.siteID)
		elem.VectorClock.Merge(elem.Index.VectorClock)
		// Increment main clock for next element
		ma.clock.Increment(ma.siteID)
	}

	ma.invalidateCache()
}

// Shuffle randomizes array order
func (ma *MArrayCRDT[T]) Shuffle() {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elements := ma.getSortedElementsLocked()
	if len(elements) == 0 {
		return
	}

	// Generate random positions
	indices := make([]float64, len(elements))
	for i := range indices {
		indices[i] = float64(i+1) * ma.config.IndexSpacing
	}

	// Shuffle positions
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	ma.clock.Increment(ma.siteID)

	for i, elem := range elements {
		elem.Index.Position = indices[i]
		// Give each element a unique clock
		elem.Index.VectorClock = ma.clock.Fork()
		elem.Index.VectorClock.Increment(ma.siteID)
		elem.VectorClock.Merge(elem.Index.VectorClock)
		ma.clock.Increment(ma.siteID)
	}

	ma.invalidateCache()
}

// Rotate rotates array by n positions
func (ma *MArrayCRDT[T]) Rotate(n int) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elements := ma.getSortedElementsLocked()
	length := len(elements)
	if length == 0 {
		return
	}

	// Normalize rotation
	n = n % length
	if n < 0 {
		n += length
	}

	ma.clock.Increment(ma.siteID)

	for i, elem := range elements {
		newPos := (i + n) % length
		elem.Index.Position = float64(newPos+1) * ma.config.IndexSpacing
		// Give each element a unique clock
		elem.Index.VectorClock = ma.clock.Fork()
		elem.Index.VectorClock.Increment(ma.siteID)
		elem.VectorClock.Merge(elem.Index.VectorClock)
		ma.clock.Increment(ma.siteID)
	}

	ma.invalidateCache()
}

// Swap swaps two elements
func (ma *MArrayCRDT[T]) Swap(id1, id2 string) bool {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	elem1, exists1 := ma.items[id1]
	elem2, exists2 := ma.items[id2]

	if !exists1 || !exists2 || elem1.Deleted || elem2.Deleted {
		return false
	}

	ma.clock.Increment(ma.siteID)

	// Swap positions
	elem1.Index.Position, elem2.Index.Position = elem2.Index.Position, elem1.Index.Position

	// Give each element a unique clock
	elem1.Index.VectorClock = ma.clock.Fork()
	elem1.Index.VectorClock.Increment(ma.siteID)
	elem1.VectorClock.Merge(elem1.Index.VectorClock)
	
	ma.clock.Increment(ma.siteID)
	
	elem2.Index.VectorClock = ma.clock.Fork()
	elem2.Index.VectorClock.Increment(ma.siteID)
	elem2.VectorClock.Merge(elem2.Index.VectorClock)

	ma.invalidateCache()
	return true
}

// Merge merges another MArrayCRDT into this one
func (ma *MArrayCRDT[T]) Merge(other *MArrayCRDT[T]) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	for id, remoteElem := range other.items {
		localElem, exists := ma.items[id]

		if !exists {
			// New element - just copy it
			ma.items[id] = remoteElem.Clone()
			ma.clock.Merge(remoteElem.VectorClock)
			ma.invalidateCache()
			continue
		}

		// FIXED: Properly handle delete vs move/edit conflicts with LWW
		ma.mergeElementWithLWW(localElem, remoteElem)

		// Update overall clock
		localElem.VectorClock.Merge(remoteElem.VectorClock)
		ma.clock.Merge(remoteElem.VectorClock)
	}

	if ma.config.KeepSorted {
		ma.maintainSortLocked()
	}
}

// mergeElementWithLWW merges elements using Last-Writer-Wins semantics
func (ma *MArrayCRDT[T]) mergeElementWithLWW(local, remote *Element[T]) {
	// First, merge Value (edit) operations independently
	if remote.Value.VectorClock.After(local.Value.VectorClock) {
		local.Value = &VersionedValue[T]{
			Data:        remote.Value.Data,
			VectorClock: remote.Value.VectorClock.Clone(),
		}
	} else if local.Value.VectorClock.Concurrent(remote.Value.VectorClock) {
		if remote.Value.VectorClock.GetMaxSite() > local.Value.VectorClock.GetMaxSite() {
			local.Value = &VersionedValue[T]{
				Data:        remote.Value.Data,
				VectorClock: remote.Value.VectorClock.Clone(),
			}
		}
	}

	// Second, merge Index (move) operations independently
	if remote.Index.VectorClock.After(local.Index.VectorClock) {
		local.Index = &VersionedIndex{
			Position:    remote.Index.Position,
			VectorClock: remote.Index.VectorClock.Clone(),
		}
		ma.invalidateCache()
	} else if local.Index.VectorClock.Concurrent(remote.Index.VectorClock) {
		if remote.Index.VectorClock.GetMaxSite() > local.Index.VectorClock.GetMaxSite() {
			local.Index = &VersionedIndex{
				Position:    remote.Index.Position,
				VectorClock: remote.Index.VectorClock.Clone(),
			}
			ma.invalidateCache()
		}
	}

	// CRITICAL FIX: Resolve delete status using LWW between ALL operations
	local.Deleted = ma.resolveDeleteStatusLWW(local, remote)

	// Update delete clock if needed
	if local.Deleted {
		if remote.Deleted && remote.DeleteClock != nil {
			if local.DeleteClock == nil || remote.DeleteClock.After(local.DeleteClock) {
				local.DeleteClock = remote.DeleteClock.Clone()
			}
		}
	} else {
		// Item is alive - clear delete clock
		local.DeleteClock = nil
	}
}

// resolveDeleteStatusLWW determines delete status using Last-Writer-Wins
func (ma *MArrayCRDT[T]) resolveDeleteStatusLWW(local, remote *Element[T]) bool {
	// Collect all relevant timestamps
	type Operation struct {
		Clock    *VectorClock
		IsDelete bool
		Source   string
	}

	operations := []Operation{}

	// Add ALL delete operations (not just if currently deleted)
	if local.DeleteClock != nil {
		operations = append(operations, Operation{
			Clock:    local.DeleteClock,
			IsDelete: true,
			Source:   "local-delete",
		})
	}

	if remote.DeleteClock != nil {
		operations = append(operations, Operation{
			Clock:    remote.DeleteClock,
			IsDelete: true,
			Source:   "remote-delete",
		})
	}

	// Add move operations as "undelete" operations
	// Local move
	if local.Index != nil && local.Index.VectorClock != nil {
		operations = append(operations, Operation{
			Clock:    local.Index.VectorClock,
			IsDelete: false,
			Source:   "local-move",
		})
	}

	// Remote move
	if remote.Index != nil && remote.Index.VectorClock != nil {
		operations = append(operations, Operation{
			Clock:    remote.Index.VectorClock,
			IsDelete: false,
			Source:   "remote-move",
		})
	}

	// Find the operation with the latest timestamp
	var latestOp *Operation
	for i := range operations {
		op := &operations[i]
		if latestOp == nil {
			latestOp = op
			continue
		}

		if op.Clock.After(latestOp.Clock) {
			latestOp = op
		} else if latestOp.Clock.After(op.Clock) {
			// Keep current latest
		} else {
			// Concurrent - use tiebreaker
			// For concurrent operations, prefer the deterministic site ID ordering
			if op.Clock.GetMaxSite() > latestOp.Clock.GetMaxSite() {
				latestOp = op
			}
		}
	}

	// Return whether the latest operation was a delete
	if latestOp != nil {
		return latestOp.IsDelete
	}

	// No operations found - default to not deleted
	return false
}

// Clone creates a deep copy of the array
func (ma *MArrayCRDT[T]) Clone() *MArrayCRDT[T] {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	newArray := &MArrayCRDT[T]{
		items:  make(map[string]*Element[T]),
		siteID: ma.siteID,
		clock:  ma.clock.Clone(),
		config: ma.config,
	}

	for id, elem := range ma.items {
		newArray.items[id] = elem.Clone()
	}

	return newArray
}

// Len returns the number of non-deleted elements
func (ma *MArrayCRDT[T]) Len() int {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	count := 0
	for _, elem := range ma.items {
		if !elem.Deleted {
			count++
		}
	}
	return count
}

// ToSlice returns array as slice
func (ma *MArrayCRDT[T]) ToSlice() []T {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	sorted := ma.getSortedElementsLocked()
	result := make([]T, 0, len(sorted))

	for _, elem := range sorted {
		result = append(result, elem.Value.Data)
	}

	return result
}

// IDs returns all element IDs in order
func (ma *MArrayCRDT[T]) IDs() []string {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	sorted := ma.getSortedElementsLocked()
	result := make([]string, 0, len(sorted))

	for _, elem := range sorted {
		result = append(result, elem.ID)
	}

	return result
}

// Clear removes all elements
func (ma *MArrayCRDT[T]) Clear() {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	ma.clock.Increment(ma.siteID)
	clock := ma.clock.Fork()
	clock.Increment(ma.siteID)

	for _, elem := range ma.items {
		if !elem.Deleted {
			elem.Deleted = true
			elem.DeleteClock = clock.Clone()
			elem.VectorClock.Merge(clock)
		}
	}

	ma.invalidateCache()
}

// Helper methods (internal, must hold lock)

func (ma *MArrayCRDT[T]) getSortedElementsLocked() []*Element[T] {
	if ma.cacheValid {
		return ma.sortedCache
	}

	elements := make([]*Element[T], 0, len(ma.items))
	for _, elem := range ma.items {
		if !elem.Deleted {
			elements = append(elements, elem)
		}
	}

	sort.Slice(elements, func(i, j int) bool {
		// First compare by position
		if elements[i].Index.Position != elements[j].Index.Position {
			return elements[i].Index.Position < elements[j].Index.Position
		}
		// If positions are equal, use UUID as tiebreaker for deterministic ordering
		return elements[i].ID < elements[j].ID
	})

	ma.sortedCache = elements
	ma.cacheValid = true

	return elements
}

func (ma *MArrayCRDT[T]) invalidateCache() {
	ma.cacheValid = false
}

func (ma *MArrayCRDT[T]) findMaxIndexLocked() float64 {
	if len(ma.items) == 0 {
		return ma.config.InitialIndex
	}

	maxIndex := -math.MaxFloat64
	for _, elem := range ma.items {
		if !elem.Deleted && elem.Index.Position > maxIndex {
			maxIndex = elem.Index.Position
		}
	}

	if maxIndex == -math.MaxFloat64 {
		return ma.config.InitialIndex
	}

	return maxIndex
}

func (ma *MArrayCRDT[T]) findMinIndexLocked() float64 {
	if len(ma.items) == 0 {
		return ma.config.InitialIndex
	}

	minIndex := math.MaxFloat64
	for _, elem := range ma.items {
		if !elem.Deleted && elem.Index.Position < minIndex {
			minIndex = elem.Index.Position
		}
	}

	if minIndex == math.MaxFloat64 {
		return ma.config.InitialIndex
	}

	return minIndex
}

func (ma *MArrayCRDT[T]) checkReindexLocked() {
	if !ma.config.AutoReindex {
		return
	}

	sorted := ma.getSortedElementsLocked()
	if len(sorted) < 2 {
		return
	}

	needsReindex := false
	for i := 1; i < len(sorted); i++ {
		diff := sorted[i].Index.Position - sorted[i-1].Index.Position
		if diff < ma.config.ReindexThreshold {
			needsReindex = true
			break
		}
	}

	if needsReindex {
		ma.reindexLocked()
	}
}

func (ma *MArrayCRDT[T]) reindexLocked() {
	sorted := ma.getSortedElementsLocked()

	ma.clock.Increment(ma.siteID)
	clock := ma.clock.Fork()
	clock.Increment(ma.siteID)

	for i, elem := range sorted {
		elem.Index.Position = float64(i+1) * ma.config.IndexSpacing
		elem.Index.VectorClock = clock.Clone()
		elem.VectorClock.Merge(clock)
	}

	ma.invalidateCache()
}

func (ma *MArrayCRDT[T]) maintainSortLocked() {
	if !ma.config.KeepSorted || ma.config.LessFunc == nil {
		return
	}

	elements := ma.getSortedElementsLocked()

	sort.Slice(elements, func(i, j int) bool {
		return ma.config.LessFunc(elements[i].Value.Data, elements[j].Value.Data)
	})

	ma.clock.Increment(ma.siteID)
	clock := ma.clock.Fork()
	clock.Increment(ma.siteID)

	for i, elem := range elements {
		elem.Index.Position = float64(i+1) * ma.config.IndexSpacing
		elem.Index.VectorClock = clock.Clone()
		elem.VectorClock.Merge(clock)
	}

	ma.invalidateCache()
}

// GetElement returns the full element by ID (for debugging)
func (ma *MArrayCRDT[T]) GetElement(id string) (*Element[T], bool) {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	elem, exists := ma.items[id]
	if !exists || elem.Deleted {
		return nil, false
	}

	return elem.Clone(), true
}

// String returns a string representation
func (ma *MArrayCRDT[T]) String() string {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	sorted := ma.getSortedElementsLocked()
	values := make([]string, 0, len(sorted))

	for _, elem := range sorted {
		values = append(values, fmt.Sprintf("%v", elem.Value.Data))
	}

	return fmt.Sprintf("[%v]", values)
}
