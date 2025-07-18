package ziptree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"math/rand/v2"
	"testing"
)

func TestZipTrees(t *testing.T) {
	arrays := [][]int32{
		{1, 3, 4, 6}, {9, 7, 8, 4}, {-10, 12, -99, 8, 7, 6, 1},
	}
	var empty = struct{}{}

	for _, arr := range arrays {
		n := len(arr)
		lo, hi := 0, n-1
		sortedArray := arr
		slices.Sort(sortedArray)
		tree := NewZipTree[int32, struct{}](func(a, b int32) bool {
			return a < b
		})
		for _, v := range sortedArray {
			tree.Insert(v, empty)
		}

		assert.Equal(t, sortedArray[lo], tree.Ceiling(sortedArray[lo]-1).Key())
		assert.Equal(t, SENTINEL, tree.Ceiling(sortedArray[hi]+1).Index())

		assert.Equal(t, sortedArray[lo], tree.UpperBound(sortedArray[lo]-1).Key())
		assert.Equal(t, SENTINEL, tree.UpperBound(sortedArray[hi]).Index())

		assert.Equal(t, sortedArray[hi], tree.Floor(sortedArray[hi]+1).Key())
		assert.Equal(t, SENTINEL, tree.Floor(sortedArray[lo]-1).Index())

		indices := make([]int, n)
		for i := range arr {
			indices[i] = i
		}
		rand.Shuffle(len(indices), func(i, j int) { indices[i], indices[j] = indices[j], indices[i] })

		for _, k := range indices {
			v := sortedArray[k]
			assert.Equal(t, v, tree.Find(v).Key())
			assert.Equal(t, v, tree.Ceiling(v).Key())
			assert.Equal(t, v, tree.Floor(v).Key())
			assert.Equal(t, v, tree.UpperBound(v-1).Key())

			if k != hi {
				assert.Equal(t, sortedArray[k+1], tree.UpperBound(v).Key())
			}
		}
	}
}

func TestZipTreeDeletion(t *testing.T) {
	tree := NewZipTree[int32, struct{}](func(a, b int32) bool {
		return a < b
	})

	var empty = struct{}{}
	key := int32(0)
	tree.Insert(key, empty)
	assert.Equal(t, 1, tree.Size())
	assert.Equal(t, key, tree.Minimum().Key())
	assert.Equal(t, key, tree.Maximum().Key())
	assert.Equal(t, true, tree.Delete(key))
	assert.Equal(t, 0, tree.Size())
	assert.Equal(t, SENTINEL, tree.Minimum().Index())
	assert.Equal(t, SENTINEL, tree.Maximum().Index())

	arr := []int32{6, 4, 3, 1}

	for _, v := range arr {
		tree.Insert(v, empty)
	}

	slices.Sort(arr)

	sortedArray := make([]int32, len(arr))
	iter := tree.NewIterator()
	for i := 0; i < len(arr); i++ {
		sortedArray[i] = iter.Key()
		iter.Next()
	}
	assert.Equal(t, arr, sortedArray)
	assert.True(t, iter.IsEmpty())

	iter = tree.NewPrevIterator()
	for i := 0; i < len(arr); i++ {
		sortedArray[i] = iter.Key()
		iter.Prev()
	}
	slices.Reverse(arr)
	assert.Equal(t, arr, sortedArray)
	assert.True(t, iter.IsEmpty())

	assert.Equal(t, 4, tree.Size())
	tree.DeleteIter(tree.Minimum())
	tree.DeleteIter(tree.Maximum())
	assert.Equal(t, 2, tree.Size())
	assert.Equal(t, arr[2], tree.Minimum().Key())
	assert.Equal(t, arr[1], tree.Maximum().Key())

	// stays same size since no node inserted
	assert.Equal(t, false, tree.Insert(arr[1], empty))
	assert.Equal(t, false, tree.Insert(arr[2], empty))
	assert.Equal(t, 2, tree.Size())
}

func TestZipTreeDisplayAndSize(t *testing.T) {
	tree := NewZipTreeWithRandomGenerator[int32, struct{}](func(a, b int32) bool {
		return a < b
	}, rand.New(rand.NewPCG(123, 456)))

	var empty = struct{}{}
	tree.Insert(3, empty)
	tree.Insert(1, empty)
	tree.Insert(2, empty)

	assert.Equal(t, true, tree.Delete(1))
	assert.Equal(t, false, tree.Delete(11))

	tree.Insert(6, empty)
	tree.Insert(8, empty)
	tree.Insert(1, empty)
	tree.Insert(2, empty)
	tree.Insert(8, empty)
	tree.Insert(9, empty)
	tree.Insert(17, empty)
	tree.Insert(-12, empty)
	tree.Insert(-33, empty)

	assert.Equal(t, 9, tree.Size())
	assert.Equal(t, true, tree.Insert(222, empty))
	assert.Equal(t, tree.Size(), 10)

	expected := `└── Idx: 2, Key: 6,  Value: {}, Rank: (5, 1), Parent: 4294967295
    ├── Idx: 7, Key: -12,  Value: {}, Rank: (3, 21), Parent: 2
    │   ├── Idx: 8, Key: -33,  Value: {}, Rank: (0, 2), Parent: 7
    │   └── Idx: 0, Key: 3,  Value: {}, Rank: (2, 1), Parent: 7
    │       └── Idx: 4, Key: 1,  Value: {}, Rank: (0, 8), Parent: 0
    │           └── Idx: 1, Key: 2,  Value: {}, Rank: (0, 1), Parent: 4
    └── Idx: 6, Key: 17,  Value: {}, Rank: (4, 5), Parent: 2
        ├── Idx: 5, Key: 9,  Value: {}, Rank: (0, 7), Parent: 6
        │   └── Idx: 3, Key: 8,  Value: {}, Rank: (0, 5), Parent: 5
        └── Idx: 9, Key: 222,  Value: {}, Rank: (0, 1), Parent: 6
`
	assert.Equal(t, expected, tree.String())

	// test ordered display
	expected = `Key: -33,  Value: {}, Rank: (0, 2)
Key: -12,  Value: {}, Rank: (3, 21)
Key: 1,  Value: {}, Rank: (0, 8)
Key: 2,  Value: {}, Rank: (0, 1)
Key: 3,  Value: {}, Rank: (2, 1)
Key: 6,  Value: {}, Rank: (5, 1)
Key: 8,  Value: {}, Rank: (0, 5)
Key: 9,  Value: {}, Rank: (0, 7)
Key: 17,  Value: {}, Rank: (4, 5)
Key: 222,  Value: {}, Rank: (0, 1)
`
	assert.Equal(t, expected, tree.DisplayTreeNodesInOrder())

	// inert two new values
	tree.Insert(12, empty)
	tree.Insert(0, empty)
	assert.Equal(t, 12, tree.Size())

	// test iterations

	orderedNodes := ""
	iter := tree.NewIterator()
	for !iter.IsEmpty() {
		current := iter.Index()
		key, value := iter.Key(), iter.Value()
		parent := iter.Parent()
		orderedNodes += fmt.Sprintf("Key: %v, Idx: %d, Value: %v, Parent: %d\n", key, current, value, parent)
		iter.Next()
	}
	expected = "Key: -33, Idx: 8, Value: {}, Parent: 7\nKey: -12, Idx: 7, Value: {}, Parent: 2\nKey: 0, Idx: 11, Value: {}, Parent: 7\nKey: 1, Idx: 4, Value: {}, Parent: 0\nKey: 2, Idx: 1, Value: {}, Parent: 4\nKey: 3, Idx: 0, Value: {}, Parent: 11\nKey: 6, Idx: 2, Value: {}, Parent: 4294967295\nKey: 8, Idx: 3, Value: {}, Parent: 5\nKey: 9, Idx: 5, Value: {}, Parent: 10\nKey: 12, Idx: 10, Value: {}, Parent: 6\nKey: 17, Idx: 6, Value: {}, Parent: 2\nKey: 222, Idx: 9, Value: {}, Parent: 6\n"
	assert.Equal(t, expected, orderedNodes)

	// backwards
	orderedNodes = ""
	iter = tree.NewPrevIterator()
	for !iter.IsEmpty() {
		current := iter.Index()
		key, value := iter.Key(), iter.Value()
		parent := iter.Parent()
		orderedNodes += fmt.Sprintf("Key: %v, Idx: %d, Value: %v, Parent: %d\n", key, current, value, parent)
		iter.Prev()
	}

	expected = "Key: 222, Idx: 9, Value: {}, Parent: 6\nKey: 17, Idx: 6, Value: {}, Parent: 2\nKey: 12, Idx: 10, Value: {}, Parent: 6\nKey: 9, Idx: 5, Value: {}, Parent: 10\nKey: 8, Idx: 3, Value: {}, Parent: 5\nKey: 6, Idx: 2, Value: {}, Parent: 4294967295\nKey: 3, Idx: 0, Value: {}, Parent: 11\nKey: 2, Idx: 1, Value: {}, Parent: 4\nKey: 1, Idx: 4, Value: {}, Parent: 0\nKey: 0, Idx: 11, Value: {}, Parent: 7\nKey: -12, Idx: 7, Value: {}, Parent: 2\nKey: -33, Idx: 8, Value: {}, Parent: 7\n"
	assert.Equal(t, expected, orderedNodes)
	tree.Delete(tree.Maximum().Key())
	assert.Equal(t, int32(17), tree.Maximum().Key())
}
