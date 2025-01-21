package hook

// UseCache 使用内存缓存一些临时数据
// 由于是临时数据，其 value 应该是一次性数据，用完即丢
func UseCache[K comparable, V any](fn func(K) (V, error)) func(K) (V, error) {
	cache := make(map[K]V)
	return func(key K) (V, error) {
		v, ok := cache[key]
		if ok {
			return v, nil
		}
		v, err := fn(key)
		if err == nil {
			cache[key] = v
		}
		return v, err
	}
}
