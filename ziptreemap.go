package ziptree

import "math/rand/v2"

type Map[K, V any] struct {
	ZipTree[K]
	values []V
}

type MapIterator[K, V any] struct {
	ZipIterator[K]
	values []V
}

func NewMap[K, V any](less LessFn[K]) *Map[K, V] {
	return &Map[K, V]{
		ZipTree: ZipTree[K]{
			entries:         make([]ZipNode[K], 0),
			root:            SENTINEL,
			comparator:      Comparator[K](less),
			randomGenerator: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		},
		values: make([]V, 0),
	}
}

func NewMapWithRandomGenerator[K, V any](less LessFn[K], randomGenerator *rand.Rand) *Map[K, V] {
	return &Map[K, V]{
		ZipTree: ZipTree[K]{
			entries:         make([]ZipNode[K], 0),
			root:            SENTINEL,
			comparator:      Comparator[K](less),
			randomGenerator: randomGenerator,
		},
		values: make([]V, 0),
	}
}

func (z *Map[K, V]) NewIterator() *MapIterator[K, V] {
	iter := &MapIterator[K, V]{}
	if z.root == SENTINEL {
		iter.current = SENTINEL
		return iter
	}
	// Find the leftmost node
	iter.current = z.leftMost()
	iter.entries = z.entries
	iter.values = z.values
	return iter
}

func (z *Map[K, V]) NewPrevIterator() *MapIterator[K, V] {
	iter := &MapIterator[K, V]{}
	if z.root == SENTINEL {
		iter.current = SENTINEL
		return iter
	}
	// Find the right most node
	iter.current = z.rightMost()
	iter.entries = z.entries
	iter.values = z.values
	return iter
}

func (z *Map[K, V]) Insert(_ K) {
	panic("not supported for map")
	return
}

// Insert returns true if entry was inserted,
// returns false to indicate update
func (z *Map[K, V]) Put(key K, value V) bool {
	found := z.find(key)
	if found == SENTINEL {
		z.insert(z.root, key)
		z.values = append(z.values, value)
		return true
	} else {
		z.values[found] = value
		return false
	}
}

func (z *Map[K, V]) deleteInternalWithValue(keyIdx ZipNodeEntryIndex) bool {
	deleted := z.deleteInternal(keyIdx)
	if deleted {
		last := ZipNodeEntryIndex(len(z.values) - 1)
		if keyIdx != last {
			z.values[keyIdx] = z.values[last]
		}
		z.values = z.values[:last]
	}
	return deleted
}

// DeleteIter returns true if entry was deleted
// returns false if entry not found
func (z *Map[K, V]) DeleteIter(iter *MapIterator[K, V]) bool {
	keyIdx := iter.Index()
	return z.deleteInternalWithValue(keyIdx)
}

// Delete returns true if entry was deleted
// returns false if entry not found
func (z *Map[K, V]) Delete(key K) bool {
	keyIdx := z.find(key)
	return z.deleteInternalWithValue(keyIdx)
}

func (it *MapIterator[K, V]) Value() V {
	var ret V
	if it.current != SENTINEL {
		ret = it.values[it.current]
	}
	return ret
}

func (z *Map[K, V]) iterator(idx ZipNodeEntryIndex) *MapIterator[K, V] {
	if idx == SENTINEL {
		return &MapIterator[K, V]{
			ZipIterator: ZipIterator[K]{
				current: SENTINEL,
			},
		}
	} else {
		return &MapIterator[K, V]{
			ZipIterator: ZipIterator[K]{
				current: idx,
				entries: z.entries,
			},
			values: z.values,
		}
	}
}

// Ceiling Returns an iterator pointing to the largest element in the BST greater than or equal to key
func (z *Map[K, V]) Ceiling(key K) *MapIterator[K, V] {
	return z.iterator(z.ceiling(key))

}

// Floor Returns an iterator pointing to largest element in the BST less than or equal to key
func (z *Map[K, V]) Floor(key K) *MapIterator[K, V] {
	return z.iterator(z.floor(key))
}

// UpperBound Returns an iterator pointing to the first element in the tree which is ordered after key
func (z *Map[K, V]) UpperBound(key K) *MapIterator[K, V] {
	return z.iterator(z.upperBound(key))
}

// LowerBound Returns an iterator pointing to the first element in the tree which is not ordered before key
func (z *Map[K, V]) LowerBound(key K) *MapIterator[K, V] {
	return z.iterator(z.lowerBound(key))
}

func (z *Map[K, V]) Find(key K) *MapIterator[K, V] {
	return z.iterator(z.find(key))
}

func (z *Map[K, V]) Minimum() *MapIterator[K, V] {
	return z.iterator(z.minimum())
}

func (z *Map[K, V]) Maximum() *MapIterator[K, V] {
	return z.iterator(z.maximum())
}

func (z *Map[K, V]) AtIndex(idx uint32) *MapIterator[K, V] {
	return z.iterator(z.atIndex(idx))
}
