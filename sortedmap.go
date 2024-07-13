/*
Copyright © 2023, 2024  M.Watermann, 10247 Berlin, Germany

		All rights reserved
	EMail : <support@mwat.de>
*/
package sortedlists

import (
	"cmp"
	"fmt"
	"slices"
	"sort"
	"sync"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

// `TSortedMap` is a generic type that accepts two type parameters:
// - K for the key type (which must be cmp.Ordered)
// - V for the value type
//
// The `cmp.Ordered` interface is defined as:
//
//	~int | ~int8 | ~int16 | ~int32 | ~int64 |
//		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
//		~float32 | ~float64 |
//		~string
//
// `comparable` is an interface that is implemented by all comparable
// types (booleans, numbers, strings, pointers, channels, arrays of
// comparable types, structs whose fields are all comparable types).
//
// All methods are optionally thread-safe and can be called concurrently.
type TSortedMap[K cmp.Ordered, V comparable] struct {
	data map[K]V
	keys []K
	mtx  sync.RWMutex
	safe bool
}

// --------------------------------------------------------------------------
// constructor function

// `NewSortedMap()` creates a new instance of `TSortedMap` with the
// specified key and value types.
//
// The returned map is initially empty.
//
// Parameters:
//   - `K`: The type of the keys in the sorted map.
//   - `V`: The type of the values in the sorted map.
//   - `aSafe`: Flag to decide whether the returned map should be
//     thread safe, i.e. use a `sync.RWMutex` in all methods.
//
// Returns:
// - `*TSortedMap[K, V]`: A pointer to a new instance with the given
// key and value types.
func NewSortedMap[K cmp.Ordered, V comparable](aSafe bool) *TSortedMap[K, V] {
	return &TSortedMap[K, V]{
		data: make(map[K]V),
		keys: make([]K, 0),
		safe: aSafe,
	}
} // NewSortedMap()

// --------------------------------------------------------------------------
// methods of TSortedMap

// `Clear()` empties the internal data structures:
// all map entrie are deleted.
//
// Returns:
// - `*TSortedMap`: The cleared hash map.
func (sm *TSortedMap[K, V]) Clear() *TSortedMap[K, V] {
	if sm.safe {
		sm.mtx.Lock()
		defer sm.mtx.Unlock()
	}

	sm.data = make(map[K]V)
	sm.keys = make([]K, 0)

	return sm
} // Clear()

func (sm *TSortedMap[K, V]) delete(aKey K) bool {
	// Check if the key actually exists
	if _, exists := sm.data[aKey]; exists {
		delete(sm.data, aKey)

		// Update the keys slice
		for idx, key := range sm.keys {
			if key == aKey {
				sm.keys = append(sm.keys[:idx], sm.keys[idx+1:]...)
				break
			}
		}
		return true
	}

	return false
} // delete()

// `Delete()` removes a key/value pair from the map.
//
// Parameters:
// - `aKey`: The key of the entry to be deleted.
//
// Returns:
// - `bool`: `true` if `aKey` was removed, or `false` otherwise.
func (sm *TSortedMap[K, V]) Delete(aKey K) bool {
	if sm.safe {
		sm.mtx.Lock()
		defer sm.mtx.Unlock()
	}

	return sm.delete(aKey)
} // Delete()

// - `[]K`: The index of `aID` in the list.
func (sm *TSortedMap[K, V]) findIndex(aValue V) []K {
	var result []K

	for _, key := range sm.keys {
		if sm.data[key] == aValue {
			result = append(result, key)
		}
	}

	return result
} // findIndex()

// `FindIndex()` returns a slice of keys that have the given value.
//
// Parameters:
// - `aValue`: The element to look up.
//
// Returns:
// - `[]K`: The index of `aID` in the list.
func (sm *TSortedMap[K, V]) FindIndex(aValue V) []K {
	if sm.safe {
		sm.mtx.RLock()
		defer sm.mtx.RUnlock()
	}

	return sm.findIndex(aValue)
} // FindIndex()

// `Get()` retrieves a value by its key from the SortedMap
//
// Parameters:
// - `aKey`: The key of the entry to be retrieved.
//
// Returns:
// - `V`: The value associated with the `aKey`.
// - `bool`: An indication whether the key was found in the map.
func (sm *TSortedMap[K, V]) Get(aKey K) (V, bool) {
	if sm.safe {
		sm.mtx.RLock()
		defer sm.mtx.RUnlock()
	}

	value, exists := sm.data[aKey]

	return value, exists
} // Get()

// Keys returns a slice of all keys in sorted order

// `Keys()` returns a slice of all keys in sorted order
//
// This method returns a slice of all keys in the map in sorted order.
//
// Returns:
// - `[]K`: A slice of keys in the sorted map.
func (sm *TSortedMap[K, V]) Keys() []K {
	if sm.safe {
		sm.mtx.RLock()
		defer sm.mtx.RUnlock()
	}

	return append([]K{}, sm.keys...)
} // Keys()

func (sm *TSortedMap[K, V]) insert(aKey K, aValue V) bool {
	if _, exists := sm.data[aKey]; exists {
		sm.data[aKey] = aValue

		return true
	}

	// There are different situations to consider:
	// 1: the key-list is empty,
	// 2: the key-list doesn't already contain the key,
	// 3: the key-list contains the key but with a different value
	sLen := len(sm.keys)
	if 0 == sLen {
		// 1: empty list: just add the new item
		sm.keys = append(sm.keys, aKey)
	} else {
		// find the insertion index using binary search
		idx := sort.Search(sLen, func(i int) bool {
			return sm.keys[i] >= aKey
		})

		if sLen == idx {
			// 2: key not found: add key at the end
			sm.keys = append(sm.keys, aKey)
		} else if (sm.keys)[idx] != aKey {
			// 3: the search index doesn't point to the required key
			sm.keys = append(sm.keys, aKey)
			copy((sm.keys)[idx+1:], (sm.keys)[idx:])
			(sm.keys)[idx] = aKey
		} else {
			// dummy instruction for debugger
			sLen = 0
		}
	}
	sm.data[aKey] = aValue

	return true
} // Insert()

// `Insert()` adds or updates a key/value pair in the sorted map.
//
// Parameters:
// - `aKey`: The key of the entry to be added or updated.
// - `aValue`: The value to be associated with the key.
//
// Returns:
// - `bool`: `true` if `aID` was inserted, or `false` otherwise.
func (sm *TSortedMap[K, V]) Insert(aKey K, aValue V) bool {
	if sm.safe {
		sm.mtx.Lock()
		defer sm.mtx.Unlock()
	}

	return sm.insert(aKey, aValue)
} // Insert()

// `Iterate()` allows iteration over the map in sorted key order.
//
// Parameters:
// - `f`: A function that takes a key and its associated value as
// arguments and performs some operation on them.
//
// Returns:
// - `*TSortedMap[K, V]`: A pointer to the same SortedMap instance,
// allowing method chaining.
func (sm *TSortedMap[K, V]) Iterate(f func(K, V)) *TSortedMap[K, V] {
	if sm.safe {
		sm.mtx.RLock()
		defer sm.mtx.RUnlock()
	}

	for _, key := range sm.keys {
		f(key, sm.data[key])
	}

	return sm
} // Iterate()

func (sm *TSortedMap[K, V]) Iterator() func() (K, V, bool) {
	var idx int

	return func() (K, V, bool) {
		var (
			key K
			val V
		)
		if idx < len(sm.keys) {
			key = sm.keys[idx]
			val = sm.data[key]
			idx++ // used from outer closure

			return key, val, true
		}

		return key, val, false
	}
} // Iterator()

func (sm *TSortedMap[K, V]) rename(aOldKey, aNewKey K) bool {
	// Check if the new key already exists
	if _, exists := sm.data[aNewKey]; exists {
		return false
	}

	// Check if the old key exists
	oldValue, exists := sm.data[aOldKey]
	if !exists || (aOldKey == aNewKey) {
		return false
	}

	// Remove the old key and add the new key
	delete(sm.data, aOldKey)
	sm.data[aNewKey] = oldValue

	// Update the keys slice
	for idx, key := range sm.keys {
		if key == aOldKey {
			sm.keys[idx] = aNewKey
			break
		}
	}

	// Re-sort the keys
	slices.Sort(sm.keys) // ascending

	return true
} // Rename()

// `Rename()` changes the key of an existing entry without affecting its value.
//
// If `aOldKey` equals `aNewKey`, or aOldKey` doesn't exist then they are
// silently ignored (i.e. this method does nothing), returning `false`.
//
// If `aNewKey` already exists, it is ignored and the method returns `false`.
//
// Parameters:
// - `aOldKey`: the key to be replaced in this list.
// - `aNewKey`: The replacement key in this list.
//
// Returns:
// - `bool`: `true` if the the renaming was successful, or `false` otherwise.
func (sm *TSortedMap[K, V]) Rename(aOldKey, aNewKey K) bool {
	if sm.safe {
		sm.mtx.Lock()
		defer sm.mtx.Unlock()
	}

	return sm.rename(aOldKey, aNewKey)
} // Rename()

func (sm *TSortedMap[K, V]) string() (rStr string) {
	// Access items in sorted order:
	iter := sm.Iterator()
	for key, value, hasNext := iter(); hasNext; key, value, hasNext = iter() {
		rStr += fmt.Sprintf("[%v]\n%v\n", key, value)
	}
	return
} // string()

func (sm *TSortedMap[K, V]) String() string {
	if sm.safe {
		sm.mtx.RLock()
		defer sm.mtx.RUnlock()
	}

	return sm.string()
} // String()

/* EoF */
