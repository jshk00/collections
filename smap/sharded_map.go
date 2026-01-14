package smap

import (
	"hash/fnv"
	"iter"
	"unsafe"
)

const (
	DefaultShardBuffer = 1024
	DefaultShardCount  = 32
)

type ShardedMap[K comparable, V any] struct {
	shards []*Map[K, V]
	hash   func(key K) uint64
	count  uint
}

func NewShardedMap[K comparable, V any](
	count uint,
	hashFn func(key K) uint64,
) *ShardedMap[K, V] {
	if count <= 0 {
		count = DefaultShardCount
	}
	shards := make([]*Map[K, V], 0, count)
	for range count {
		shards = append(shards, New[K, V](DefaultShardBuffer))
	}
	return &ShardedMap[K, V]{shards: shards, hash: hashFn, count: count}
}

func (sm *ShardedMap[K, V]) GetShard(key K) *Map[K, V] {
	return sm.shards[uint(sm.hash(key))%sm.count]
}

func (sm *ShardedMap[K, V]) Set(k K, v V) {
	shard := sm.GetShard(k)
	shard.Set(k, v)
}

func (sm *ShardedMap[K, V]) Get(key K) (V, bool) {
	shard := sm.GetShard(key)
	return shard.Get(key)
}

func (sm *ShardedMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for _, shard := range sm.shards {
			shard.mu.RLock()
			for k := range shard.items {
				if !yield(k) {
					shard.mu.RUnlock()
					return
				}
			}
			shard.mu.RUnlock()
		}
	}
}

func (sm *ShardedMap[K, V]) Vals() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, shard := range sm.shards {
			shard.mu.RLock()
			for _, v := range shard.items {
				if !yield(v) {
					shard.mu.RUnlock()
					return
				}
			}
			shard.mu.RUnlock()
		}
	}
}

func (sm *ShardedMap[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, shard := range sm.shards {
			shard.mu.RLock()
			for k, v := range shard.items {
				if !yield(k, v) {
					shard.mu.RUnlock()
					return
				}
			}
			shard.mu.RUnlock()
		}
	}
}

func (sm *ShardedMap[K, V]) HasKey(k K) bool {
	shard := sm.GetShard(k)
	_, ok := shard.Get(k)
	return ok
}

func (sm *ShardedMap[K, V]) Len() int {
	count := 0
	for _, m := range sm.shards {
		count += m.Len()
	}
	return count
}

func (sm *ShardedMap[K, V]) Delete(k K) {
	shard := sm.GetShard(k)
	shard.Delete(k)
}

func (sm *ShardedMap[K, V]) DeleteFunc(fn func(k K, v V) bool) {
	for _, shard := range sm.shards {
		shard.DeleteFunc(fn)
	}
}

func (sm *ShardedMap[K, V]) Stats() []ShardStat {
	shards := make([]ShardStat, 0, sm.count)
	for i, shard := range sm.shards {
		shards = append(shards, ShardStat{Index: i, Items: shard.Len()})
	}
	return shards
}

type ShardStat struct {
	Index int
	Items int
}

// Hashes the string using fnv64-1a
func HashString(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(unsafe.Slice(unsafe.StringData(s), len(s)))
	return h.Sum64()
}

const (
	prime1 uint64 = 0x9e3779b97f4a7c15
	prime2 uint64 = 0xbf58476d1ce4e5b9
	prime3 uint64 = 0x94d049bb133111eb
)

// Hashes the integer using SplitMix64 Method.
func hasUint64(x uint64) uint64 {
	x += prime1                  // Increment with golden ration costant
	x = (x ^ (x >> 30)) * prime2 // xor the variable with variable right bit shifted 30 and multiply with constant
	x = (x ^ (x >> 27)) * prime3 // xor the variable with variable right bit shifted 27 and multiply with constant
	return x ^ (x >> 31)         // xor the variable with variable right bit shifted 31
}

func HashUint[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](x T) uint64 {
	return hasUint64(uint64(x))
}

func HashInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64](x T) uint64 {
	return hasUint64(uint64(int64(x)))
}
