package ziptree

import (
	"fmt"
	"math/bits"
	"math/rand/v2"
	"strings"
)

type ZipNodeEntryIndex uint32

const SENTINEL = ^ZipNodeEntryIndex(0)

type ZipNode[K, V any] struct {
	key                 K
	left, right, parent ZipNodeEntryIndex
	rank, count         uint32
	data                V
}

type ZipTree[K, V any] struct {
	entries         []ZipNode[K, V]
	root            ZipNodeEntryIndex
	comparator      Comparator[K]
	randomGenerator *rand.Rand
}

func (zn *ZipNode[K, V]) String() string {
	return fmt.Sprintf("Key: %v,  Value: %v, Rank: (%d, %d), Count: %d", zn.key, zn.data, zn.rank>>16, zn.rank&0x0000ffff, zn.count)
}

func (z *ZipTree[K, V]) String() string {
	var sb strings.Builder
	z.displayTree(z.root, "", false, false, &sb)
	return sb.String()
}

func (z *ZipTree[K, V]) find(key K) ZipNodeEntryIndex {
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

func (z *ZipTree[K, V]) insert(rootIdx ZipNodeEntryIndex, key K, value V) {
	idx := ZipNodeEntryIndex(len(z.entries))
	// zip-zip tree
	var r1 uint32 = 0
	for z.randomGenerator.Int32N(2) != 0 {
		r1++
	}
	n := uint32(len(z.entries))
	r2 := uint32(0)
	if n > 0 {
		logOfN := bits.Len32(n+1) - 1
		r2 = z.randomGenerator.Uint32N(uint32(logOfN * logOfN * logOfN))
	}
	rank := r1<<16 | (1 + r2)
	z.entries = append(z.entries, ZipNode[K, V]{
		key:    key,
		data:   value,
		rank:   rank,
		left:   SENTINEL,
		right:  SENTINEL,
		parent: SENTINEL,
		count:  1,
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
	} else {
		if z.comparator.LessThan(key, z.entries[prev].key) {
			z.entries[prev].left = idx
		} else {
			z.entries[prev].right = idx
		}
		z.entries[idx].parent = prev
	}

	if curr != SENTINEL {
		if z.comparator.LessThan(key, z.entries[curr].key) {
			z.entries[idx].right = curr
		} else {
			z.entries[idx].left = curr
		}
		z.entries[curr].parent = idx
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
			z.fixupCount(fix, idx)
		}
	}
	z.fixupCount(idx, SENTINEL)
}

func (z *ZipTree[K, V]) compact(keyIdx ZipNodeEntryIndex) {
	last := ZipNodeEntryIndex(len(z.entries) - 1)
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

func (z *ZipTree[K, V]) deleteInternal(keyIdx ZipNodeEntryIndex) bool {
	if keyIdx == SENTINEL {
		return false
	}
	z.deleteIndex(keyIdx)
	z.compact(keyIdx)
	return true
}

func (z *ZipTree[K, V]) deleteIndex(keyIdx ZipNodeEntryIndex) {
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

	if z.root == keyIdx {
		z.root = curr
	} else {
		if z.comparator.LessThan(key, z.entries[prev].key) {
			z.entries[prev].left = curr
		} else {
			z.entries[prev].right = curr
		}
	}

	if curr != SENTINEL {
		z.entries[curr].parent = prev
	}

	for left != SENTINEL && right != SENTINEL {
		leftNode := &z.entries[left]
		rightNode := &z.entries[right]

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
	z.fixupCount(prev, SENTINEL)
}

func (z *ZipTree[K, V]) fixupCount(curr, limit ZipNodeEntryIndex) {
	for curr != SENTINEL {
		var count uint32 = 1
		left, right := z.entries[curr].left, z.entries[curr].right
		if left != SENTINEL {
			count += z.entries[left].count
		}
		if right != SENTINEL {
			count += z.entries[right].count
		}
		z.entries[curr].count = count
		curr = z.entries[curr].parent
	}
}

// DisplayTree the tree in a human-readable way
func (z *ZipTree[K, V]) displayTree(rootIdx ZipNodeEntryIndex, prefix string, isLeft bool, hasBoth bool, sb *strings.Builder) {
	if rootIdx != SENTINEL {
		node := &z.entries[rootIdx]

		sb.WriteString(prefix)
		if isLeft && hasBoth {
			sb.WriteString("├── ")
		} else {
			sb.WriteString("└── ")
		}
		sb.WriteString(fmt.Sprintf("Idx: %d, ", rootIdx) + node.String() + fmt.Sprintf(", Parent: %d\n", node.parent))
		newPrefix := prefix
		if isLeft && hasBoth {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		nodeHasBoth := node.left != SENTINEL && node.right != SENTINEL
		z.displayTree(node.left, newPrefix, true, nodeHasBoth, sb)
		z.displayTree(node.right, newPrefix, false, nodeHasBoth, sb)
	}
}

func (z *ZipTree[K, V]) displayTreeNodesInOrder(rootIdx ZipNodeEntryIndex, sb *strings.Builder) {
	if rootIdx != SENTINEL {
		node := &z.entries[rootIdx]
		z.displayTreeNodesInOrder(node.left, sb)
		sb.WriteString(node.String() + "\n")
		z.displayTreeNodesInOrder(node.right, sb)
	}
}

func (z *ZipTree[K, V]) leftMost() ZipNodeEntryIndex {
	current := z.root
	for z.entries[current].left != SENTINEL {
		current = z.entries[current].left
	}
	return current
}

func (z *ZipTree[K, V]) rightMost() ZipNodeEntryIndex {
	current := z.root
	for z.entries[current].right != SENTINEL {
		current = z.entries[current].right
	}
	return current
}

func (z *ZipTree[K, V]) minimum() ZipNodeEntryIndex {
	if z.root == SENTINEL {
		return z.root
	}
	return z.leftMost()
}

func (z *ZipTree[K, V]) maximum() ZipNodeEntryIndex {
	if z.root == SENTINEL {
		return z.root
	}
	return z.rightMost()
}

func (z *ZipTree[K, V]) floor(key K) ZipNodeEntryIndex {
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

func (z *ZipTree[K, V]) ceiling(key K) ZipNodeEntryIndex {
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

func (z *ZipTree[K, V]) upperBound(key K) ZipNodeEntryIndex {
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

func (z *ZipTree[K, V]) lowerBound(key K) ZipNodeEntryIndex {
	return z.ceiling(key)
}

func (z *ZipTree[K, V]) atIndex(idx uint32) ZipNodeEntryIndex {
	root := z.root
	for root != SENTINEL {
		left := z.entries[root].left
		var leftCount = uint32(0)
		if left != SENTINEL {
			leftCount = z.entries[left].count
		}
		if idx < leftCount {
			root = left
		} else if idx > leftCount {
			root = z.entries[root].right
			idx = idx - leftCount - 1
		} else {
			break
		}
	}
	return root
}

func (z *ZipTree[K, V]) indexOf(key K) uint32 {
	root := z.root
	res := uint32(0)
	for root != SENTINEL {
		left := z.entries[root].left
		if z.comparator.LessThan(key, z.entries[root].key) {
			root = left
		} else {
			if left != SENTINEL {
				res += z.entries[left].count
			}
			if z.comparator.GreaterThan(key, z.entries[root].key) {
				res += 1
				root = z.entries[root].right
			} else {
				break
			}
		}
	}
	if root == SENTINEL {
		return ^uint32(0)
	}
	return res
}

func (z *ZipTree[K, V]) iterator(idx ZipNodeEntryIndex) *ZipIterator[K, V] {
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

// Floor Returns an iterator pointing to largest element in the BST less than or equal to key
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

func (z *ZipTree[K, V]) AtIndex(idx uint32) *ZipIterator[K, V] {
	return z.iterator(z.atIndex(idx))
}

func (z *ZipTree[K, V]) IndexOf(key K) uint32 {
	return z.indexOf(key)
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
	return z.deleteInternal(keyIdx)
}

// Delete returns true if entry was deleted
// returns false if entry not found
func (z *ZipTree[K, V]) Delete(key K) bool {
	keyIdx := z.find(key)
	return z.deleteInternal(keyIdx)
}

func (z *ZipTree[K, V]) DisplayTreeNodesInOrder() string {
	var sb strings.Builder
	z.displayTreeNodesInOrder(z.root, &sb)
	return sb.String()
}

func (z *ZipTree[K, V]) Size() int {
	return len(z.entries)
}

func (z *ZipTree[K, V]) Count() int {
	if z.root == SENTINEL {
		return 0
	} else {
		return int(z.entries[z.root].count)
	}
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
