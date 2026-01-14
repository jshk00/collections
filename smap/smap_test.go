package smap

import (
	"sync"
	"testing"
)

// TestMap checks if concurrent put call to maps are safe or not
func TestMap(t *testing.T) {
	data := []struct {
		key  string
		val  string
		want string
	}{
		{"key1", "val1", "val1"},
		{"key2", "val2", "val2"},
		{"key3", "val3", "val3"},
		{"key4", "val4", "val4"},
		{"key5", "val5", "val5"},
		{"key6", "val6", "val6"},
		{"key7", "val7", "val7"},
		{"key8", "val8", "val8"},
		{"key9", "val9", "val9"},
		{"key10", "val10", "val10"},
		{"key11", "val11", "val11"},
		{"key12", "val12", "val12"},
	}

	t.Run("concurrent-put", func(t *testing.T) {
		var wg sync.WaitGroup
		mm := New[string, string](len(data))
		wg.Add(len(data))
		for _, el := range data {
			el := el
			go func() {
				mm.Set(el.key, el.val)
				wg.Done()
			}()
		}
		wg.Wait()
		for _, d := range data {
			if v, _ := mm.Get(d.key); v != d.want {
				t.Errorf(
					"must be concurrency issue we wanted %s but got %s for %s key",
					d.want,
					v,
					d.key,
				)
			}
		}
	})

	t.Run("concurrent-get", func(t *testing.T) {
		mm := New[string, string](len(data))
		for _, el := range data {
			mm.Set(el.key, el.val)
		}
		var wg sync.WaitGroup
		wg.Add(len(data))
		for _, el := range data {
			el := el
			go func() {
				if v, _ := mm.Get(el.key); v == "" {
					t.Error("value should not be empty")
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})

	t.Run("concurrent-put-get", func(t *testing.T) {
		mm := New[string, string](len(data))
		var wg sync.WaitGroup
		var mu sync.RWMutex
		wg.Add(2)
		go func() {
			mu.RLock()
			for _, el := range data {
				mm.Set(el.key, el.val)
			}
			mu.RUnlock()
			wg.Done()
		}()

		mu.RLock()
		go func() {
			for _, el := range data {
				if v, _ := mm.Get(el.key); v == "" {
					t.Log("values is empty")
				}
			}
			wg.Done()
		}()
		mu.RUnlock()
		wg.Wait()
		t.Log(mm.Keys())
		t.Log(mm.Vals())
	})
}
