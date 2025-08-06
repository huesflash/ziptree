package ziptree

import (
	"fmt"
	"math/bits"
	"math/rand/v2"
	"strings"
)

type ZipNodeEntryIndex uint32

const SENTINEL = ^ZipNodeEntryIndex(0)

type ZipNode[K any] struct {
	key                 K
	left, right, parent ZipNodeEntryIndex
	rank, count         uint32
}

type ZipTree[K any] struct {
	entries         []ZipNode[K]
	root            ZipNodeEntryIndex
	lessThan        LessFn[K]
	randomGenerator *rand.Rand
}

type LessFn[T any] func(a, b T) bool

func (zn *ZipNode[K]) String() string {
	return fmt.Sprintf("Key: %v, Rank: (%d, %d), Count: %d", zn.key, zn.rank>>16, zn.rank&0x0000ffff, zn.count)
}

func (z *ZipTree[K]) String() string {
	var sb strings.Builder
	z.displayTree(z.root, "", false, false, &sb)
	return sb.String()
}

func (z *ZipTree[K]) find(key K) ZipNodeEntryIndex {
	root := z.root
	for root != SENTINEL {
		if z.lessThan(key, z.entries[root].key) {
			root = z.entries[root].left
		} else if z.lessThan(z.entries[root].key, key) { // b < a == a > b
			root = z.entries[root].right
		} else {
			break
		}
	}
	return root
}

func (z *ZipTree[K]) insert(key K) {
	rootIdx := z.root
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
	z.entries = append(z.entries, ZipNode[K]{
		key:    key,
		rank:   rank,
		left:   SENTINEL,
		right:  SENTINEL,
		parent: SENTINEL,
		count:  1,
	})
	curr := rootIdx
	prev := SENTINEL
	for curr != SENTINEL && (rank < z.entries[curr].rank || (rank == z.entries[curr].rank &&
		z.lessThan(z.entries[curr].key, key))) { // b < a == a > b
		prev = curr
		if z.lessThan(key, z.entries[curr].key) {
			curr = z.entries[curr].left
		} else {
			curr = z.entries[curr].right
		}
	}
	if curr == rootIdx {
		z.root = idx
	} else {
		if z.lessThan(key, z.entries[prev].key) {
			z.entries[prev].left = idx
		} else {
			z.entries[prev].right = idx
		}
		z.entries[idx].parent = prev
	}

	if curr != SENTINEL {
		if z.lessThan(key, z.entries[curr].key) {
			z.entries[idx].right = curr
		} else {
			z.entries[idx].left = curr
		}
		z.entries[curr].parent = idx
		prev = idx
		for curr != SENTINEL {
			fix := prev
			if z.lessThan(z.entries[curr].key, key) { // b < a  == a > b
				for curr != SENTINEL && !z.lessThan(key, z.entries[curr].key) { // !(a < b) == a >= b
					prev = curr
					curr = z.entries[curr].right
				}
			} else {
				for curr != SENTINEL && !z.lessThan(z.entries[curr].key, key) { // !(b < a) == a <= b
					prev = curr
					curr = z.entries[curr].left
				}
			}

			if z.lessThan(key, z.entries[fix].key) || (fix == idx && z.lessThan(key, z.entries[prev].key)) {
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

func (z *ZipTree[K]) compact(keyIdx ZipNodeEntryIndex) {
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

func (z *ZipTree[K]) deleteInternal(keyIdx ZipNodeEntryIndex) bool {
	if keyIdx == SENTINEL {
		return false
	}
	z.deleteIndex(keyIdx)
	z.compact(keyIdx)
	return true
}

func (z *ZipTree[K]) deleteIndex(keyIdx ZipNodeEntryIndex) {
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
		if z.lessThan(key, z.entries[prev].key) {
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

func (z *ZipTree[K]) fixupCount(curr, limit ZipNodeEntryIndex) {
	for curr != limit {
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
func (z *ZipTree[K]) displayTree(rootIdx ZipNodeEntryIndex, prefix string, isLeft bool, hasBoth bool, sb *strings.Builder) {
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

func (z *ZipTree[K]) displayTreeNodesInOrder(sb *strings.Builder) {
	iter := z.NewIterator()
	for !iter.IsEmpty() {
		current := iter.Index()
		sb.WriteString(z.entries[current].String() + "\n")
		iter.Next()
	}
}

func (z *ZipTree[K]) leftMost() ZipNodeEntryIndex {
	current := z.root
	for z.entries[current].left != SENTINEL {
		current = z.entries[current].left
	}
	return current
}

func (z *ZipTree[K]) rightMost() ZipNodeEntryIndex {
	current := z.root
	for z.entries[current].right != SENTINEL {
		current = z.entries[current].right
	}
	return current
}

func (z *ZipTree[K]) minimum() ZipNodeEntryIndex {
	if z.root == SENTINEL {
		return z.root
	}
	return z.leftMost()
}

func (z *ZipTree[K]) maximum() ZipNodeEntryIndex {
	if z.root == SENTINEL {
		return z.root
	}
	return z.rightMost()
}

func (z *ZipTree[K]) floor(key K) ZipNodeEntryIndex {
	res := SENTINEL
	root := z.root

	for root != SENTINEL {
		if z.lessThan(key, z.entries[root].key) {
			root = z.entries[root].left
		} else {
			res = root
			root = z.entries[root].right
		}
	}
	return res
}

func (z *ZipTree[K]) ceiling(key K) ZipNodeEntryIndex {
	res := SENTINEL
	root := z.root

	for root != SENTINEL {
		if z.lessThan(z.entries[root].key, key) { // b < a == a > b
			root = z.entries[root].right
		} else {
			res = root
			root = z.entries[root].left
		}
	}
	return res
}

func (z *ZipTree[K]) upperBound(key K) ZipNodeEntryIndex {
	res := SENTINEL
	root := z.root

	for root != SENTINEL {
		if !z.lessThan(key, z.entries[root].key) { // !(a < b) == a >= b
			root = z.entries[root].right
		} else {
			res = root
			root = z.entries[root].left
		}
	}
	return res
}

func (z *ZipTree[K]) lowerBound(key K) ZipNodeEntryIndex {
	return z.ceiling(key)
}

func (z *ZipTree[K]) atIndex(idx uint32) ZipNodeEntryIndex {
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

func (z *ZipTree[K]) indexOf(key K) uint32 {
	root := z.root
	res := uint32(0)
	for root != SENTINEL {
		left := z.entries[root].left
		if z.lessThan(key, z.entries[root].key) {
			root = left
		} else {
			if left != SENTINEL {
				res += z.entries[left].count
			}
			if z.lessThan(z.entries[root].key, key) { // b < a == a > b
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

func (z *ZipTree[K]) iterator(idx ZipNodeEntryIndex) *ZipIterator[K] {
	if idx == SENTINEL {
		return &ZipIterator[K]{
			current: SENTINEL,
		}
	} else {
		return &ZipIterator[K]{
			current: idx,
			entries: z.entries,
		}
	}
}

func NewZipTree[K any](less LessFn[K]) *ZipTree[K] {
	return &ZipTree[K]{
		entries:         make([]ZipNode[K], 0),
		root:            SENTINEL,
		lessThan:        less,
		randomGenerator: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

func NewZipTreeWithRandomGenerator[K any](less LessFn[K], randomGenerator *rand.Rand) *ZipTree[K] {
	return &ZipTree[K]{
		entries:         make([]ZipNode[K], 0),
		root:            SENTINEL,
		lessThan:        less,
		randomGenerator: randomGenerator,
	}
}

// Ceiling Returns an iterator pointing to the largest element in the BST greater than or equal to key
func (z *ZipTree[K]) Ceiling(key K) *ZipIterator[K] {
	return z.iterator(z.ceiling(key))

}

// Floor Returns an iterator pointing to largest element in the BST less than or equal to key
func (z *ZipTree[K]) Floor(key K) *ZipIterator[K] {
	return z.iterator(z.floor(key))
}

// UpperBound Returns an iterator pointing to the first element in the tree which is ordered after key
func (z *ZipTree[K]) UpperBound(key K) *ZipIterator[K] {
	return z.iterator(z.upperBound(key))
}

// LowerBound Returns an iterator pointing to the first element in the tree which is not ordered before key
func (z *ZipTree[K]) LowerBound(key K) *ZipIterator[K] {
	return z.iterator(z.lowerBound(key))
}

func (z *ZipTree[K]) Find(key K) *ZipIterator[K] {
	return z.iterator(z.find(key))
}

func (z *ZipTree[K]) Minimum() *ZipIterator[K] {
	return z.iterator(z.minimum())
}

func (z *ZipTree[K]) Maximum() *ZipIterator[K] {
	return z.iterator(z.maximum())
}

func (z *ZipTree[K]) AtIndex(idx uint32) *ZipIterator[K] {
	return z.iterator(z.atIndex(idx))
}

func (z *ZipTree[K]) IndexOf(key K) uint32 {
	return z.indexOf(key)
}

// Insert returns true if entry was inserted,
// returns false to indicate update
func (z *ZipTree[K]) Insert(key K) bool {
	found := z.find(key)
	if found == SENTINEL {
		z.insert(key)
		return true
	} else {
		return false
	}
}

// DeleteIter returns true if entry was deleted
// returns false if entry not found
func (z *ZipTree[K]) DeleteIter(iter *ZipIterator[K]) bool {
	keyIdx := iter.Index()
	return z.deleteInternal(keyIdx)
}

// Delete returns true if entry was deleted
// returns false if entry not found
func (z *ZipTree[K]) Delete(key K) bool {
	keyIdx := z.find(key)
	return z.deleteInternal(keyIdx)
}

func (z *ZipTree[K]) DisplayTreeNodesInOrder() string {
	var sb strings.Builder
	z.displayTreeNodesInOrder(&sb)
	return sb.String()
}

func (z *ZipTree[K]) Size() int {
	return len(z.entries)
}

func (z *ZipTree[K]) Count() int {
	if z.root == SENTINEL {
		return 0
	} else {
		return int(z.entries[z.root].count)
	}
}

func (z *ZipTree[K]) NewIterator() *ZipIterator[K] {
	iter := &ZipIterator[K]{}
	if z.root == SENTINEL {
		iter.current = SENTINEL
		return iter
	}
	// Find the leftmost node
	iter.current = z.leftMost()
	iter.entries = z.entries
	return iter
}

func (z *ZipTree[K]) NewPrevIterator() *ZipIterator[K] {
	iter := &ZipIterator[K]{}
	if z.root == SENTINEL {
		iter.current = SENTINEL
		return iter
	}
	// Find the right most node
	iter.current = z.rightMost()
	iter.entries = z.entries
	return iter
}
