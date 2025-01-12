package conc

import "time"

// TTLMap 带有过期时间的 map
type TTLMap[K comparable, V any] struct {
	data Map[K, V]
	exp  Map[K, time.Time]
}

func NewTTLMap[K comparable, V any]() *TTLMap[K, V] {
	c := TTLMap[K, V]{}
	go c.cleanup()
	return &c
}

func (c *TTLMap[K, V]) cleanup() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.exp.Range(func(key K, value time.Time) bool {
			if now.After(value) {
				c.data.Delete(key)
				c.exp.Delete(key)
			}
			return true
		})
	}
}

// Set 将在 ttl 后自动删除 k/v
func (c *TTLMap[K, V]) Set(key K, value V, ttl time.Duration) {
	c.data.Store(key, value)
	c.exp.Store(key, time.Now().Add(ttl))
}

// Get 获取未过期的 k/v
func (c *TTLMap[K, V]) Get(key K) (V, bool) {
	var v V
	expAt, ok := c.exp.Load(key)
	if !ok {
		return v, false
	}
	if time.Now().After(expAt) {
		c.data.Delete(key)
		c.exp.Delete(key)
		return v, false
	}
	return c.data.Load(key)
}

// Del 删除 k/v
func (c *TTLMap[K, V]) Del(key K) {
	c.data.Delete(key)
	c.exp.Delete(key)
}
