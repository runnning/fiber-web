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

func TestIntersection(t *testing.T) {
	t.Run("intersection of integers", func(t *testing.T) {
		slice1 := []int{1, 2, 3, 4, 5}
		slice2 := []int{4, 5, 6, 7, 8}
		result := Intersection(slice1, slice2)
		expected := []int{4, 5}

		sort.Ints(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("intersection of strings", func(t *testing.T) {
		slice1 := []string{"a", "b", "c", "d"}
		slice2 := []string{"c", "d", "e", "f"}
		result := Intersection(slice1, slice2)
		expected := []string{"c", "d"}

		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("intersection with duplicates", func(t *testing.T) {
		slice1 := []int{1, 2, 2, 3, 3}
		slice2 := []int{2, 2, 3, 4}
		result := Intersection(slice1, slice2)
		expected := []int{2, 3}

		sort.Ints(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestDifference(t *testing.T) {
	t.Run("difference of integers", func(t *testing.T) {
		slice1 := []int{1, 2, 3, 4, 5}
		slice2 := []int{4, 5, 6, 7, 8}
		result := Difference(slice1, slice2)
		expected := []int{1, 2, 3}

		sort.Ints(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("difference of strings", func(t *testing.T) {
		slice1 := []string{"a", "b", "c", "d"}
		slice2 := []string{"c", "d", "e", "f"}
		result := Difference(slice1, slice2)
		expected := []string{"a", "b"}

		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("difference with duplicates", func(t *testing.T) {
		slice1 := []int{1, 2, 2, 3, 3}
		slice2 := []int{2, 3, 4}
		result := Difference(slice1, slice2)
		expected := []int{1}

		sort.Ints(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

type User struct {
	ID   int
	Role string
}

func TestGroupBy(t *testing.T) {
	t.Run("group users by role", func(t *testing.T) {
		users := []User{
			{ID: 1, Role: "admin"},
			{ID: 2, Role: "user"},
			{ID: 3, Role: "admin"},
			{ID: 4, Role: "user"},
		}
		result := GroupBy(users, func(u User) string { return u.Role })

		if len(result) != 2 {
			t.Errorf("Expected 2 groups, got %d", len(result))
		}
		if len(result["admin"]) != 2 {
			t.Errorf("Expected 2 admins, got %d", len(result["admin"]))
		}
		if len(result["user"]) != 2 {
			t.Errorf("Expected 2 users, got %d", len(result["user"]))
		}
	})

	t.Run("group numbers by parity", func(t *testing.T) {
		nums := []int{1, 2, 3, 4, 5, 6}
		result := GroupBy(nums, func(n int) string {
			if n%2 == 0 {
				return "even"
			}
			return "odd"
		})

		if len(result["even"]) != 3 {
			t.Errorf("Expected 3 even numbers, got %d", len(result["even"]))
		}
		if len(result["odd"]) != 3 {
			t.Errorf("Expected 3 odd numbers, got %d", len(result["odd"]))
		}
	})
}

func TestToMap(t *testing.T) {
	t.Run("users to map by ID", func(t *testing.T) {
		users := []User{
			{ID: 1, Role: "admin"},
			{ID: 2, Role: "user"},
		}
		result := ToMap(users, func(u User) int { return u.ID })

		if len(result) != 2 {
			t.Errorf("Expected map length 2, got %d", len(result))
		}
		if result[1].Role != "admin" {
			t.Errorf("Expected user 1 to be admin, got %s", result[1].Role)
		}
		if result[2].Role != "user" {
			t.Errorf("Expected user 2 to be user, got %s", result[2].Role)
		}
	})
}

func TestUnion(t *testing.T) {
	t.Run("union of integer slices", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		slice2 := []int{3, 4, 5}
		slice3 := []int{5, 6, 7}
		result := Union(slice1, slice2, slice3)

		expected := []int{1, 2, 3, 4, 5, 6, 7}
		sort.Ints(result)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestShuffle(t *testing.T) {
	t.Run("shuffle maintains length and elements", func(t *testing.T) {
		original := []int{1, 2, 3, 4, 5}
		result := Shuffle(original)

		if len(result) != len(original) {
			t.Errorf("Expected length %d, got %d", len(original), len(result))
		}

		// 检查所有元素都存在
		sort.Ints(result)
		expected := []int{1, 2, 3, 4, 5}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Shuffled slice is missing elements")
		}
	})
}

func TestDeleteAt(t *testing.T) {
	t.Run("delete at valid index", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		result := DeleteAt(slice, 2)
		expected := []int{1, 2, 4, 5}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("delete at invalid index", func(t *testing.T) {
		slice := []int{1, 2, 3}
		result := DeleteAt(slice, -1)
		if !reflect.DeepEqual(result, slice) {
			t.Errorf("Expected original slice for invalid index")
		}

		result = DeleteAt(slice, len(slice))
		if !reflect.DeepEqual(result, slice) {
			t.Errorf("Expected original slice for invalid index")
		}
	})
}

func TestInsertAt(t *testing.T) {
	t.Run("insert at valid index", func(t *testing.T) {
		slice := []int{1, 2, 3}
		result := InsertAt(slice, 1, 4)
		expected := []int{1, 4, 2, 3}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("insert at invalid index", func(t *testing.T) {
		slice := []int{1, 2, 3}
		result := InsertAt(slice, -1, 4)
		if !reflect.DeepEqual(result, slice) {
			t.Errorf("Expected original slice for invalid index")
		}

		result = InsertAt(slice, len(slice)+1, 4)
		if !reflect.DeepEqual(result, slice) {
			t.Errorf("Expected original slice for invalid index")
		}
	})
}

func TestCompact(t *testing.T) {
	t.Run("compact integers", func(t *testing.T) {
		slice := []int{0, 1, 0, 2, 0, 3, 0}
		result := Compact(slice)
		expected := []int{1, 2, 3}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("compact strings", func(t *testing.T) {
		slice := []string{"", "a", "", "b", "c", ""}
		result := Compact(slice)
		expected := []string{"a", "b", "c"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestEqual(t *testing.T) {
	t.Run("equal slices", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		slice2 := []int{1, 2, 3}
		if !Equal(slice1, slice2) {
			t.Error("Expected slices to be equal")
		}
	})

	t.Run("unequal slices", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		slice2 := []int{1, 3, 2}
		if Equal(slice1, slice2) {
			t.Error("Expected slices to be unequal")
		}
	})

	t.Run("different length slices", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		slice2 := []int{1, 2}
		if Equal(slice1, slice2) {
			t.Error("Expected slices of different lengths to be unequal")
		}
	})
}

func TestEqualUnordered(t *testing.T) {
	t.Run("equal unordered slices", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		slice2 := []int{3, 1, 2}
		if !EqualUnordered(slice1, slice2) {
			t.Error("Expected slices to be equal regardless of order")
		}
	})

	t.Run("unequal slices", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		slice2 := []int{1, 2, 4}
		if EqualUnordered(slice1, slice2) {
			t.Error("Expected slices to be unequal")
		}
	})

	t.Run("slices with duplicates", func(t *testing.T) {
		slice1 := []int{1, 2, 2, 3}
		slice2 := []int{2, 1, 3, 2}
		if !EqualUnordered(slice1, slice2) {
			t.Error("Expected slices with same elements and duplicates to be equal")
		}
	})
}
