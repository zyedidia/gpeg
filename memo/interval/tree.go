// Package interval provides an interval tree backed by an AVL tree. In addition,
// the interval tree supports a lazy shifting algorithm.
package interval

type key struct {
	id  int
	pos int
}

// compare orders keys by pos and then id.
func (k key) compare(other key) int {
	if k.pos < other.pos {
		return -1
	} else if k.pos > other.pos {
		return 1
	} else if k.id < other.id {
		return -1
	} else if k.id > other.id {
		return 1
	}
	return 0
}

type Tree struct {
	root *node
}

// Adds the given interval to the tree. An id can also be given to the interval
// to separate different types of intervals.
func (t *Tree) Add(id, low, high int, val Value) (pos Pos) {
	t.root, pos = t.root.add(key{id, low}, high, val, nil)
	return pos
}

// FindLargest returns the largest interval associated with (id, pos).
func (t *Tree) FindLargest(id, pos int) Value {
	n := t.root.search(key{id, pos})
	if n == nil || len(n.iv.ivs) == 0 {
		return nil
	}

	var max, maxi int
	for i := range n.iv.ivs {
		if n.iv.ivs[i].interval.High > max {
			max = n.iv.ivs[i].interval.High
			maxi = i
		}
	}
	return n.iv.ivs[maxi].value
}

// RemoveAndShift removes all entries that overlap with [low, high) and then shifts
// all entries greater than low by amt.
func (t *Tree) RemoveAndShift(low, high, amt int) {
	t.root = t.root.removeOverlaps(low, high, t.root)
	if amt != 0 {
		t.root.addShift(shift{low, amt})
	}
}

// Size returns the number of intervals in the tree.
func (t *Tree) Size() int {
	return t.root.getSize()
}

type ivalues struct {
	ivs  []ivalue
	node *node
}

func (iv *ivalues) Pos() int {
	iv.node.applyAllShifts()
	return iv.node.key.pos
}

type ivalue struct {
	interval Interval
	value    Value
}

// A shift of intervals in the tree. The shift starts at idx and moves
// intervals after idx by amt. Shifts are lazily applied in the tree to avoid
// frequent linear time costs.
type shift struct {
	idx int
	amt int
}

type node struct {
	key    key
	max    int
	iv     *ivalues
	shifts []shift

	// height counts nodes (not edges)
	height int
	left   *node
	right  *node
	parent *node
}

func (n *node) addShift(sh shift) {
	if n == nil {
		return
	}

	n.shifts = append(n.shifts, sh)
}

func (n *node) applyShifts() {
	if n == nil {
		return
	}
	for _, sh := range n.shifts {
		if n.max >= sh.idx {
			if n.key.pos >= sh.idx {
				n.key.pos += sh.amt
				for i, iv := range n.iv.ivs {
					n.iv.ivs[i].interval = iv.interval.Shift(sh.amt)
				}
			}
			n.max += sh.amt
			// n.updateMax()
		}

		n.left.addShift(sh)
		n.right.addShift(sh)
	}
	n.shifts = nil
}

func (n *node) add(key key, high int, value Value, parent *node) (*node, *ivalues) {
	if n == nil {
		n = new(node)
		*n = node{
			key: key,
			max: high,
			iv: &ivalues{
				ivs: []ivalue{ivalue{
					interval: Interval{key.pos, high},
					value:    value,
				}},
				node: n,
			},
			height: 1,
			left:   nil,
			right:  nil,
			parent: parent,
		}
		return n, n.iv
	}
	n.applyShifts()

	var iv *ivalues
	if key.compare(n.key) < 0 {
		n.left, iv = n.left.add(key, high, value, n)
	} else if key.compare(n.key) > 0 {
		n.right, iv = n.right.add(key, high, value, n)
	} else {
		// if same key exists update value
		n.iv.ivs = append(n.iv.ivs, ivalue{
			interval: Interval{key.pos, high},
			value:    value,
		})
		iv = n.iv
	}
	return n.rebalanceTree(parent), iv
}

func (n *node) calcMax() int {
	max := 0
	for _, iv := range n.iv.ivs {
		if iv.interval.High > max {
			max = iv.interval.High
		}
	}
	return max
}

func (n *node) updateMax() {
	if n != nil {
		if n.right != nil {
			n.max = max(n.max, n.right.max)
		}
		if n.left != nil {
			n.max = max(n.max, n.left.max)
		}
		n.max = max(n.max, n.calcMax())
	}
}

func (n *node) remove(key key, parent *node) *node {
	if n == nil {
		return nil
	}
	n.applyShifts()
	if key.compare(n.key) < 0 {
		n.left = n.left.remove(key, n)
	} else if key.compare(n.key) > 0 {
		n.right = n.right.remove(key, n)
	} else {
		if n.left != nil && n.right != nil {
			n.left.applyShifts()
			n.right.applyShifts()
			// node to delete found with both children;
			// replace values with smallest node of the right sub-tree
			rightMinNode := n.right.findSmallest()
			n.key = rightMinNode.key
			// n.iv.node = nil
			n.iv = rightMinNode.iv
			n.iv.node = n
			n.shifts = rightMinNode.shifts
			// delete smallest node that we replaced
			n.right = n.right.remove(rightMinNode.key, n)
		} else if n.left != nil {
			n.left.applyShifts()
			// node only has left child
			n = n.left
		} else if n.right != nil {
			n.right.applyShifts()
			// node only has right child
			n = n.right
		} else {
			// node has no children
			n = nil
			return n
		}

	}
	n.parent = parent
	return n.rebalanceTree(parent)
}

func (n *node) search(key key) *node {
	if n == nil {
		return nil
	}
	n.applyShifts()
	if key.compare(n.key) < 0 {
		return n.left.search(key)
	} else if key.compare(n.key) > 0 {
		return n.right.search(key)
	} else {
		return n
	}
}

func (n *node) overlaps(low, high int, result []Value) []Value {
	if n == nil {
		return result
	}

	n.applyShifts()

	if low >= n.max {
		return result
	}

	result = n.left.overlaps(low, high, result)

	for _, iv := range n.iv.ivs {
		if Overlaps(iv.interval, Interval{low, high}) {
			result = append(result, iv.value)
		}
	}

	if high <= n.key.pos {
		return result
	}

	result = n.right.overlaps(low, high, result)
	return result
}

func (n *node) removeOverlaps(low, high int, parent *node) *node {
	if n == nil {
		return n
	}

	n.applyShifts()

	if low >= n.max {
		return n
	}

	n.left = n.left.removeOverlaps(low, high, n)

	for i := 0; i < len(n.iv.ivs); {
		if Overlaps(n.iv.ivs[i].interval, Interval{low, high}) {
			n.iv.ivs[i] = n.iv.ivs[len(n.iv.ivs)-1]
			n.iv.ivs = n.iv.ivs[:len(n.iv.ivs)-1]
		} else {
			i++
		}
	}

	if len(n.iv.ivs) == 0 {
		doright := high > n.key.pos
		n = n.remove(n.key, parent)
		if doright {
			return n.removeOverlaps(low, high, parent)
		}
		return n
	}

	if high <= n.key.pos {
		return n
	}

	n.right = n.right.removeOverlaps(low, high, n)
	return n
}

func (n *node) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *node) getSize() int {
	if n == nil {
		return 0
	}
	return n.left.getSize() + n.right.getSize() + 1
}

func (n *node) updateHeightAndMax() {
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
	n.updateMax()
}

// Checks if node is balanced and rebalance
func (n *node) rebalanceTree(parent *node) *node {
	if n == nil {
		return n
	}
	n.updateHeightAndMax()

	// check balance factor and rotateLeft if right-heavy and rotateRight if left-heavy
	balanceFactor := n.left.getHeight() - n.right.getHeight()
	if balanceFactor == -2 {
		// check if child is left-heavy and rotateRight first
		if n.right.left.getHeight() > n.right.right.getHeight() {
			n.right = n.right.rotateRight(n)
		}
		return n.rotateLeft(parent)
	} else if balanceFactor == 2 {
		// check if child is right-heavy and rotateLeft first
		if n.left.right.getHeight() > n.left.left.getHeight() {
			n.left = n.left.rotateLeft(n)
		}
		return n.rotateRight(parent)
	}
	return n
}

// Rotate nodes left to balance node
func (n *node) rotateLeft(newParent *node) *node {
	n.applyShifts()
	if n.right != nil {
		n.right.applyShifts()
	}

	newRoot := n.right
	n.right = newRoot.left
	newRoot.left = n

	n.parent = newRoot

	if n.right != nil {
		n.right.parent = n
	}

	newRoot.parent = newParent

	n.updateHeightAndMax()
	newRoot.updateHeightAndMax()
	return newRoot
}

// Rotate nodes right to balance node
func (n *node) rotateRight(newParent *node) *node {
	n.applyShifts()
	if n.left != nil {
		n.left.applyShifts()
	}
	newRoot := n.left
	n.left = newRoot.right
	newRoot.right = n

	n.parent = newRoot

	if n.left != nil {
		n.left.parent = n
	}

	newRoot.parent = newParent

	n.updateHeightAndMax()
	newRoot.updateHeightAndMax()
	return newRoot
}

// Finds the smallest child (based on the key) for the current node
func (n *node) findSmallest() *node {
	if n.left != nil {
		n.left.applyShifts()
		return n.left.findSmallest()
	} else {
		return n
	}
}

func (n *node) applyAllShifts() {
	if n.parent != nil && n.parent != n {
		n.parent.applyAllShifts()
	}
	n.applyShifts()
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
