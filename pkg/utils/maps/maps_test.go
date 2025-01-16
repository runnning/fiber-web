package maps

import (
	"reflect"
	"sort"
	"testing"
)

func TestKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	result := Keys(m)
	sort.Strings(result)
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	result := Values(m)
	sort.Ints(result)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestMerge(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}
	m3 := map[string]int{"c": 5, "d": 6}
	result := Merge(m1, m2, m3)
	expected := map[string]int{"a": 1, "b": 3, "c": 5, "d": 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestFilter(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	result := Filter(m, func(k string, v int) bool {
		return v%2 == 0
	})
	expected := map[string]int{"b": 2, "d": 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestMapValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	result := MapValues(m, func(v int) string {
		return string(rune('A' + v - 1))
	})
	expected := map[string]string{"a": "A", "b": "B", "c": "C"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestInvert(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	result := Invert(m)
	expected := map[int]string{1: "a", 2: "b", 3: "c"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestPick(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	result := Pick(m, []string{"b", "c", "e"})
	expected := map[string]int{"b": 2, "c": 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestOmit(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	result := Omit(m, []string{"b", "c", "e"})
	expected := map[string]int{"a": 1, "d": 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestHasKey(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	tests := []struct {
		key      string
		expected bool
	}{
		{"a", true},
		{"b", true},
		{"c", false},
	}

	for _, test := range tests {
		if result := HasKey(m, test.key); result != test.expected {
			t.Errorf("HasKey(%q) = %v, 期望 %v", test.key, result, test.expected)
		}
	}
}

func TestHasValue(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	tests := []struct {
		value    int
		expected bool
	}{
		{1, true},
		{2, true},
		{3, false},
	}

	for _, test := range tests {
		if result := HasValue(m, test.value); result != test.expected {
			t.Errorf("HasValue(%d) = %v, 期望 %v", test.value, result, test.expected)
		}
	}
}

func TestGetOrDefault(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	tests := []struct {
		key      string
		def      int
		expected int
	}{
		{"a", 0, 1},
		{"b", 0, 2},
		{"c", 3, 3},
	}

	for _, test := range tests {
		if result := GetOrDefault(m, test.key, test.def); result != test.expected {
			t.Errorf("GetOrDefault(%q, %d) = %d, 期望 %d", test.key, test.def, result, test.expected)
		}
	}
}

func TestUpdate(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	updates := map[string]int{"b": 20, "c": 3}
	Update(m, updates)
	expected := map[string]int{"a": 1, "b": 20, "c": 3}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, m)
	}
}

func TestDeleteKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	DeleteKeys(m, []string{"b", "c", "e"})
	expected := map[string]int{"a": 1, "d": 4}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, m)
	}
}

func TestGroupByValue(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 1, "d": 3}
	result := GroupByValue(m)

	for _, keys := range result {
		sort.Strings(keys)
	}

	expected := map[int][]string{
		1: {"a", "c"},
		2: {"b"},
		3: {"d"},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("期望 %v, 得到 %v", expected, result)
	}
}

func TestClone(t *testing.T) {
	t.Run("基本克隆测试", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		result := Clone(m)
		if !reflect.DeepEqual(result, m) {
			t.Errorf("期望 %v, 得到 %v", m, result)
		}
	})

	t.Run("修改克隆后的map", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		result := Clone(m)
		result["c"] = 3
		if reflect.DeepEqual(result, m) {
			t.Error("修改克隆后的map不应影响原map")
		}
	})
}
