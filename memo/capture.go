package memo

import (
	"bytes"
	"fmt"
)

const Dummy = -1

// A Capture represents an AST capture node. It stores the ID of the capture, the
// start position and length, and any children.
type Capture struct {
	Id       int
	Children []*Capture

	// The memoization entry this capture is stored in. This may be nil if the
	// capture is standalone. If the memoization is non-nil it will be used for
	// calculating the start position of this capture. Since memoization
	// entries are relocatable when an edit occurs, this makes the capture also
	// relocatable.
	ment *Entry
	// the start position is an offset from the start of this capture's
	// memoization entry if one exists. If not, the start position is an
	// absolute offset.
	start  int
	length int
}

// NewCapture constructs a new AST node.
func NewCapture(id int, start int, length int, children []*Capture) *Capture {
	return &Capture{
		Id:       id,
		Children: children,
		start:    start,
		length:   length,
	}
}

// setMemo sets this node's locator so that the start position of this
// capture will be relative to the locator.
func (c *Capture) setMemo(e *Entry) {
	c.ment = e
	c.start = c.start - e.Start()

	for _, c := range c.Children {
		if !c.memoized() {
			c.setMemo(e)
		}
	}
}

// memoized returns if this capture is stored within a memoization entry.
func (c *Capture) memoized() bool {
	return c.ment != nil
}

// Start returns the start index of this AST capture.
func (c *Capture) Start() int {
	if c.ment != nil {
		return c.ment.Start() + c.start
	}
	return c.start
}

// End returns the end index of this AST capture.
func (c *Capture) End() int {
	return c.Start() + c.length
}

// String returns a readable string representation of this node, showing the ID
// of this node and its children.
func (c *Capture) String() string {
	buf := &bytes.Buffer{}
	for i, c := range c.Children {
		buf.WriteString(c.String())
		if i != len(c.Children)-1 {
			buf.WriteString(", ")
		}
	}
	return fmt.Sprintf("{%d, [%s]}", c.Id, buf.String())
}
