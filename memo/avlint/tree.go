// Package avlint provides an interval tree backed by an AVL tree. In addition,
// the interval tree supports shifting intervals in amortized constant time
// using lazy shifts.
package avlint

// ShiftThreshold is the number of shifts to accumulate before applying all
// shifts.
const ShiftThreshold = 1024

// A Key stores the start position of an interval, and a unique ID if you would
// like to store multiple intervals starting from the same position. The Key is
// used for uniquely identifying a particular interval when searching or
// removing from the tree.
type Key struct {
	Val int
	Id  int
}

func (k Key) Compare(other Key) int {
	if k.Val < other.Val {
		return -1
	} else if k.Val > other.Val {
		return 1
	} else if k.Id < other.Id {
		return -1
	} else if k.Id > other.Id {
		return 1
	}
	return 0
}

type shift struct {
	idx    int
	amt    int
	tstamp uint64
}

type Tree struct {
	root   *node
	shifts []shift
	tstamp uint64
}

func NewTree() *Tree {
	return &Tree{
		shifts: make([]shift, 0),
	}
}

func (t *Tree) Add(key Key, value Interval) {
	t.root = t.root.add(t, key, value)
}

func (t *Tree) Remove(key Key) {
	t.root = t.root.remove(key)
}

func (t *Tree) Search(key Key) Interval {
	return t.root.search(key).interval
}

func (t *Tree) Overlap(i Interval) []Interval {
	result := make([]Interval, 0)
	return t.root.overlap(i, result)
}

func (t *Tree) Height() int {
	return t.root.getHeight()
}

func (t *Tree) Shift(idx, amt int) {
	t.tstamp++
	t.shifts = append(t.shifts, shift{
		idx:    idx,
		amt:    amt,
		tstamp: t.tstamp,
	})
	if len(t.shifts) >= ShiftThreshold {
		t.applyAllShifts()
	}
}

func (t *Tree) applyAllShifts() {
	t.root.applyAllShifts()
}

type node struct {
	key      Key
	max      int
	interval Interval
	tstamp   uint64 // timestamp to determine which shifts to apply
	tree     *Tree

	// height counts nodes (not edges)
	height int
	left   *node
	right  *node
}

// Adds a new node
func (n *node) add(tree *Tree, key Key, value Interval) *node {
	if n == nil {
		return &node{
			tree:     tree,
			key:      key,
			max:      value.High(),
			interval: value,
			height:   1,
			left:     nil,
			right:    nil,
		}
	}

	if key.Compare(n.key) < 0 {
		n.left = n.left.add(tree, key, value)
	} else if key.Compare(n.key) > 0 {
		n.right = n.right.add(tree, key, value)
	} else {
		// if same key exists update value
		n.interval = value
	}
	return n.rebalanceTree()
}

func (n *node) updateMax() {
	if n != nil {
		if n.right != nil {
			n.max = max(n.max, n.right.max)
		}
		if n.left != nil {
			n.max = max(n.max, n.left.max)
		}
	}
}

// Removes a node
func (n *node) remove(key Key) *node {
	if n == nil {
		return nil
	}
	n.applyShifts()
	if key.Compare(n.key) < 0 {
		n.left = n.left.remove(key)
	} else if key.Compare(n.key) > 0 {
		n.right = n.right.remove(key)
	} else {
		if n.left != nil && n.right != nil {
			// node to delete found with both children;
			// replace values with smallest node of the right sub-tree
			rightMinNode := n.right.findSmallest()
			n.key = rightMinNode.key
			n.interval = rightMinNode.interval
			// delete smallest node that we replaced
			n.right = n.right.remove(rightMinNode.key)
		} else if n.left != nil {
			// node only has left child
			n = n.left
		} else if n.right != nil {
			// node only has right child
			n = n.right
		} else {
			// node has no children
			n = nil
			return n
		}

	}
	return n.rebalanceTree()
}

// Searches for a node
func (n *node) search(key Key) *node {
	if n == nil {
		return nil
	}
	n.applyShifts()
	if key.Compare(n.key) < 0 {
		return n.left.search(key)
	} else if key.Compare(n.key) > 0 {
		return n.right.search(key)
	} else {
		return n
	}
}

func (n *node) overlap(i Interval, result []Interval) []Interval {
	if n == nil {
		return result
	}

	n.applyShifts()

	if i.Low() >= n.max {
		return result
	}

	result = n.left.overlap(i, result)

	if n.interval.Overlaps(i) {
		result = append(result, n.interval)
	}

	if i.High() <= n.interval.Low() {
		return result
	}

	result = n.right.overlap(i, result)
	return result
}

func (n *node) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *node) recalculateHeight() {
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
}

// Checks if node is balanced and rebalance
func (n *node) rebalanceTree() *node {
	if n == nil {
		return n
	}
	n.recalculateHeight()
	n.updateMax()

	// check balance factor and rotateLeft if right-heavy and rotateRight if left-heavy
	balanceFactor := n.left.getHeight() - n.right.getHeight()
	if balanceFactor <= -2 {
		// check if child is left-heavy and rotateRight first
		if n.right.left.getHeight() > n.right.right.getHeight() {
			n.right = n.right.rotateRight()
		}
		return n.rotateLeft()
	} else if balanceFactor >= 2 {
		// check if child is right-heavy and rotateLeft first
		if n.left.right.getHeight() > n.left.left.getHeight() {
			n.left = n.left.rotateLeft()
		}
		return n.rotateRight()
	}
	return n
}

// Rotate nodes left to balance node
func (n *node) rotateLeft() *node {
	newRoot := n.right
	n.right = newRoot.left
	newRoot.left = n

	n.recalculateHeight()
	n.updateMax()
	newRoot.recalculateHeight()
	newRoot.updateMax()
	return newRoot
}

// Rotate nodes right to balance node
func (n *node) rotateRight() *node {
	newRoot := n.left
	n.left = newRoot.right
	newRoot.right = n

	n.recalculateHeight()
	n.updateMax()
	newRoot.recalculateHeight()
	newRoot.updateMax()
	return newRoot
}

// Finds the smallest child (based on the key) for the current node
func (n *node) findSmallest() *node {
	if n.left != nil {
		return n.left.findSmallest()
	} else {
		return n
	}
}

func (n *node) applyShift(s *shift) {
	if n.tstamp >= s.tstamp {
		// this shift is outdated and we have already applied it
		return
	}

	n.tstamp = s.tstamp
	if n.max < s.idx {
		return
	}
	low, high := n.interval.Low(), n.interval.High()
	n.max += s.amt
	if n.key.Val > s.idx {
		n.key.Val += s.amt
	}
	if high >= s.idx {
		n.interval.ShiftHigh(s.amt)
	}
	if low >= s.idx {
		n.interval.ShiftLow(s.amt)
	}
}

func (n *node) applyShifts() {
	for i := range n.tree.shifts {
		n.applyShift(&n.tree.shifts[i])
	}
}

func (n *node) applyAllShifts() {
	if n == nil {
		return
	}

	n.applyShifts()
	n.left.applyAllShifts()
	n.right.applyAllShifts()
}

// Returns max number
func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
