package ziptree

type ZipIterator[K any] struct {
	current ZipNodeEntryIndex // index to the current node in the traversal
	entries []ZipNode[K]
}

func (it *ZipIterator[K]) IsEmpty() bool {
	return it.current == SENTINEL
}

func (it *ZipIterator[K]) Index() ZipNodeEntryIndex {
	return it.current
}

// Next move iterator forward
func (it *ZipIterator[K]) Next() {
	if it.IsEmpty() {
		return // Or handle error appropriately
	}
	root := it.current
	// Move to the next node
	if it.entries[root].right != SENTINEL {
		// If there's a right child, go to its leftmost descendant
		root = it.entries[root].right
		for it.entries[root].left != SENTINEL {
			root = it.entries[root].left
		}
		it.current = root
	} else {
		// If no right child, move up to the parent until we come from a left child
		parent := it.entries[root].parent
		for parent != SENTINEL && it.entries[parent].right == root {
			root = parent
			parent = it.entries[parent].parent
		}
		it.current = parent
	}
}

// Prev move iterator backwards
func (it *ZipIterator[K]) Prev() {
	if it.IsEmpty() {
		return // Or handle error appropriately
	}
	root := it.current
	// Move to the next node
	if it.entries[root].left != SENTINEL {
		// Find the rightmost node in the left subtree
		root = it.entries[root].left
		for it.entries[root].right != SENTINEL {
			root = it.entries[root].right
		}
		it.current = root
	} else {
		// Traverse up using parent pointers
		parent := it.entries[root].parent
		for parent != SENTINEL && it.entries[parent].left == root {
			root = parent
			parent = it.entries[parent].parent
		}
		it.current = parent
	}
}

func (it *ZipIterator[K]) Key() K {
	var ret K
	if it.current != SENTINEL {
		ret = it.entries[it.current].key
	}
	return ret
}

func (it *ZipIterator[K]) Parent() ZipNodeEntryIndex {
	ret := SENTINEL
	if it.current != SENTINEL {
		ret = it.entries[it.current].parent
	}
	return ret
}
