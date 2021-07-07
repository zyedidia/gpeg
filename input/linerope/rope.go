package linerope

import (
	"io"
	"runtime"
	"sync"
)

var DefaultOptions = Options{
	SplitLen:       4096,
	JoinLen:        2048,
	RebalanceRatio: 1.2,
	LineSep:        []byte{'\n'},
}

type Options struct {
	// SplitLen is the threshold above which slices will be split into separate
	// nodes.
	SplitLen int
	// JoinLen is the threshold below which nodes will be merged into slices.
	JoinLen int
	// RebalanceRatio is the threshold used to trigger a rebuild during a
	// rebalance operation.
	RebalanceRatio float64
	// LineSep is the newline byte sequence (usually '\n' or '\r\n').
	LineSep []byte
}

type nodeType byte

const (
	tLeaf nodeType = iota
	tNode
)

// A Node in the rope structure. If the kind is tLeaf, only the value and
// length are valid, and if the kind is tNode, only length, left, right are
// valid.
type Node struct {
	kind        nodeType
	value       []byte
	length      int
	llength     loc
	left, right *Node
	opts        *Options
}

// New returns a new rope node from the given byte slice. The underlying
// data is not copied so the user should ensure that it is okay to insert and
// delete from the input slice.
func New(b []byte, opts *Options) *Node {
	// We build the tree from the bottom up for extra efficiency. This avoids
	// counting duplicate newlines a logarithmic number of times (for each
	// level of the tree).
	//
	// We make the chunk size equal to SplitLength which means a node will be
	// split when the first edit is made. Since most nodes will never be
	// edited, it makes sense to fill them all up to avoid wasting space, even
	// if it means inserting will require a split the first time a node is
	// edited.
	chunksz := opts.SplitLen
	nchunks := len(b) / chunksz
	nodes := make([]*Node, nchunks, nchunks+1)

	// For even better performance, we load the chunks in parallel. Chunk
	// loading is distributed among the cores available on the machine.
	var nthreads = runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(nthreads)
	for t := 0; t < nthreads; t++ {
		go func(t int) {
			start := t * (nchunks / nthreads)
			end := t*(nchunks/nthreads) + (nchunks / nthreads)
			if t == nthreads-1 {
				end = nchunks
			}
			for i := start; i < end; i++ {
				j := i * chunksz
				// triple index slice notation allows a sort of copy-on-write behavior
				// which is extremely beneficial to us because it's likely that this
				// slice is backed by a memory-mapped file.
				slc := b[j : j+chunksz : j+chunksz]
				nodes[i] = &Node{
					kind:    tLeaf,
					value:   slc,
					length:  len(slc),
					llength: llen(slc, opts.LineSep),
					opts:    opts,
				}
			}
			wg.Done()
		}(t)
	}
	wg.Wait()
	// load any extra bytes
	slc := b[nchunks*chunksz : len(b) : len(b)]
	nodes = append(nodes, &Node{
		kind:    tLeaf,
		value:   slc,
		length:  len(slc),
		llength: llen(slc, opts.LineSep),
		opts:    opts,
	})
	return buildTree(nodes)
}

// recursively creates parent nodes
func buildTree(nodes []*Node) *Node {
	if len(nodes) == 1 {
		return nodes[0]
	}
	if len(nodes)%2 != 0 {
		l := len(nodes)
		nodes[l-2] = join(nodes[l-2], nodes[l-1])
		nodes = nodes[:l-1]
	}

	newnodes := make([]*Node, 0, len(nodes)/2+1)
	for i := 0; i < len(nodes); i += 2 {
		newnodes = append(newnodes, join(nodes[i], nodes[i+1]))
	}
	return buildTree(newnodes)
}

// Len returns the number of elements stored in the rope.
func (n *Node) Len() int {
	return n.length
}

// LLen returns the line/col location one byte beyond the last position in the
// file.
func (n *Node) LLen() (lines, cols int) {
	return n.llength.line, n.llength.col
}

func (n *Node) NumLines() int {
	return n.llength.line
}

func (n *Node) adjust() {
	switch n.kind {
	case tLeaf:
		if n.length > n.opts.SplitLen {
			divide := n.length / 2
			n.left = New(n.value[:divide], n.opts)
			n.right = New(n.value[divide:], n.opts)
			n.value = nil
			n.kind = tNode
			n.length = n.left.length + n.right.length
			n.llength = addlocs(n.left.llength, n.right.llength)
		}
	default: // case tNode
		if n.length < n.opts.JoinLen {
			n.value = n.Value()
			n.left = nil
			n.right = nil
			n.kind = tLeaf
			n.length = len(n.value)
			n.llength = llen(n.value, n.opts.LineSep)
		}
	}
}

// Value returns the elements of this node concatenated into a slice. May
// return the underyling slice without copying, so do not modify the returned
// slice.
func (n *Node) Value() []byte {
	switch n.kind {
	case tLeaf:
		return n.value
	default: // case tNode
		return concat(n.left.Value(), n.right.Value())
	}
}

// Remove deletes the range [start:end) (exclusive bound) from the rope.
func (n *Node) Remove(start, end int) {
	switch n.kind {
	case tLeaf:
		// slice tricks delete
		n.value = remove(n.value, start, end)
		n.length = len(n.value)
		n.llength = llen(n.value, n.opts.LineSep)
	default: // case tNode
		leftLength := n.left.length
		leftStart := min(start, leftLength)
		leftEnd := min(end, leftLength)
		rightLength := n.right.length
		rightStart := max(0, min(start-leftLength, rightLength))
		rightEnd := max(0, min(end-leftLength, rightLength))
		if leftStart < leftLength {
			n.left.Remove(leftStart, leftEnd)
		}
		if rightEnd > 0 {
			n.right.Remove(rightStart, rightEnd)
		}
		n.length = n.left.length + n.right.length
		n.llength = addlocs(n.left.llength, n.right.llength)
	}
	n.adjust()
}

// Insert inserts the given value at pos.
func (n *Node) Insert(pos int, value []byte) {
	switch n.kind {
	case tLeaf:
		// slice tricks insert
		n.value = insert(n.value, pos, value)
		n.length = len(n.value)
		n.llength = llen(n.value, n.opts.LineSep)
	default: // case tNode
		leftLength := n.left.length
		if pos < leftLength {
			n.left.Insert(pos, value)
		} else {
			n.right.Insert(pos-leftLength, value)
		}
		n.length = n.left.length + n.right.length
		n.llength = addlocs(n.left.llength, n.right.llength)
	}
	n.adjust()
}

// slice returns the range of the rope from [start:end).
func (n *Node) slice(start, end int) []byte {
	if start >= end {
		return []byte{}
	}

	switch n.kind {
	case tLeaf:
		return n.value[start:end]
	default: // case tNode
		leftLength := n.left.length
		leftStart := min(start, leftLength)
		leftEnd := min(end, leftLength)
		rightLength := n.right.length
		rightStart := max(0, min(start-leftLength, rightLength))
		rightEnd := max(0, min(end-leftLength, rightLength))

		if leftStart != leftEnd {
			if rightStart != rightEnd {
				return concat(n.left.slice(leftStart, leftEnd), n.right.slice(rightStart, rightEnd))
			} else {
				return n.left.slice(leftStart, leftEnd)
			}
		} else {
			if rightStart != rightEnd {
				return n.right.slice(rightStart, rightEnd)
			} else {
				return []byte{}
			}
		}
	}
}

// OffsetAt returns the absolute character offset of a line/col position.
func (n *Node) OffsetAt(line, col int) int {
	pos := loc{line, col}
	switch n.kind {
	case tLeaf:
		return indexN(n.value, n.opts.LineSep, line) + len(n.opts.LineSep) + col
	default: // case tNode
		leftLength := n.left.llength
		if pos.cmp(leftLength) < 0 {
			return n.left.OffsetAt(line, col)
		} else {
			l := sublocs(pos, leftLength)
			return n.left.length + n.right.OffsetAt(l.line, l.col)
		}
	}
}

// LineColAt returns the line/col position of an absolute character offset.
func (n *Node) LineColAt(pos int) (line, col int) {
	l := n.lineColAt(pos)
	return l.line, l.col
}

func (n *Node) lineColAt(pos int) loc {
	switch n.kind {
	case tLeaf:
		return lineCol(n.value, n.opts.LineSep, pos)
	default: // case tNode
		leftLength := n.left.length
		if pos < leftLength {
			return n.left.lineColAt(pos)
		} else {
			return addlocs(n.left.llength, n.right.lineColAt(pos-leftLength))
		}
	}
}

// SliceLC is the same as Slice but uses line/col positions for start and end.
func (n *Node) SliceLC(startl, startc, endl, endc int) []byte {
	return n.sliceLC(loc{startl, startc}, loc{endl, endc})
}

func (n *Node) sliceLC(start, end loc) []byte {
	if start.cmp(end) >= 0 {
		return []byte{}
	}

	switch n.kind {
	case tLeaf:
		return sliceloc(n.value, n.opts.LineSep, start, end)
	default: // case tNode
		leftLength := n.left.llength
		leftStart := minloc(start, leftLength)
		leftEnd := minloc(end, leftLength)
		rightLength := n.right.llength
		rightStart := maxloc(lzero, minloc(sublocs(start, leftLength), rightLength))
		rightEnd := maxloc(lzero, minloc(sublocs(end, leftLength), rightLength))

		if leftStart != leftEnd {
			if rightStart != rightEnd {
				return concat(n.left.sliceLC(leftStart, leftEnd), n.right.sliceLC(rightStart, rightEnd))
			} else {
				return n.left.sliceLC(leftStart, leftEnd)
			}
		} else {
			if rightStart != rightEnd {
				return n.right.sliceLC(rightStart, rightEnd)
			} else {
				return []byte{}
			}
		}
	}
}

// At returns the element at the given position.
func (n *Node) At(pos int) byte {
	s := n.slice(pos, pos+1)
	return s[0]
}

// SplitAt splits the node at the given index and returns two new ropes
// corresponding to the left and right portions of the split.
func (n *Node) SplitAt(i int) (*Node, *Node) {
	switch n.kind {
	case tLeaf:
		return New(n.value[:i], n.opts), New(n.value[i:], n.opts)
	default: // case tNode
		m := n.left.length
		if i == m {
			return n.left, n.right
		} else if i < m {
			l, r := n.left.SplitAt(i)
			return l, join(r, n.right)
		}
		l, r := n.right.SplitAt(i - m)
		return join(n.left, l), r
	}
}

func join(l, r *Node) *Node {
	n := &Node{
		left:    l,
		right:   r,
		length:  l.length + r.length,
		llength: addlocs(l.llength, r.llength),
		kind:    tNode,
		opts:    l.opts,
	}
	n.adjust()
	return n
}

// Join merges all the given ropes together into one rope.
func Join(a, b *Node, more ...*Node) *Node {
	s := join(a, b)
	for _, n := range more {
		s = join(s, n)
	}
	return s
}

// Rebuild rebuilds the entire rope structure, resulting in a balanced tree.
func (n *Node) Rebuild() {
	switch n.kind {
	case tNode:
		n.value = concat(n.left.Value(), n.right.Value())
		n.left = nil
		n.right = nil
		n.adjust()
	}
}

// Rebalance finds unbalanced nodes and rebuilds them.
func (n *Node) Rebalance() {
	switch n.kind {
	case tNode:
		lratio := float64(n.left.length) / float64(n.right.length)
		rratio := float64(n.right.length) / float64(n.left.length)
		if lratio > n.opts.RebalanceRatio || rratio > n.opts.RebalanceRatio {
			n.Rebuild()
		} else {
			n.left.Rebalance()
			n.right.Rebalance()
		}
	}
}

// Each applies the given function to every node in the rope.
func (n *Node) Each(fn func(n *Node)) {
	fn(n)
	if n.kind == tNode {
		n.left.Each(fn)
		n.right.Each(fn)
	}
}

// EachLeaf applies the given function to every leaf node in order.
func (n *Node) EachLeaf(fn func(n *Node) bool) bool {
	switch n.kind {
	case tLeaf:
		return fn(n)
	default: // case tNode
		if n.left.EachLeaf(fn) {
			return true
		}
		return n.right.EachLeaf(fn)
	}
}

// ReadAt implements the io.ReaderAt interface.
func (n *Node) ReadAt(p []byte, off int64) (nread int, err error) {
	if off > int64(n.length) {
		return 0, io.EOF
	}

	end := off + int64(len(p))
	if end >= int64(n.length) {
		end = int64(n.length)
		err = io.EOF
	}
	b := n.slice(int(off), int(end))
	nread = copy(p, b)
	return nread, err
}

// WriteTo implements the io.WriterTo interface.
func (n *Node) WriteTo(w io.Writer) (int64, error) {
	var err error
	var ntotal int64
	n.EachLeaf(func(it *Node) bool {
		var nwritten int
		nwritten, err = w.Write(it.Value())
		ntotal += int64(nwritten)
		return err != nil
	})
	return ntotal, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// from slice tricks
func insert(s []byte, k int, vs []byte) []byte {
	if n := len(s) + len(vs); n <= cap(s) {
		s2 := s[:n]
		copy(s2[k+len(vs):], s[k:])
		copy(s2[k:], vs)
		return s2
	}
	s2 := make([]byte, len(s)+len(vs))
	copy(s2, s[:k])
	copy(s2[k:], vs)
	copy(s2[k+len(vs):], s[k:])
	return s2
}

func concat(a, b []byte) []byte {
	c := make([]byte, 0, len(a)+len(b))
	c = append(c, a...)
	c = append(c, b...)
	return c
}

func remove(s []byte, start, end int) []byte {
	if len(s) == cap(s) {
		// "copy-on-write" for slices where len == cap.
		ns := make([]byte, len(s)-(end-start), cap(s))
		copy(ns, s[:start])
		copy(ns[start:], s[end:])
		return ns
	}
	return append(s[:start], s[end:]...)
}
