package ziptree

import (
	"fmt"
	"math/rand/v2"
)

type ZipNodeIndex uint32

const SENTINEL = ^ZipNodeIndex(0)

type ZipNode[K, V any] struct {
	key                 K
	left, right, parent ZipNodeIndex
	rank                uint32
	data                V
}

type ZipTree[K, V any] struct {
	entries         []ZipNode[K, V]
	root            ZipNodeIndex
	comparator      Comparator[K]
	randomGenerator *rand.Rand
}

func (z *ZipTree[K, V]) find(key K) ZipNodeIndex {
	root := z.root
	for root != SENTINEL {
		if z.comparator.LessThan(key, z.entries[root].key) {
			root = z.entries[root].left
		} else if z.comparator.GreaterThan(key, z.entries[root].key) {
			root = z.entries[root].right
		} else {
			break
		}
	}
	return root
}

func (z *ZipTree[K, V]) insert(rootIdx ZipNodeIndex, key K, value V) {
	idx := ZipNodeIndex(len(z.entries))
	var rank uint32 = 0
	for z.randomGenerator.Int32N(2) != 0 {
		rank++
	}
	z.entries = append(z.entries, ZipNode[K, V]{
		key:    key,
		data:   value,
		rank:   rank,
		left:   SENTINEL,
		right:  SENTINEL,
		parent: SENTINEL,
	})
	curr := rootIdx
	prev := SENTINEL
	for curr != SENTINEL && (rank < z.entries[curr].rank || (rank == z.entries[curr].rank &&
		z.comparator.GreaterThan(key, z.entries[curr].key))) {
		prev = curr
		if z.comparator.LessThan(key, z.entries[curr].key) {
			curr = z.entries[curr].left
		} else {
			curr = z.entries[curr].right
		}
	}
	if curr == rootIdx {
		z.root = idx
	} else if z.comparator.LessThan(key, z.entries[prev].key) {
		z.entries[prev].left = idx
		z.entries[idx].parent = prev
	} else {
		z.entries[prev].right = idx
		z.entries[idx].parent = prev
	}

	if curr == SENTINEL {
		// check if we can just insert
		return
	} else if z.comparator.LessThan(key, z.entries[curr].key) {
		z.entries[idx].right = curr
		z.entries[curr].parent = idx
	} else {
		z.entries[idx].left = curr
		z.entries[curr].parent = idx
	}
	prev = idx
	for curr != SENTINEL {
		fix := prev
		if z.comparator.GreaterThan(key, z.entries[curr].key) {
			for curr != SENTINEL && z.comparator.GreaterThanOrEqual(key, z.entries[curr].key) {
				prev = curr
				curr = z.entries[curr].right
			}
		} else {
			for curr != SENTINEL && z.comparator.LessThanOrEqual(key, z.entries[curr].key) {
				prev = curr
				curr = z.entries[curr].left
			}
		}

		if z.comparator.LessThan(key, z.entries[fix].key) || (fix == idx && z.comparator.LessThan(key, z.entries[prev].key)) {
			z.entries[fix].left = curr
		} else {
			z.entries[fix].right = curr
		}
		if curr != SENTINEL {
			z.entries[curr].parent = fix
		}

	}
}

func (z *ZipTree[K, V]) deleteIndex(keyIdx ZipNodeIndex) {
	z.delete(keyIdx)
	last := ZipNodeIndex(len(z.entries) - 1)
	if keyIdx != last {
		z.entries[keyIdx] = z.entries[last]
		left, right := z.entries[keyIdx].left, z.entries[keyIdx].right
		if left != SENTINEL {
			z.entries[left].parent = keyIdx
		}
		if right != SENTINEL {
			z.entries[right].parent = keyIdx
		}
		parent := z.entries[keyIdx].parent
		if parent != SENTINEL {
			if last == z.entries[parent].left {
				z.entries[parent].left = keyIdx
			} else {
				z.entries[parent].right = keyIdx
			}
		}
		if last == z.root {
			z.root = keyIdx
		}
	}
	z.entries = z.entries[:last]
}

func (z *ZipTree[K, V]) delete(keyIdx ZipNodeIndex) {
	curr := keyIdx
	key := z.entries[curr].key
	prev := z.entries[curr].parent
	left, right := z.entries[curr].left, z.entries[curr].right
	if left == SENTINEL {
		curr = right
	} else if right == SENTINEL {
		curr = left
	} else if z.entries[left].rank >= z.entries[right].rank {
		curr = left
	} else {
		curr = right
	}

	parent := SENTINEL
	if z.root == keyIdx {
		z.root = curr
	} else {
		if z.comparator.LessThan(key, z.entries[prev].key) {
			z.entries[prev].left = curr
		} else {
			z.entries[prev].right = curr
		}
		parent = prev
	}

	if curr != SENTINEL {
		z.entries[curr].parent = parent
	}

	for left != SENTINEL && right != SENTINEL {
		leftNode := z.entries[left]
		rightNode := z.entries[right]

		if leftNode.rank >= rightNode.rank {
			for !(left == SENTINEL || z.entries[left].rank < rightNode.rank) {
				prev = left
				left = z.entries[left].right
			}
			z.entries[prev].right = right
			z.entries[right].parent = prev
		} else {
			for !(right == SENTINEL || leftNode.rank >= z.entries[right].rank) {
				prev = right
				right = z.entries[right].left
			}
			z.entries[prev].left = left
			z.entries[left].parent = prev
		}
	}
}

// DisplayTree the tree in a human-readable way
func (z *ZipTree[K, V]) displayTree(rootIdx ZipNodeIndex, prefix string, isLeft bool, hasBoth bool) string {
	if rootIdx != SENTINEL {
		node := &z.entries[rootIdx]

		ret := prefix
		if isLeft && hasBoth {
			ret = ret + "├── "
		} else {
			ret = ret + "└── "
		}
		ret = ret + fmt.Sprintf("Key: %v, Idx: %d, Value: %v, Rank: %d, Parent: %d\n", node.key, rootIdx, node.data, node.rank, node.parent)
		newPrefix := prefix
		if isLeft && hasBoth {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		nodeHasBoth := node.left != SENTINEL && node.right != SENTINEL
		left, right := z.displayTree(node.left, newPrefix, true, nodeHasBoth), z.displayTree(node.right, newPrefix, false, nodeHasBoth)
		ret = ret + left + right
		return ret
	}
	return ""
}

func (z *ZipTree[K, V]) displayTreeNodesInOrder(rootIdx ZipNodeIndex) string {
	if rootIdx != SENTINEL {
		node := z.entries[rootIdx]
		ret := z.displayTreeNodesInOrder(node.left) +
			fmt.Sprintf("Key: %v, Idx: %d, Value: %v, Rank: %d, Parent: %d\n", node.key, rootIdx, node.data, node.rank, node.parent)
		ret = ret + z.displayTreeNodesInOrder(node.right)
		return ret
	}
	return ""
}

func (z *ZipTree[K, V]) leftMost() ZipNodeIndex {
	current := z.root
	for z.entries[current].left != SENTINEL {
		current = z.entries[current].left
	}
	return current
}

func (z *ZipTree[K, V]) rightMost() ZipNodeIndex {
	current := z.root
	for z.entries[current].right != SENTINEL {
		current = z.entries[current].right
	}
	return current
}

func (z *ZipTree[K, V]) minimum() ZipNodeIndex {
	if z.root == SENTINEL {
		return z.root
	}
	return z.leftMost()
}

func (z *ZipTree[K, V]) maximum() ZipNodeIndex {
	if z.root == SENTINEL {
		return z.root
	}
	return z.rightMost()
}

func (z *ZipTree[K, V]) floor(key K) ZipNodeIndex {
	res := SENTINEL
	root := z.root

	for root != SENTINEL {
		if z.comparator.LessThan(key, z.entries[root].key) {
			root = z.entries[root].left
		} else {
			res = root
			root = z.entries[root].right
		}
	}
	return res
}

func (z *ZipTree[K, V]) ceiling(key K) ZipNodeIndex {
	res := SENTINEL
	root := z.root

	for root != SENTINEL {
		if z.comparator.GreaterThan(key, z.entries[root].key) {
			root = z.entries[root].right
		} else {
			res = root
			root = z.entries[root].left
		}
	}
	return res
}

func (z *ZipTree[K, V]) upperBound(key K) ZipNodeIndex {
	res := SENTINEL
	root := z.root

	for root != SENTINEL {
		if z.comparator.GreaterThanOrEqual(key, z.entries[root].key) {
			root = z.entries[root].right
		} else {
			res = root
			root = z.entries[root].left
		}
	}
	return res
}

func (z *ZipTree[K, V]) lowerBound(key K) ZipNodeIndex {
	return z.ceiling(key)
}

func (z *ZipTree[K, V]) iterator(idx ZipNodeIndex) *ZipIterator[K, V] {
	if idx == SENTINEL {
		return &ZipIterator[K, V]{
			current: SENTINEL,
		}
	} else {
		return &ZipIterator[K, V]{
			current: idx,
			entries: z.entries,
		}
	}
}

func NewZipTree[K, V any](less LessFn[K]) *ZipTree[K, V] {
	return &ZipTree[K, V]{
		entries:         make([]ZipNode[K, V], 0),
		root:            SENTINEL,
		comparator:      Comparator[K](less),
		randomGenerator: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

func NewZipTreeWithRandomGenerator[K, V any](less LessFn[K], randomGenerator *rand.Rand) *ZipTree[K, V] {
	return &ZipTree[K, V]{
		entries:         make([]ZipNode[K, V], 0),
		root:            SENTINEL,
		comparator:      Comparator[K](less),
		randomGenerator: randomGenerator,
	}
}

// Ceiling Returns an iterator pointing to the largest element in the BST greater than or equal to key
func (z *ZipTree[K, V]) Ceiling(key K) *ZipIterator[K, V] {
	return z.iterator(z.ceiling(key))

}

// Floor Returns an iterator pointing to largest element in the BST less than or equal to key)
func (z *ZipTree[K, V]) Floor(key K) *ZipIterator[K, V] {
	return z.iterator(z.floor(key))
}

// UpperBound Returns an iterator pointing to the first element in the tree which is ordered after key
func (z *ZipTree[K, V]) UpperBound(key K) *ZipIterator[K, V] {
	return z.iterator(z.upperBound(key))
}

// LowerBound Returns an iterator pointing to the first element in the tree which is not ordered before key
func (z *ZipTree[K, V]) LowerBound(key K) *ZipIterator[K, V] {
	return z.iterator(z.lowerBound(key))
}

func (z *ZipTree[K, V]) Find(key K) *ZipIterator[K, V] {
	return z.iterator(z.find(key))
}

func (z *ZipTree[K, V]) Minimum() *ZipIterator[K, V] {
	return z.iterator(z.minimum())
}

func (z *ZipTree[K, V]) Maximum() *ZipIterator[K, V] {
	return z.iterator(z.maximum())
}

// Insert returns true if entry was inserted,
// returns false to indicate update
func (z *ZipTree[K, V]) Insert(key K, value V) bool {
	found := z.find(key)
	if found == SENTINEL {
		z.insert(z.root, key, value)
		return true
	} else {
		z.entries[found].data = value
		return false
	}
}

// DeleteIter returns true if entry was deleted
// returns false if entry not found
func (z *ZipTree[K, V]) DeleteIter(iter *ZipIterator[K, V]) bool {
	keyIdx := iter.Index()
	if keyIdx == SENTINEL {
		return false
	}
	z.deleteIndex(keyIdx)
	return true
}

// Delete returns true if entry was deleted
// returns false if entry not found
func (z *ZipTree[K, V]) Delete(key K) bool {
	keyIdx := z.find(key)
	if keyIdx == SENTINEL {
		return false
	}
	z.deleteIndex(keyIdx)
	return true
}

func (z *ZipTree[K, V]) DisplayTree() string {
	return z.displayTree(z.root, "", false, false)
}

func (z *ZipTree[K, V]) DisplayTreeNodesInOrder() string {
	return z.displayTreeNodesInOrder(z.root)
}

func (z *ZipTree[K, V]) Size() int {
	return len(z.entries)
}

func (z *ZipTree[K, V]) NewIterator() *ZipIterator[K, V] {
	iter := &ZipIterator[K, V]{}
	if z.root == SENTINEL {
		iter.current = SENTINEL
		return iter
	}
	// Find the leftmost node
	iter.current = z.leftMost()
	iter.entries = z.entries
	return iter
}

func (z *ZipTree[K, V]) NewPrevIterator() *ZipIterator[K, V] {
	iter := &ZipIterator[K, V]{}
	if z.root == SENTINEL {
		iter.current = SENTINEL
		return iter
	}
	// Find the right most node
	iter.current = z.rightMost()
	iter.entries = z.entries
	return iter
}
