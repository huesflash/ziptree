package ziptree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"math/rand/v2"
	"testing"
	"time"
)

func TestZipTrees(t *testing.T) {
	arrays := [][]int32{
		{1, 3, 4, 6}, {9, 7, 8, 4}, {-10, 12, -99, 8, 3, 7, 4, 6, 1, 2},
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
	t.Run("test simple delete", func(t *testing.T) {
		tree.Insert(3, empty)
		tree.Insert(1, empty)
		tree.Insert(2, empty)
		assert.Equal(t, true, tree.Delete(1))
		assert.Equal(t, false, tree.Delete(11))
	})

	t.Run("test simple insert", func(t *testing.T) {
		treeValues := []int32{6, 8, 1, 2, 8, 9, 17, -12, -33}
		for _, v := range treeValues {
			tree.Insert(v, empty)
		}

		assert.Equal(t, 9, tree.Size())
		assert.Equal(t, true, tree.Insert(222, empty))
		assert.Equal(t, tree.Size(), 10)
	})

	t.Run("test display tree", func(t *testing.T) {
		expected := `└── Idx: 2, Key: 6,  Value: {}, Rank: (5, 1), Count: 10, Parent: 4294967295
    ├── Idx: 7, Key: -12,  Value: {}, Rank: (3, 21), Count: 5, Parent: 2
    │   ├── Idx: 8, Key: -33,  Value: {}, Rank: (0, 2), Count: 1, Parent: 7
    │   └── Idx: 0, Key: 3,  Value: {}, Rank: (2, 1), Count: 3, Parent: 7
    │       └── Idx: 4, Key: 1,  Value: {}, Rank: (0, 8), Count: 2, Parent: 0
    │           └── Idx: 1, Key: 2,  Value: {}, Rank: (0, 1), Count: 1, Parent: 4
    └── Idx: 6, Key: 17,  Value: {}, Rank: (4, 5), Count: 4, Parent: 2
        ├── Idx: 5, Key: 9,  Value: {}, Rank: (0, 7), Count: 2, Parent: 6
        │   └── Idx: 3, Key: 8,  Value: {}, Rank: (0, 5), Count: 1, Parent: 5
        └── Idx: 9, Key: 222,  Value: {}, Rank: (0, 1), Count: 1, Parent: 6
`
		assert.Equal(t, expected, tree.String())
	})

	t.Run("test ordered display", func(t *testing.T) {
		// test ordered display
		expected := `Key: -33,  Value: {}, Rank: (0, 2), Count: 1
Key: -12,  Value: {}, Rank: (3, 21), Count: 5
Key: 1,  Value: {}, Rank: (0, 8), Count: 2
Key: 2,  Value: {}, Rank: (0, 1), Count: 1
Key: 3,  Value: {}, Rank: (2, 1), Count: 3
Key: 6,  Value: {}, Rank: (5, 1), Count: 10
Key: 8,  Value: {}, Rank: (0, 5), Count: 1
Key: 9,  Value: {}, Rank: (0, 7), Count: 2
Key: 17,  Value: {}, Rank: (4, 5), Count: 4
Key: 222,  Value: {}, Rank: (0, 1), Count: 1
`
		assert.Equal(t, expected, tree.DisplayTreeNodesInOrder())
	})

	t.Run("test iterations", func(t *testing.T) {
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
		expected := "Key: -33, Idx: 8, Value: {}, Parent: 7\nKey: -12, Idx: 7, Value: {}, Parent: 2\nKey: 1, Idx: 4, Value: {}, Parent: 0\nKey: 2, Idx: 1, Value: {}, Parent: 4\nKey: 3, Idx: 0, Value: {}, Parent: 7\nKey: 6, Idx: 2, Value: {}, Parent: 4294967295\nKey: 8, Idx: 3, Value: {}, Parent: 5\nKey: 9, Idx: 5, Value: {}, Parent: 6\nKey: 17, Idx: 6, Value: {}, Parent: 2\nKey: 222, Idx: 9, Value: {}, Parent: 6\n"
		assert.Equal(t, expected, orderedNodes)

	})

	t.Run("test backward iterations", func(t *testing.T) {
		// backwards
		orderedNodes := ""
		iter := tree.NewPrevIterator()
		for !iter.IsEmpty() {
			current := iter.Index()
			key, value := iter.Key(), iter.Value()
			parent := iter.Parent()
			orderedNodes += fmt.Sprintf("Key: %v, Idx: %d, Value: %v, Parent: %d\n", key, current, value, parent)
			iter.Prev()
		}

		expected := "Key: 222, Idx: 9, Value: {}, Parent: 6\nKey: 17, Idx: 6, Value: {}, Parent: 2\nKey: 9, Idx: 5, Value: {}, Parent: 6\nKey: 8, Idx: 3, Value: {}, Parent: 5\nKey: 6, Idx: 2, Value: {}, Parent: 4294967295\nKey: 3, Idx: 0, Value: {}, Parent: 7\nKey: 2, Idx: 1, Value: {}, Parent: 4\nKey: 1, Idx: 4, Value: {}, Parent: 0\nKey: -12, Idx: 7, Value: {}, Parent: 2\nKey: -33, Idx: 8, Value: {}, Parent: 7\n"
		assert.Equal(t, expected, orderedNodes)
	})

	t.Run("delete root", func(t *testing.T) {
		tree.Delete(6)
		expected := `└── Idx: 6, Key: 17,  Value: {}, Rank: (4, 5), Count: 9, Parent: 4294967295
    ├── Idx: 7, Key: -12,  Value: {}, Rank: (3, 21), Count: 7, Parent: 6
    │   ├── Idx: 8, Key: -33,  Value: {}, Rank: (0, 2), Count: 1, Parent: 7
    │   └── Idx: 0, Key: 3,  Value: {}, Rank: (2, 1), Count: 5, Parent: 7
    │       ├── Idx: 4, Key: 1,  Value: {}, Rank: (0, 8), Count: 2, Parent: 0
    │       │   └── Idx: 1, Key: 2,  Value: {}, Rank: (0, 1), Count: 1, Parent: 4
    │       └── Idx: 5, Key: 9,  Value: {}, Rank: (0, 7), Count: 2, Parent: 0
    │           └── Idx: 3, Key: 8,  Value: {}, Rank: (0, 5), Count: 1, Parent: 5
    └── Idx: 2, Key: 222,  Value: {}, Rank: (0, 1), Count: 1, Parent: 6
`
		assert.Equal(t, expected, tree.String())
		var orderedNodes []int32
		iter := tree.NewIterator()
		for !iter.IsEmpty() {
			orderedNodes = append(orderedNodes, iter.Key())
			iter.Next()
		}
		expectedKeyes := []int32{
			-33, -12, 1, 2, 3, 8, 9, 17, 222,
		}
		assert.Equal(t, expectedKeyes, orderedNodes)
		assert.Equal(t, tree.Size(), tree.Count())
	})

	t.Run("delete leaf", func(t *testing.T) {
		tree.Delete(8)
		expected := `└── Idx: 6, Key: 17,  Value: {}, Rank: (4, 5), Count: 8, Parent: 4294967295
    ├── Idx: 7, Key: -12,  Value: {}, Rank: (3, 21), Count: 6, Parent: 6
    │   ├── Idx: 3, Key: -33,  Value: {}, Rank: (0, 2), Count: 1, Parent: 7
    │   └── Idx: 0, Key: 3,  Value: {}, Rank: (2, 1), Count: 4, Parent: 7
    │       ├── Idx: 4, Key: 1,  Value: {}, Rank: (0, 8), Count: 2, Parent: 0
    │       │   └── Idx: 1, Key: 2,  Value: {}, Rank: (0, 1), Count: 1, Parent: 4
    │       └── Idx: 5, Key: 9,  Value: {}, Rank: (0, 7), Count: 1, Parent: 0
    └── Idx: 2, Key: 222,  Value: {}, Rank: (0, 1), Count: 1, Parent: 6
`
		assert.Equal(t, expected, tree.String())
		var orderedNodes []int32
		iter := tree.NewIterator()
		for !iter.IsEmpty() {
			orderedNodes = append(orderedNodes, iter.Key())
			iter.Next()
		}
		expectedKeyes := []int32{
			-33, -12, 1, 2, 3, 9, 17, 222,
		}
		assert.Equal(t, expectedKeyes, orderedNodes)
		assert.Equal(t, tree.Size(), tree.Count())
	})

	t.Run("delete interior node", func(t *testing.T) {
		tree.Delete(3)
		expected := `└── Idx: 6, Key: 17,  Value: {}, Rank: (4, 5), Count: 7, Parent: 4294967295
    ├── Idx: 0, Key: -12,  Value: {}, Rank: (3, 21), Count: 5, Parent: 6
    │   ├── Idx: 3, Key: -33,  Value: {}, Rank: (0, 2), Count: 1, Parent: 0
    │   └── Idx: 4, Key: 1,  Value: {}, Rank: (0, 8), Count: 3, Parent: 0
    │       └── Idx: 5, Key: 9,  Value: {}, Rank: (0, 7), Count: 2, Parent: 4
    │           └── Idx: 1, Key: 2,  Value: {}, Rank: (0, 1), Count: 1, Parent: 5
    └── Idx: 2, Key: 222,  Value: {}, Rank: (0, 1), Count: 1, Parent: 6
`
		assert.Equal(t, expected, tree.String())
		var orderedNodes []int32
		iter := tree.NewIterator()
		for !iter.IsEmpty() {
			orderedNodes = append(orderedNodes, iter.Key())
			iter.Next()
		}
		expectedKeyes := []int32{
			-33, -12, 1, 2, 9, 17, 222,
		}
		assert.Equal(t, expectedKeyes, orderedNodes)
		assert.Equal(t, tree.Size(), tree.Count())
	})
}

func checkOrderedNodes(t *testing.T, tree *ZipTree[int32, struct{}]) {
	assert.Equal(t, tree.Size(), tree.Count())
	maxKey := tree.Maximum().Key()
	maxIdx := tree.IndexOf(maxKey)
	assert.Equal(t, maxKey, tree.AtIndex(maxIdx).Key())
	minKey := tree.Minimum().Key()
	minIdx := tree.IndexOf(minKey)
	assert.Equal(t, minKey, tree.AtIndex(minIdx).Key())
	iter := tree.NewIterator()
	var orderedNodes []int32
	for !iter.IsEmpty() {
		idx := tree.IndexOf(iter.Key())
		assert.Equal(t, iter.Key(), tree.AtIndex(idx).Key())
		orderedNodes = append(orderedNodes, iter.Key())
		iter.Next()
	}
	assert.True(t, slices.IsSorted(orderedNodes))
}

func TestArrayIndexIndex(t *testing.T) {
	treeValues := [][]struct {
		value  int32
		delete bool
	}{
		{{1, false}, {14, false}, {-1, false}, {27, true}, {14, false}, {9, false}, {5, false}, {3, false}, {14, false}, {1, false}, {5, true}, {14, false}, {34, false}, {97, true}, {99, false}},
		{{8, false}, {2, false}, {-11, false}, {14, true}, {19, false}, {20, false}, {20, true}, {20, false}, {19, true}, {8, false}, {4, false}, {2, false}},
		{{1, false}, {12, false}, {-1, true}, {47, true}, {16, false}, {-1, true}, {14, false}, {14, false}, {14, false}, {14, false}, {12, false}, {18, false}},
	}
	var empty = struct{}{}
	for _, treeValue := range treeValues {
		tree := NewZipTreeWithRandomGenerator[int32, struct{}](func(a, b int32) bool {
			return a < b
		}, rand.New(rand.NewPCG(123, 456)))
		for _, entry := range treeValue {
			if entry.delete {
				tree.Delete(entry.value)
			} else {
				tree.Insert(entry.value, empty)
			}
		}
		checkOrderedNodes(t, tree)
	}

	t.Run("index on random generated", func(t *testing.T) {
		gen := rand.New(rand.NewPCG(uint64(time.Now().Unix()), uint64(time.Now().Add(time.Second*time.Duration(rand.Int32N(120))).Unix())))
		for k := 0; k < 10; k++ {
			tree := NewZipTreeWithRandomGenerator[int32, struct{}](func(a, b int32) bool {
				return a < b
			}, gen)
			n := int32(997)
			for i := 0; i < int(n); i++ {
				tree.Insert(-5777+gen.Int32N(7987), empty)
			}

			for i := 0; i < int(n/2); i++ {
				idx := uint32(rand.Int32N(n))
				if rand.Int32N(2) == 0 {
					tree.Delete(tree.AtIndex(idx).Key())
					if rand.Int32N(2) == 0 {
						tree.Insert(-879+gen.Int32N(9876), empty)
					}
					checkOrderedNodes(t, tree)
				}
			}

			iter := tree.NewIterator()
			var orderedNodes []int32
			for !iter.IsEmpty() {
				idx := tree.IndexOf(iter.Key())
				assert.Equal(t, iter.Key(), tree.AtIndex(idx).Key())
				orderedNodes = append(orderedNodes, iter.Key())
				iter.Next()
			}
			assert.True(t, slices.IsSorted(orderedNodes))
			assert.Equal(t, tree.Size(), tree.Count())

			for _, value := range orderedNodes {
				tree.Delete(value)
				assert.Equal(t, tree.Size(), tree.Count())
				iterAfterDeletion := tree.NewIterator()
				for !iterAfterDeletion.IsEmpty() {
					idx := tree.IndexOf(iterAfterDeletion.Key())
					assert.Equal(t, iterAfterDeletion.Key(), tree.AtIndex(idx).Key())
					iterAfterDeletion.Next()
				}
				checkOrderedNodes(t, tree)
			}
		}
	})
	t.Run("not found", func(t *testing.T) {
		tree := NewZipTreeWithRandomGenerator[int32, struct{}](func(a, b int32) bool {
			return a < b
		}, rand.New(rand.NewPCG(123, 456)))
		treeValues := []int32{6, 8, 1, 2, 8, 9, 17, -12, -33}
		for _, v := range treeValues {
			tree.Insert(v, empty)
		}
		assert.Equal(t, ^uint32(0), tree.IndexOf(int32(-34)))
		assert.Equal(t, ^uint32(0), tree.IndexOf(int32(34)))
		assert.Equal(t, SENTINEL, tree.AtIndex(uint32(34)).Key())
	})
}
