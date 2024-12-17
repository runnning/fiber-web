package maps

import (
	"reflect"
	"sort"
	"testing"
)

func TestKeys(t *testing.T) {
	t.Run("get map keys", func(t *testing.T) {
		input := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}
		result := Keys(input)
		expected := []string{"a", "b", "c"}

		sort.Strings(result) // 排序以确保比较顺序一致
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestValues(t *testing.T) {
	t.Run("get map values", func(t *testing.T) {
		input := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}
		result := Values(input)
		expected := []int{1, 2, 3}

		sort.Ints(result) // 排序以确保比较顺序一致
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestMerge(t *testing.T) {
	t.Run("merge maps", func(t *testing.T) {
		map1 := map[string]int{"a": 1, "b": 2}
		map2 := map[string]int{"b": 3, "c": 4}
		result := Merge(map1, map2)
		expected := map[string]int{"a": 1, "b": 3, "c": 4}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestFilter(t *testing.T) {
	t.Run("filter map entries", func(t *testing.T) {
		input := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
			"d": 4,
		}
		result := Filter(input, func(k string, v int) bool {
			return v%2 == 0
		})
		expected := map[string]int{"b": 2, "d": 4}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestMapValues(t *testing.T) {
	t.Run("transform map values", func(t *testing.T) {
		input := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}
		result := MapValues(input, func(v int) int {
			return v * 2
		})
		expected := map[string]int{
			"a": 2,
			"b": 4,
			"c": 6,
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
