package ziptree

import (
	"math/rand/v2"
)

type Map[K, V any] struct {
	tree   *ZipTree[K]
	values []V
}

type MapIterator[K, V any] struct {
	iterator *ZipIterator[K]
	values   []V
}

func NewMap[K, V any](less LessFn[K]) *Map[K, V] {
	return &Map[K, V]{
		tree:   NewZipTree[K](less),
		values: make([]V, 0),
	}
}

func NewMapWithRandomGenerator[K, V any](less LessFn[K], randomGenerator *rand.Rand) *Map[K, V] {
	return &Map[K, V]{
		tree:   NewZipTreeWithRandomGenerator(less, randomGenerator),
		values: make([]V, 0),
	}
}

func (z *Map[K, V]) NewIterator() *MapIterator[K, V] {
	iter := &MapIterator[K, V]{
		iterator: z.tree.NewIterator(),
	}
	if z.tree.root == SENTINEL {
		return iter
	}
	iter.values = z.values
	return iter
}

func (z *Map[K, V]) NewPrevIterator() *MapIterator[K, V] {
	iter := &MapIterator[K, V]{
		iterator: z.tree.NewPrevIterator(),
	}
	if z.tree.root == SENTINEL {
		return iter
	}
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
	found := z.tree.find(key)
	if found == SENTINEL {
		z.tree.insert(key)
		z.values = append(z.values, value)
		return true
	} else {
		z.values[found] = value
		return false
	}
}

func (z *Map[K, V]) deleteInternalWithValue(keyIdx ZipNodeEntryIndex) bool {
	deleted := z.tree.deleteInternal(keyIdx)
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
	keyIdx := iter.iterator.Index()
	return z.deleteInternalWithValue(keyIdx)
}

// Delete returns true if entry was deleted
// returns false if entry not found
func (z *Map[K, V]) Delete(key K) bool {
	keyIdx := z.tree.find(key)
	return z.deleteInternalWithValue(keyIdx)
}

func (it *MapIterator[K, V]) Value() V {
	var ret V
	if it.iterator.Index() != SENTINEL {
		ret = it.values[it.iterator.Index()]
	}
	return ret
}

func (z *Map[K, V]) iterator(idx ZipNodeEntryIndex) *MapIterator[K, V] {
	if idx == SENTINEL {
		return &MapIterator[K, V]{
			iterator: &ZipIterator[K]{
				current: SENTINEL,
			},
		}
	} else {
		return &MapIterator[K, V]{
			iterator: &ZipIterator[K]{
				current: idx,
				entries: z.tree.entries,
			},
			values: z.values,
		}
	}
}

// Ceiling Returns an iterator pointing to the largest element in the BST greater than or equal to key
func (z *Map[K, V]) Ceiling(key K) *MapIterator[K, V] {
	return z.iterator(z.tree.ceiling(key))

}

// Floor Returns an iterator pointing to largest element in the BST less than or equal to key
func (z *Map[K, V]) Floor(key K) *MapIterator[K, V] {
	return z.iterator(z.tree.floor(key))
}

// UpperBound Returns an iterator pointing to the first element in the tree which is ordered after key
func (z *Map[K, V]) UpperBound(key K) *MapIterator[K, V] {
	return z.iterator(z.tree.upperBound(key))
}

// LowerBound Returns an iterator pointing to the first element in the tree which is not ordered before key
func (z *Map[K, V]) LowerBound(key K) *MapIterator[K, V] {
	return z.iterator(z.tree.lowerBound(key))
}

func (z *Map[K, V]) Find(key K) *MapIterator[K, V] {
	return z.iterator(z.tree.find(key))
}

func (z *Map[K, V]) Minimum() *MapIterator[K, V] {
	return z.iterator(z.tree.minimum())
}

func (z *Map[K, V]) Maximum() *MapIterator[K, V] {
	return z.iterator(z.tree.maximum())
}

func (z *Map[K, V]) AtIndex(idx uint32) *MapIterator[K, V] {
	return z.iterator(z.tree.atIndex(idx))
}

func (it *MapIterator[K, V]) IsEmpty() bool {
	return it.iterator.IsEmpty()
}

func (it *MapIterator[K, V]) Index() ZipNodeEntryIndex {
	return it.iterator.Index()
}

func (it *MapIterator[K, V]) Next() {
	it.iterator.Next()
}

func (it *MapIterator[K, V]) Prev() {
	it.iterator.Prev()
}

func (it *MapIterator[K, V]) Key() K {
	return it.iterator.Key()
}

func (it *MapIterator[K, V]) Parent() ZipNodeEntryIndex {
	return it.iterator.Parent()
}

func (z *Map[K, V]) Size() int {
	return z.tree.Size()
}

func (z *Map[K, V]) Count() int {
	return z.tree.Count()
}
