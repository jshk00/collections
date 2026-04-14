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

func (m *Map[K, V]) Set(k K, v V) {
	if n, ok := m.data[k]; ok {
		n.Value.val = v
		return
	}
	n := m.l.PushBack(Pair[K, V]{key: k, val: v})
	m.data[k] = n
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	v, ok := m.data[k]
	return v.Value.val, ok
}

func (m *Map[K, V]) Delete(k K) {
	v, ok := m.data[k]
	if ok {
		delete(m.data, k)
		m.l.Remove(v)
	}
}

func (m *Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for e := m.l.Front(); e != nil; e = e.Next() {
			if !yield(e.Value.key) {
				return
			}
		}
	}
}

func (m *Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for e := m.l.Front(); e != nil; e = e.Next() {
			if !yield(e.Value.val) {
				return
			}
		}
	}
}

func (m *Map[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for e := m.l.Front(); e != nil; e = e.Next() {
			if !yield(e.Value.key, e.Value.val) {
				return
			}
		}
	}
}

func (m *Map[K, V]) IterBack() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for e := m.l.Back(); e != nil; e = e.Prev() {
			if !yield(e.Value.key, e.Value.val) {
				return
			}
		}
	}
}
