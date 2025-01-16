package slice

import (
	"math/rand"

	"golang.org/x/exp/constraints"
)

// Contains 检查切片中是否存在指定元素
func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

// Map 对切片中的每个元素应用函数并返回新切片
func Map[T, U any](slice []T, f func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = f(v)
	}
	return result
}

// Filter 返回一个新切片，仅包含满足断言函数的元素
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce 使用累加器函数将切片归约为单个值
func Reduce[T, U any](slice []T, initial U, f func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = f(result, v)
	}
	return result
}

// Find 返回第一个满足断言函数的元素和是否找到的标志
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Unique 返回一个新切片，去除重复元素
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

// Sort 对切片进行升序排序
func Sort[T constraints.Ordered](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	quickSort[T](result, 0, len(result)-1)
	return result
}

// quickSort 快速排序实现
func quickSort[T constraints.Ordered](slice []T, low, high int) {
	if low < high {
		pivot := partition[T](slice, low, high)
		quickSort[T](slice, low, pivot-1)
		quickSort[T](slice, pivot+1, high)
	}
}

// partition 快速排序的分区函数
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

// Chunk 将切片分割成指定大小的块
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

// Reverse 返回一个新切片，元素顺序相反
func Reverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// Intersection 返回一个新切片，包含两个切片中都存在的元素
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

// Difference 返回一个新切片，包含在 slice1 中但不在 slice2 中的元素
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

// GroupBy 根据键函数对切片元素进行分组
func GroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := keyFunc(v)
		result[key] = append(result[key], v)
	}
	return result
}

// ToMap 使用键函数将切片转换为映射
func ToMap[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	result := make(map[K]T)
	for _, v := range slice {
		result[keyFunc(v)] = v
	}
	return result
}

// Union 返回一个新切片，包含所有输入切片中的唯一元素
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

// Shuffle 返回一个新切片，元素顺序随机
func Shuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	for i := len(result) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// DeleteAt 从切片中删除指定索引的元素
func DeleteAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// InsertAt 在切片的指定索引处插入元素
func InsertAt[T any](slice []T, index int, element T) []T {
	if index < 0 || index > len(slice) {
		return slice
	}
	slice = append(slice, element)
	copy(slice[index+1:], slice[index:])
	slice[index] = element
	return slice
}

// Compact 返回一个新切片，移除零值元素
func Compact[T comparable](slice []T) []T {
	var zero T
	return Filter(slice, func(v T) bool {
		return v != zero
	})
}

// Equal 判断两个切片是否相等（元素顺序必须相同）
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

// EqualUnordered 判断两个切片是否相等（元素顺序可以不同）
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
