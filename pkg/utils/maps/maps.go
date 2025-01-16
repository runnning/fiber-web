package maps

// Keys 返回 map 中所有的键
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回 map 中所有的值
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Merge 合并多个 map 到一个新的 map 中
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Filter 返回一个新的 map，仅包含满足断言函数的键值对
func Filter[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// MapValues 返回一个新的 map，保持键不变，值经过转换函数处理
func MapValues[K comparable, V, R any](m map[K]V, transform func(V) R) map[K]R {
	result := make(map[K]R)
	for k, v := range m {
		result[k] = transform(v)
	}
	return result
}

// Invert 返回一个新的 map，键值对交换
func Invert[K, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K)
	for k, v := range m {
		result[v] = k
	}
	return result
}

// Pick 返回一个新的 map，仅包含指定的键
func Pick[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V)
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}

// Omit 返回一个新的 map，排除指定的键
func Omit[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		exclude := false
		for _, key := range keys {
			if k == key {
				exclude = true
				break
			}
		}
		if !exclude {
			result[k] = v
		}
	}
	return result
}

// HasKey 检查 map 中是否存在指定的键
func HasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}

// HasValue 检查 map 中是否存在指定的值
func HasValue[K comparable, V comparable](m map[K]V, value V) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}

// GetOrDefault 获取指定键的值，如果不存在则返回默认值
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultValue
}

// Update 使用另一个 map 中的键值对更新当前 map
func Update[K comparable, V any](m map[K]V, updates map[K]V) {
	for k, v := range updates {
		m[k] = v
	}
}

// DeleteKeys 从 map 中删除指定的多个键
func DeleteKeys[K comparable, V any](m map[K]V, keys []K) {
	for _, k := range keys {
		delete(m, k)
	}
}

// GroupByValue 根据值对键进行分组
func GroupByValue[K comparable, V comparable](m map[K]V) map[V][]K {
	result := make(map[V][]K)
	for k, v := range m {
		result[v] = append(result[v], k)
	}
	return result
}

// Clone 返回 map 的浅拷贝
func Clone[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
