// Package shifti provides an implementation of a shifting interval tree. The
// underlying interval tree used by this implementation comes from
// github.com/Workiva/go-datastructures. Changes include making ranges
// exclusive and adding support for amortized constant time shifts.
package shifti

import (
	"math"
)

type Tree struct {
	root                 *node
	maxDimension, number uint64
	dummy                node
	timestamp            uint64
	shifts               []shift
}

func NewTree(maxDimension uint64) *Tree {
	return &Tree{
		maxDimension: maxDimension,
		dummy:        newDummy(),
		shifts:       make([]shift, 0),
	}
}

type node struct {
	interval  Interval
	max, min  int64    // max/min value held by children
	children  [2]*node // array to hold left/right
	red       bool     // indicates if this node is red
	id        uint64
	timestamp uint64 // timestamp to determine which shifts to apply
	tree      *Tree  // reference to tree to find shift list
}

func (n *node) applyShift(s *shift) {
	// update timestamp
	n.timestamp = s.timestamp

	low, high := n.interval.Low(s.dim), n.interval.High(s.dim)
	if s.dim > 1 {
		if s.index >= high {
			return
		}
		n.interval.ShiftHigh(s.dim, s.count)
		if low >= s.index {
			n.interval.ShiftLow(s.dim, s.count)
		}

		if s.index <= low && -s.count >= high-s.index {
			// TODO: delete
			return
		}
		return
	}

	if n.max <= s.index {
		return
	}
	n.max += s.count
	if n.min >= s.index {
		n.min += s.count
	}

	if high > s.index {
		n.interval.ShiftHigh(s.dim, s.count)
	}
	if low >= s.index {
		n.interval.ShiftLow(s.dim, s.count)
	}
}

func (n *node) query(low, high int64, interval Interval, maxDimension uint64, fn func(node *node)) {
	tree := n.tree
	// apply shifts to this node and children
	for _, shift := range tree.shifts {
		if n.children[0] != nil && shift.timestamp > n.children[0].timestamp {
			n.children[0].applyShift(&shift)
		}
		if shift.timestamp > n.timestamp {
			n.applyShift(&shift)
		}
		if n.children[1] != nil && shift.timestamp > n.children[1].timestamp {
			n.children[1].applyShift(&shift)
		}
	}

	if n.children[0] != nil && overlaps(n.children[0].max, high, n.children[0].min, low) {
		n.children[0].query(low, high, interval, maxDimension, fn)
	}

	if intervalOverlaps(n, low, high, interval, maxDimension) {
		fn(n)
	}

	if n.children[1] != nil && overlaps(n.children[1].max, high, n.children[1].min, low) {
		n.children[1].query(low, high, interval, maxDimension, fn)
	}
}

func (n *node) adjustRanges() {
	for i := 0; i <= 1; i++ {
		if n.children[i] != nil {
			n.children[i].adjustRanges()
		}
	}

	n.adjustRange()
}

func (n *node) adjustRange() {
	setMin(n)
	setMax(n)
}

func newDummy() node {
	return node{
		children: [2]*node{},
	}
}

func newNode(interval Interval, min, max int64, dimension uint64) *node {
	itn := &node{
		interval: interval,
		min:      min,
		max:      max,
		red:      true,
		children: [2]*node{},
	}
	if interval != nil {
		itn.id = interval.Id()
	}

	return itn
}

type shift struct {
	dim       uint64
	index     int64
	count     int64
	timestamp uint64
}

func (t *Tree) Traverse(fn func(id Interval)) {
	nodes := []*node{t.root}

	for len(nodes) != 0 {
		c := nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
		if c != nil {
			fn(c.interval)
			if c.children[0] != nil {
				nodes = append(nodes, c.children[0])
			}
			if c.children[1] != nil {
				nodes = append(nodes, c.children[1])
			}
		}
	}
}

func (tree *Tree) resetDummy() {
	tree.dummy.children[0], tree.dummy.children[1] = nil, nil
	tree.dummy.red = false
}

// Len returns the number of items in this tree.
func (tree *Tree) Len() uint64 {
	return tree.number
}

// add will add the provided interval to the tree.
func (tree *Tree) add(iv Interval) {
	if tree.root == nil {
		tree.root = newNode(
			iv, iv.Low(1),
			iv.High(1),
			1,
		)
		tree.root.tree = tree
		tree.root.timestamp = tree.timestamp
		tree.root.red = false
		tree.number++
		return
	}

	tree.resetDummy()
	var (
		dummy               = tree.dummy
		parent, grandParent *node
		node                = tree.root
		dir, last           int
		otherLast           = 1
		id                  = iv.Id()
		max                 = iv.High(1)
		ivLow               = iv.Low(1)
		helper              = &dummy
	)

	// set this AFTER clearing dummy
	helper.children[1] = tree.root
	for {
		if node == nil {
			node = newNode(iv, ivLow, max, 1)
			node.tree = tree
			node.timestamp = tree.timestamp
			parent.children[dir] = node
			tree.number++
		} else if isRed(node.children[0]) && isRed(node.children[1]) {
			node.red = true
			node.children[0].red = false
			node.children[1].red = false
		}
		if max > node.max {
			node.max = max
		}

		if ivLow < node.min {
			node.min = ivLow
		}

		if isRed(parent) && isRed(node) {
			localDir := bool2int(helper.children[1] == grandParent)

			if node == parent.children[last] {
				helper.children[localDir] = rotate(grandParent, otherLast)
			} else {
				helper.children[localDir] = doubleRotate(grandParent, otherLast)
			}
		}

		if node.id == id {
			break
		}

		last = dir
		otherLast = takeOpposite(last)
		dir = compare(node.interval.Low(1), ivLow, node.id, id)

		if grandParent != nil {
			helper = grandParent
		}
		grandParent, parent, node = parent, node, node.children[dir]
	}

	tree.root = dummy.children[1]
	tree.root.red = false
}

// Add will add the provided intervals to this tree.
func (tree *Tree) Add(intervals ...Interval) {
	for _, iv := range intervals {
		tree.add(iv)
	}
}

// delete will remove the provided interval from the tree.
func (tree *Tree) delete(iv Interval) {
	if tree.root == nil {
		return
	}

	tree.resetDummy()
	var (
		dummy                      = tree.dummy
		found, parent, grandParent *node
		last, otherDir, otherLast  int // keeping track of last direction
		id                         = iv.Id()
		dir                        = 1
		node                       = &dummy
		ivLow                      = iv.Low(1)
	)

	node.children[1] = tree.root
	for node.children[dir] != nil {
		last = dir
		otherLast = takeOpposite(last)

		grandParent, parent, node = parent, node, node.children[dir]

		dir = compare(node.interval.Low(1), ivLow, node.id, id)
		otherDir = takeOpposite(dir)

		if node.id == id {
			found = node
		}

		if !isRed(node) && !isRed(node.children[dir]) {
			if isRed(node.children[otherDir]) {
				parent.children[last] = rotate(node, dir)
				parent = parent.children[last]
			} else if !isRed(node.children[otherDir]) {
				t := parent.children[otherLast]

				if t != nil {
					if !isRed(t.children[otherLast]) && !isRed(t.children[last]) {
						parent.red = false
						node.red = true
						t.red = true
					} else {
						localDir := bool2int(grandParent.children[1] == parent)

						if isRed(t.children[last]) {
							grandParent.children[localDir] = doubleRotate(
								parent, last,
							)
						} else if isRed(t.children[otherLast]) {
							grandParent.children[localDir] = rotate(
								parent, last,
							)
						}

						node.red = true
						grandParent.children[localDir].red = true
						grandParent.children[localDir].children[0].red = false
						grandParent.children[localDir].children[1].red = false
					}
				}
			}
		}
	}

	if found != nil {
		tree.number--
		found.interval, found.max, found.min, found.id = node.interval, node.max, node.min, node.id
		parentDir := bool2int(parent.children[1] == node)
		childDir := bool2int(node.children[0] == nil)

		parent.children[parentDir] = node.children[childDir]
	}

	tree.root = dummy.children[1]
	if tree.root != nil {
		tree.root.red = false
	}
}

// Delete will remove the provided intervals from this tree.
func (tree *Tree) Delete(intervals ...Interval) {
	for _, iv := range intervals {
		tree.delete(iv)
	}
	if tree.root != nil {
		tree.root.adjustRanges()
	}
}

// Query will return a list of intervals that intersect the provided
// interval.  The provided interval's ID method is ignored so the
// provided ID is irrelevant.
func (tree *Tree) Query(interval Interval) Intervals {
	if tree.root == nil {
		return nil
	}

	var (
		Intervals = iPool.Get().(Intervals)
		ivLow     = interval.Low(1)
		ivHigh    = interval.High(1)
	)

	tree.root.query(ivLow, ivHigh, interval, tree.maxDimension, func(node *node) {
		Intervals = append(Intervals, node.interval)
	})

	return Intervals
}

const shiftThreshold = 1024

func (tree *Tree) Shift(dim uint64, index, count int64) {
	tree.timestamp++
	tree.shifts = append(tree.shifts, shift{
		dim:       dim,
		index:     index,
		count:     count,
		timestamp: tree.timestamp,
	})

	if len(tree.shifts) >= shiftThreshold {
		tree.applyAllShifts()
	}
}

func (tree *Tree) applyAllShifts() {
	tree.root.query(math.MinInt64, math.MaxInt64, nil, tree.maxDimension, func(n *node) {
		for _, shift := range tree.shifts {
			if shift.timestamp > n.timestamp {
				n.applyShift(&shift)
			}
		}
	})
	tree.shifts = make([]shift, 0)
}

func isRed(node *node) bool {
	return node != nil && node.red
}

func setMax(parent *node) {
	parent.max = parent.interval.High(1)

	if parent.children[0] != nil && parent.children[0].max > parent.max {
		parent.max = parent.children[0].max
	}

	if parent.children[1] != nil && parent.children[1].max > parent.max {
		parent.max = parent.children[1].max
	}
}

func setMin(parent *node) {
	parent.min = parent.interval.Low(1)
	if parent.children[0] != nil && parent.children[0].min < parent.min {
		parent.min = parent.children[0].min
	}

	if parent.children[1] != nil && parent.children[1].min < parent.min {
		parent.min = parent.children[1].min
	}

	if parent.interval.Low(1) < parent.min {
		parent.min = parent.interval.Low(1)
	}
}

func rotate(parent *node, dir int) *node {
	otherDir := takeOpposite(dir)

	child := parent.children[otherDir]
	parent.children[otherDir] = child.children[dir]
	child.children[dir] = parent
	parent.red = true
	child.red = false
	child.max = parent.max
	setMax(child)
	setMax(parent)
	setMin(child)
	setMin(parent)

	return child
}

func doubleRotate(parent *node, dir int) *node {
	otherDir := takeOpposite(dir)

	parent.children[otherDir] = rotate(parent.children[otherDir], otherDir)
	return rotate(parent, dir)
}

func intervalOverlaps(n *node, low, high int64, interval Interval, maxDimension uint64) bool {
	if !overlaps(n.interval.High(1), high, n.interval.Low(1), low) {
		return false
	}

	if interval == nil {
		return true
	}

	for i := uint64(2); i <= maxDimension; i++ {
		if !n.interval.Overlaps(interval, i) {
			return false
		}
	}

	return true
}

func overlaps(high, otherHigh, low, otherLow int64) bool {
	return high > otherLow && low < otherHigh
}

// compare returns an int indicating which direction the node
// should go.
func compare(nodeLow, ivLow int64, nodeID, ivID uint64) int {
	if ivLow > nodeLow {
		return 1
	}

	if ivLow < nodeLow {
		return 0
	}

	return bool2int(ivID > nodeID)
}

func bool2int(value bool) int {
	if value {
		return 1
	}

	return 0
}

func takeOpposite(value int) int {
	return 1 - value
}
