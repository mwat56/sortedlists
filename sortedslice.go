/*
Copyright ©  2024  M.Watermann, 10247 Berlin, Germany

		All rights reserved
	EMail : <support@mwat.de>
*/
package sortedlists

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"sync"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

type (
	// `TSortedSlice` represents a sorted slice of any ordered type.
	//
	// This is a generic type that accepts a type parameter:
	// - T for the ordered value type.
	//
	// The `Ordered` interface is defined as:
	// 	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	// 		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
	// 		~float32 | ~float64 |
	// 		~string
	//
	// All methods are optionally thread-safe and can be called concurrently.
	TSortedSlice[T cmp.Ordered] struct {
		data []T
		mtx  sync.RWMutex
		safe bool
	}
)

// --------------------------------------------------------------------------
// constructor function

// `NewSlice()` creates a new `TSortedSlice`.
//
// If the given `aList` is empty the initial capacity of the underlying
// list is set to 32 to optimise memory usage.
//
// Parameters:
// - `aList`: The slice to use with the sorted slice.
// - `aSafe`: Flag to decide whether the returned map should be
// thread safe, i.e. use a `sync.RWMutex` in all methods.
//
// Returns:
// - `*TSortedSlice[T]`: A pointer to the newly created instance.
func NewSlice[T cmp.Ordered](aList []T, aSafe bool) *TSortedSlice[T] {
	var list []T

	if 0 < len(aList) {
		list = aList
	} else {
		list = make([]T, 0, 32)
	}

	ss := &TSortedSlice[T]{
		data: list,
		safe: aSafe,
	}
	slices.Sort(ss.data)

	return ss
} // NewSlice()

// -------------------------------------------------------------------------
// methods of TSortedSlice

// `Clear()` removes all entries in this list.
//
// Returns:
// - `*TSortedSlice[T]`: The cleared list instance.
func (ss *TSortedSlice[T]) Clear() *TSortedSlice[T] {
	if ss.safe {
		ss.mtx.Lock()
		defer ss.mtx.Unlock()
	}

	ss.data = make([]T, 0, 32)

	return ss
} // Clear()

func (ss *TSortedSlice[T]) delete(aElement T) bool {
	sLen := len(ss.data)
	if 0 == sLen { // empty list
		return false
	}

	// Find the index of the given element in the sorted slice
	idx, ok := slices.BinarySearch(ss.data, aElement)
	if !ok {
		return false
	}

	if (idx < sLen) && (ss.data[idx] == aElement) {
		// `aElement` found at index `idx`
		if 0 == idx {
			if 1 == sLen { // the only element
				ss.data = make([]T, 0, 32)
			} else { // a longer list
				ss.data = ss.data[1:] // remove the first element
			}
		} else if (sLen - 1) == idx { // remove the last element
			ss.data = ss.data[:idx]
		} else { // remove element in the middle
			ss.data = append(ss.data[:idx], ss.data[idx+1:]...)
		}
		return true
	}

	return false
} // delete()

// `Delete()` removes an element from the sorted slice.
//
// Parameters:
// - `aElement`: The element to remove from the list.
//
// Returns:
// - `bool`: `true` if `aElement` was removed, or `false` otherwise.
func (ss *TSortedSlice[T]) Delete(aElement T) bool {
	if ss.safe {
		ss.mtx.Lock()
		defer ss.mtx.Unlock()
	}

	return ss.delete(aElement)
} // Delete()

// `Data()` returns the underlying data of the sorted slice.
//
// Returns:
// - `[]T`: The underlying data of the sorted slice.
func (ss *TSortedSlice[T]) Data() []T {
	if ss.safe {
		ss.mtx.RLock()
		defer ss.mtx.RUnlock()
	}

	return append([]T{}, ss.data...)
} // Data()

func (ss *TSortedSlice[T]) findIndex(aElement T) int {
	sLen := len(ss.data)
	if 0 == sLen {
		return -1
	}

	// Find the index of the given element
	idx, exists := slices.BinarySearch(ss.data, aElement)
	if !exists {
		return -1
	}

	if idx < sLen && ss.data[idx] == aElement {
		return idx
	}

	return -1 // aElement not found
} // findIndex()

// `FindIndex()` returns the list index of `aElement`.
//
// If the `aElement` is not found, the method returns -1.
//
// Parameters:
// - `aElement`: The list element to look up.
//
// Returns:
// - `int`: The index of `aID` in the list.
func (ss *TSortedSlice[T]) FindIndex(aElement T) int {
	if ss.safe {
		ss.mtx.RLock()
		defer ss.mtx.RUnlock()
	}

	return ss.findIndex(aElement)
} // FindIndex()

// `Get()` retrieves a value by its list index from the sorted slice.
//
// Parameters:
// - `aIndex`: The list index to use for returning the list element.
//
// Returns:
// - `T`: The value associated with the `aIndex`.
// - `bool`: An indication whether the index was found in the list.
func (ss *TSortedSlice[T]) Get(aIndex int) (T, bool) {
	if ss.safe {
		ss.mtx.RLock()
		defer ss.mtx.RUnlock()
	}
	var result T // variable with its zero value

	if aIndex < len(ss.data) {
		return ss.data[aIndex], true
	}

	return result, false
} // Get()

func (ss *TSortedSlice[T]) insert(aElement T) bool {
	sLen := len(ss.data)
	if 0 == sLen { // empty list
		ss.data = append(ss.data, aElement)

		return true
	}

	// find the insertion index using binary search
	idx, exists := slices.BinarySearch(ss.data, aElement)
	if !exists {
		ss.data = append(ss.data, aElement)

		return true
	}

	if sLen == idx { // key not found
		ss.data = append(ss.data, aElement) // add new element
		return true
	}
	if ss.data[idx] != aElement {
		ss.data = append(ss.data, aElement) // make room for new element
		copy(ss.data[idx+1:], ss.data[idx:])
		ss.data[idx] = aElement

		return true
	}

	return false
} // insert()

// `Insert()` adds an element to the sorted slice while maintaining order.
//
// Parameters:
// - `aElement` The element to insert to the list.
//
// Returns:
// - `bool`: `true` if `aElement` was inserted, or `false` otherwise.
func (ss *TSortedSlice[T]) Insert(aElement T) bool {
	if ss.safe {
		ss.mtx.Lock()
		defer ss.mtx.Unlock()
	}

	return ss.insert(aElement)
} // Insert()

func (ss *TSortedSlice[T]) rename(aOldValue, aNewValue T) bool {
	if (0 == len(ss.data)) || (aOldValue == aNewValue) {
		return false
	}

	idx := ss.findIndex(aOldValue)
	if 0 > idx { // ID not found
		return ss.insert(aNewValue)
	}

	if !ss.insert(aNewValue) {
		// This should only happen it there's an OOM problem.
		// Hence we just replace the aOldValue by aNewValue and
		// sort the list again.
		if ss.data[idx] != aNewValue {
			ss.data[idx] = aNewValue
			slices.Sort(ss.data) // ascending
			return true
		}
		return false
	}

	return ss.delete(aOldValue)
} // rename()

// Rename changes an element in the sorted slice and maintains order.
func (ss *TSortedSlice[T]) Rename(aOldValue, aNewValue T) bool {
	if ss.safe {
		ss.mtx.Lock()
		defer ss.mtx.Unlock()
	}

	return ss.rename(aOldValue, aNewValue)
} // Rename()

func (ss *TSortedSlice[T]) string() string {
	if 0 == len(ss.data) {
		return "[]"
	}

	var builder strings.Builder
	// Here we ignore the return values of the `builder` functions
	// since we can't do anything about them anyway.
	builder.WriteString("[")
	for idx, elem := range ss.data {
		if 0 < idx {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%v", elem))
	}
	builder.WriteString("]")

	return builder.String()
} // string()

// `String()` implements the `fmt.Stringer` interface.
//
// For each element, the method appends the element's string
// representation while elements are separated by ", ".
//
// Returns:
// - `string`: The list's contents as a string.
func (ss *TSortedSlice[T]) String() string {
	if ss.safe {
		ss.mtx.RLock()
		defer ss.mtx.RUnlock()
	}

	return ss.string()
} // String()

/* EoF */
