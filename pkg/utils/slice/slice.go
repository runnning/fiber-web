package slice

import (
	"math/rand"

	"golang.org/x/exp/constraints"
)

// Contains checks if an element exists in a slice
func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

// Map applies a function to each element in a slice and returns a new slice
func Map[T, U any](slice []T, f func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = f(v)
	}
	return result
}

// Filter returns a new slice containing only the elements that satisfy the predicate
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce reduces a slice to a single value using an accumulator function
func Reduce[T, U any](slice []T, initial U, f func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = f(result, v)
	}
	return result
}

// Find returns the first element that satisfies the predicate and true if found
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Unique returns a new slice with duplicate elements removed
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	var result []T
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Sort sorts a slice in ascending order
func Sort[T constraints.Ordered](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	quickSort[T](result, 0, len(result)-1)
	return result
}

func quickSort[T constraints.Ordered](slice []T, low, high int) {
	if low < high {
		pivot := partition[T](slice, low, high)
		quickSort[T](slice, low, pivot-1)
		quickSort[T](slice, pivot+1, high)
	}
}

func partition[T constraints.Ordered](slice []T, low, high int) int {
	pivot := slice[high]
	i := low - 1
	for j := low; j < high; j++ {
		if slice[j] <= pivot {
			i++
			slice[i], slice[j] = slice[j], slice[i]
		}
	}
	slice[i+1], slice[high] = slice[high], slice[i+1]
	return i + 1
}

// Chunk splits a slice into chunks of specified size
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Reverse returns a new slice with elements in reverse order
func Reverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// Intersection returns a new slice containing elements that exist in both slices
func Intersection[T comparable](slice1, slice2 []T) []T {
	set := make(map[T]struct{})
	var result []T

	for _, v := range slice1 {
		set[v] = struct{}{}
	}

	for _, v := range slice2 {
		if _, ok := set[v]; ok {
			result = append(result, v)
		}
	}

	return Unique(result)
}

// Difference returns a new slice containing elements that exist in slice1 but not in slice2
func Difference[T comparable](slice1, slice2 []T) []T {
	set := make(map[T]struct{})
	for _, v := range slice2 {
		set[v] = struct{}{}
	}

	return Filter(slice1, func(v T) bool {
		_, exists := set[v]
		return !exists
	})
}

// GroupBy groups slice elements by key generated from keyFunc
func GroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := keyFunc(v)
		result[key] = append(result[key], v)
	}
	return result
}

// ToMap converts a slice to map using keyFunc to generate keys
func ToMap[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	result := make(map[K]T)
	for _, v := range slice {
		result[keyFunc(v)] = v
	}
	return result
}

// Union returns a new slice containing unique elements from all input slices
func Union[T comparable](slices ...[]T) []T {
	set := make(map[T]struct{})
	var result []T
	for _, slice := range slices {
		for _, v := range slice {
			if _, ok := set[v]; !ok {
				set[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}

// Shuffle returns a new slice with elements in random order
func Shuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	for i := len(result) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// DeleteAt removes element at index from slice
func DeleteAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// InsertAt inserts element at index in slice
func InsertAt[T any](slice []T, index int, element T) []T {
	if index < 0 || index > len(slice) {
		return slice
	}
	slice = append(slice, element)
	copy(slice[index+1:], slice[index:])
	slice[index] = element
	return slice
}

// Compact returns a new slice with zero values removed
func Compact[T comparable](slice []T) []T {
	var zero T
	return Filter(slice, func(v T) bool {
		return v != zero
	})
}

// Equal returns true if two slices contain the same elements in the same order
func Equal[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

// EqualUnordered returns true if two slices contain the same elements regardless of order
func EqualUnordered[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	freq := make(map[T]int)
	for _, v := range slice1 {
		freq[v]++
	}
	for _, v := range slice2 {
		freq[v]--
		if freq[v] < 0 {
			return false
		}
	}
	return true
}
