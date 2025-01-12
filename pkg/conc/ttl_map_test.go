package conc

import (
	"testing"
	"time"
)

func TestTTLMap(t *testing.T) {
	cache := NewTTLMap[string, string]()
	cache.Set("a", "1", time.Second)
	v, ok := cache.Get("a")
	if !ok {
		t.Fatal("expect ok")
	}
	if v != "1" {
		t.Fatal("expect 1")
	}
	time.Sleep(time.Second)
	_, ok = cache.Get("a")
	if ok {
		t.Fatal("expect not ok")
	}
}
