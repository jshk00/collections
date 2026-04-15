package orderdmap

import (
	"iter"

	"github.com/jshk00/collections/list"
)

type Pair[K comparable, V any] struct {
	key K
	val V
}

type Map[K comparable, V any] struct {
	l    *list.List[Pair[K, V]]
	data map[K]*list.Node[Pair[K, V]]
}

func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		l:    list.New[Pair[K, V]](),
		data: make(map[K]*list.Node[Pair[K, V]]),
	}
}

// Set add the key and value pair into map if key is alredy present then the value is updated
func (m *Map[K, V]) Set(k K, v V) {
	if n, ok := m.data[k]; ok {
		n.Value.val = v
		return
	}
	n := m.l.PushBack(Pair[K, V]{key: k, val: v})
	m.data[k] = n
}

// Get the value from map given key. If not present will return false and default value.
func (m *Map[K, V]) Get(k K) (V, bool) {
	v, ok := m.data[k]
	var val V
	if ok {
		val = v.Value.val
	}
	return val, ok
}

// Delete removes the key from map as well underlying list
func (m *Map[K, V]) Delete(k K) {
	v, ok := m.data[k]
	if ok {
		delete(m.data, k)
		m.l.Remove(v)
	}
}

// Move the key/val pair into front of underlying list
func (m *Map[K, V]) MoveToFront(key K) {
	v, ok := m.data[key]
	if ok {
		m.l.MoveToFront(v)
	}
}

// Move the key/vak pair into back of underlying list
func (m *Map[K, V]) MoveToBack(key K) {
	v, ok := m.data[key]
	if ok {
		m.l.MoveToBack(v)
	}
}

// Keys provide iterator over map keys
func (m *Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for e := m.l.Front(); e != nil; e = e.Next() {
			if !yield(e.Value.key) {
				return
			}
		}
	}
}

// Values provide iterator over map values
func (m *Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for e := m.l.Front(); e != nil; e = e.Next() {
			if !yield(e.Value.val) {
				return
			}
		}
	}
}

// Iter provides iterator over map key and values
func (m *Map[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for e := m.l.Front(); e != nil; e = e.Next() {
			if !yield(e.Value.key, e.Value.val) {
				return
			}
		}
	}
}

// IterBack provides iterator over map key and values in reverse order
func (m *Map[K, V]) IterBack() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for e := m.l.Back(); e != nil; e = e.Prev() {
			if !yield(e.Value.key, e.Value.val) {
				return
			}
		}
	}
}
