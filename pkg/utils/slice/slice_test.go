package slice

import (
	"reflect"
	"sort"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("map integers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := Map(input, func(x int) int { return x * 2 })
		expected := []int{2, 4, 6, 8, 10}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("map strings", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		result := Map(input, func(s string) string { return s + s })
		expected := []string{"aa", "bb", "cc"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestFilter(t *testing.T) {
	t.Run("filter even numbers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6}
		result := Filter(input, func(x int) bool { return x%2 == 0 })
		expected := []int{2, 4, 6}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("filter non-empty strings", func(t *testing.T) {
		input := []string{"", "a", "", "b", "c", ""}
		result := Filter(input, func(s string) bool { return s != "" })
		expected := []string{"a", "b", "c"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestReduce(t *testing.T) {
	t.Run("sum numbers", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := Reduce(input, 0, func(acc, x int) int { return acc + x })
		expected := 15

		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("concatenate strings", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		result := Reduce(input, "", func(acc, s string) string { return acc + s })
		expected := "abc"

		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestContains(t *testing.T) {
	t.Run("contains integer", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		if !Contains(input, 3) {
			t.Error("Expected to find 3 in slice")
		}
		if Contains(input, 6) {
			t.Error("Did not expect to find 6 in slice")
		}
	})

	t.Run("contains string", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		if !Contains(input, "b") {
			t.Error("Expected to find 'b' in slice")
		}
		if Contains(input, "d") {
			t.Error("Did not expect to find 'd' in slice")
		}
	})
}

func TestUnique(t *testing.T) {
	t.Run("unique integers", func(t *testing.T) {
		input := []int{1, 2, 2, 3, 3, 3, 4}
		result := Unique(input)
		expected := []int{1, 2, 3, 4}

		sort.Ints(result) // 排序以确保比较顺序一致
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("unique strings", func(t *testing.T) {
		input := []string{"a", "b", "b", "c", "c", "c"}
		result := Unique(input)
		expected := []string{"a", "b", "c"}

		sort.Strings(result) // 排序以确保比较顺序一致
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
