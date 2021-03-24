// Package avlint provides an interval tree backed by an AVL tree. In addition,
// the interval tree supports shifting intervals in amortized constant time
// using lazy shifts.
package avlint

// ShiftThreshold is the number of shifts to accumulate before applying all
// shifts.
const ShiftThreshold = 1024

// A key stores the start position of an interval, and a unique ID if you would
// like to store multiple intervals starting from the same position. The key is
// used for uniquely identifying a particular interval when searching or
// removing from the tree.
type key struct {
	pos int
	id  int
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

// A shift of intervals in the tree. The shift starts at idx and moves
// intervals after idx by amt. Shifts are lazily applied in the tree to avoid
// linear time costs.
type shift struct {
	idx    int
	amt    int
	tstamp uint64
}

type Tree struct {
	root   *node
	shifts []shift // list of non-applied shifts
	tstamp uint64  // most recent timestamp
}

// Adds the given interval to the tree. An id should also be given to the
// interval to uniquely identify it if any other intervals begin at the same
// location.
func (t *Tree) Add(low, high int, value interface{}, id int) Locator {
	var loc Locator
	t.root, loc = t.root.add(t, key{
		pos: low,
		id:  id,
	}, interval{
		low:   low,
		high:  high,
		value: value,
	})
	return loc
}

// Remove the interval starting at pos with the given id.
func (t *Tree) Remove(pos, id int) {
	t.root = t.root.remove(key{
		pos: pos,
		id:  id,
	})
}

// Search for the interval starting at pos with the given id. Returns nil if no
// such interval exists.
func (t *Tree) Search(pos, id int) interface{} {
	n := t.root.search(key{
		pos: pos,
		id:  id,
	})
	if n != nil {
		max := 0
		for i, in := range n.interval.ins[1:] {
			if in.length() > n.interval.ins[max].length() {
				max = i + 1
			}
		}

		return n.interval.ins[max].value
	}
	return nil
}

// Overlap returns a list of intervals stored in the tree that overlap with the
// given interval. This function executes in time O(lg n + k) where n is the
// number of intervals stored in the tree, and k is the number of overlapping
// intervals (size of the return value).
func (t *Tree) Overlap(low, high int) []interface{} {
	result := make([]interface{}, 0)
	return t.root.overlap(low, high, result)
}

// Shift all intervals in the tree after idx by amt. The shift idx should not
// lie inside an interval. This could conceivably be implemented, but is not
// currently. If a negative shift is performed, ensure that there is space for
// all intervals to be shifted left without overlapping with another interval.
func (t *Tree) Shift(idx, amt int) {
	if amt == 0 {
		return
	}

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
	t.shifts = nil
}

// Size returns the total number of intervals stored in the tree.
func (t *Tree) Size() int {
	return t.root.size()
}

type node struct {
	key      key
	max      int
	interval *lazyInterval
	tstamp   uint64 // timestamp to determine which shifts to apply
	tree     *Tree

	// height counts nodes (not edges)
	height int
	left   *node
	right  *node
}

// Adds a new node
func (n *node) add(tree *Tree, key key, value interval) (*node, Locator) {
	if n == nil {
		nn := &node{
			tree:   tree,
			key:    key,
			max:    value.High(),
			height: 1,
			left:   nil,
			right:  nil,
			tstamp: tree.tstamp,
		}
		nn.interval = &lazyInterval{
			ins: []interval{value},
			n:   nn,
		}
		return nn, nn.interval
	}
	n.applyShifts()

	var loc Locator
	if key.compare(n.key) < 0 {
		n.left, loc = n.left.add(tree, key, value)
	} else if key.compare(n.key) > 0 {
		n.right, loc = n.right.add(tree, key, value)
	} else {
		// if same key exists update value
		n.interval.ins = append(n.interval.ins, value)
		n.tstamp = tree.tstamp
		loc = n.interval
	}
	return n.rebalanceTree(), loc
}

func (n *node) updateMax() {
	if n != nil {
		if n.right != nil {
			n.max = max(n.max, n.right.max)
		}
		if n.left != nil {
			n.max = max(n.max, n.left.max)
		}
		n.max = max(n.max, n.interval.High())
	}
}

// Removes a node
func (n *node) remove(key key) *node {
	if n == nil {
		return nil
	}
	n.applyShifts()
	if key.compare(n.key) < 0 {
		n.left = n.left.remove(key)
	} else if key.compare(n.key) > 0 {
		n.right = n.right.remove(key)
	} else {
		if n.left != nil && n.right != nil {
			n.left.applyShifts()
			n.right.applyShifts()
			// node to delete found with both children;
			// replace values with smallest node of the right sub-tree
			rightMinNode := n.right.findSmallest()
			n.key = rightMinNode.key
			n.interval = rightMinNode.interval
			n.interval.n = rightMinNode
			n.tstamp = rightMinNode.tstamp
			// delete smallest node that we replaced
			n.right = n.right.remove(rightMinNode.key)
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
	return n.rebalanceTree()
}

// Searches for a node
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

func (n *node) overlap(low, high int, result []interface{}) []interface{} {
	if n == nil {
		return result
	}

	n.applyShifts()

	if low >= n.max {
		return result
	}

	result = n.left.overlap(low, high, result)

	for _, in := range n.interval.ins {
		if overlaps(in, low, high) {
			result = append(result, in.value)
		}
	}

	if high <= n.key.pos {
		return result
	}

	result = n.right.overlap(low, high, result)
	return result
}

func (n *node) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *node) size() int {
	if n == nil {
		return 0
	}
	return n.left.size() + n.right.size() + 1
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
	n.applyShifts()
	if n.right != nil {
		n.right.applyShifts()
	}

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
	n.applyShifts()
	if n.left != nil {
		n.left.applyShifts()
	}

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
		n.left.applyShifts()
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
	n.max += s.amt
	if n.key.pos >= s.idx {
		n.key.pos += s.amt
		n.interval.Shift(s.amt)
	}
	n.updateMax()
}

func (n *node) applyShifts() {
	// optimization: first check if we are completely up-to-date and if so
	// there is nothing to do.
	if len(n.tree.shifts) == 0 || n.tstamp >= n.tree.shifts[len(n.tree.shifts)-1].tstamp {
		return
	}
	// optimization: search backwards to find the starting point. Alternatively
	// we could binary search? not sure which is faster.
	var j int
	for j = len(n.tree.shifts) - 1; j > 0; j-- {
		if n.tstamp >= n.tree.shifts[j].tstamp {
			j = j + 1
			break
		}
	}
	for i := range n.tree.shifts[j:] {
		n.applyShift(&n.tree.shifts[j+i])
	}
}

func (n *node) applyAllShifts() {
	if n == nil {
		return
	}

	n.left.applyAllShifts()
	n.right.applyAllShifts()
	n.applyShifts()
}

func (n *node) eachNode(fn func(*node)) {
	if n == nil {
		return
	}

	n.left.eachNode(fn)
	n.applyShifts()
	fn(n)
	n.right.eachNode(fn)
}

type lazyInterval struct {
	ins []interval
	n   *node
}

func (i *lazyInterval) Start() int {
	i.n.applyShifts()
	return i.n.key.pos
}

func (i *lazyInterval) High() int {
	high := 0
	for _, in := range i.ins {
		if in.High() > high {
			high = in.High()
		}
	}
	return high
}

func (i *lazyInterval) Shift(amt int) {
	for j := range i.ins {
		i.ins[j].low += amt
		i.ins[j].high += amt
	}
}

// Returns max number
func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
