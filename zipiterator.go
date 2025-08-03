package ziptree

type ZipIterator[K, V any] struct {
	current ZipNodeEntryIndex // index to the current node in the traversal
	entries []ZipNode[K, V]
}

func (it *ZipIterator[K, V]) IsEmpty() bool {
	return it.current == SENTINEL
}

func (it *ZipIterator[K, V]) Index() ZipNodeEntryIndex {
	return it.current
}

// Next move iterator forward
func (it *ZipIterator[K, V]) Next() {
	if it.IsEmpty() {
		return // Or handle error appropriately
	}
	// Move to the next node
	if it.entries[it.current].right != SENTINEL {
		// If there's a right child, go to its leftmost descendant
		it.current = it.entries[it.current].right
		for it.entries[it.current].left != SENTINEL {
			it.current = it.entries[it.current].left
		}
	} else {
		// If no right child, move up to the parent until we come from a left child
		parent := it.entries[it.current].parent
		for parent != SENTINEL && it.entries[parent].right == it.current {
			it.current = parent
			parent = it.entries[parent].parent
		}
		it.current = parent
	}
}

// Prev move iterator backwards
func (it *ZipIterator[K, V]) Prev() {
	if it.IsEmpty() {
		return // Or handle error appropriately
	}

	// Move to the next node
	if it.entries[it.current].left != SENTINEL {
		// Find the rightmost node in the left subtree
		it.current = it.entries[it.current].left
		for it.entries[it.current].right != SENTINEL {
			it.current = it.entries[it.current].right
		}
	} else {
		// Traverse up using parent pointers
		parent := it.entries[it.current].parent
		for parent != SENTINEL && it.entries[parent].left == it.current {
			it.current = parent
			parent = it.entries[parent].parent
		}
		it.current = parent
	}
}

func (it *ZipIterator[K, V]) Value() V {
	var ret V
	if it.current != SENTINEL {
		ret = it.entries[it.current].data
	}
	return ret
}

func (it *ZipIterator[K, V]) Key() K {
	var ret K
	if it.current != SENTINEL {
		ret = it.entries[it.current].key
	}
	return ret
}

func (it *ZipIterator[K, V]) Parent() ZipNodeEntryIndex {
	ret := SENTINEL
	if it.current != SENTINEL {
		ret = it.entries[it.current].parent
	}
	return ret
}
