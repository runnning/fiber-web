package maps

// Keys returns a slice of all keys in the map
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice of all values in the map
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Merge merges multiple maps into a new map
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Filter returns a new map containing only the key-value pairs that satisfy the predicate
func Filter[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// MapValues returns a new map with the same keys but transformed values
func MapValues[K comparable, V, R any](m map[K]V, transform func(V) R) map[K]R {
	result := make(map[K]R)
	for k, v := range m {
		result[k] = transform(v)
	}
	return result
}

// Invert returns a new map with keys and values swapped
func Invert[K, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K)
	for k, v := range m {
		result[v] = k
	}
	return result
}

// Pick returns a new map with only the specified keys
func Pick[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V)
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}

// Omit returns a new map without the specified keys
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

// HasKey checks if a key exists in the map
func HasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}

// HasValue checks if a value exists in the map
func HasValue[K comparable, V comparable](m map[K]V, value V) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}

// GetOrDefault returns the value for key if present, otherwise returns defaultValue
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultValue
}

// Update updates multiple keys in the map with their corresponding values
func Update[K comparable, V any](m map[K]V, updates map[K]V) {
	for k, v := range updates {
		m[k] = v
	}
}

// DeleteKeys deletes multiple keys from the map
func DeleteKeys[K comparable, V any](m map[K]V, keys []K) {
	for _, k := range keys {
		delete(m, k)
	}
}

// GroupByValue groups keys by their values
func GroupByValue[K comparable, V comparable](m map[K]V) map[V][]K {
	result := make(map[V][]K)
	for k, v := range m {
		result[v] = append(result[v], k)
	}
	return result
}

// Clone returns a shallow copy of the map
func Clone[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
