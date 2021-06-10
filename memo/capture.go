package memo

import (
	"bytes"
	"fmt"
)

const (
	tNode = iota
	tDummy
)

type Capture struct {
	id  int32
	typ int32

	off      int
	length   int
	ment     *Entry
	children []*Capture
}

func NewCaptureNode(id int, start, length int, children []*Capture) *Capture {
	c := &Capture{
		id:       int32(id),
		typ:      tNode,
		off:      start,
		length:   length,
		children: children,
	}
	return c
}

func NewCaptureDummy(start, length int, children []*Capture) *Capture {
	c := &Capture{
		id:       0,
		typ:      tDummy,
		off:      start,
		length:   length,
		children: children,
	}
	return c
}

func (c *Capture) ChildIterator(start int) func() *Capture {
	i := 0
	var subit, ret func() *Capture
	ret = func() *Capture {
		if i >= len(c.children) {
			return nil
		}
		ch := c.children[i]
		if ch.Dummy() && subit == nil {
			subit = ch.ChildIterator(ch.off)
		}
		if subit != nil {
			ch = subit()
		} else {
			i++
		}
		if ch == nil {
			subit = nil
			i++
			return ret()
		}
		return ch
	}
	return ret
}

func (c *Capture) Child(n int) *Capture {
	it := c.ChildIterator(0)
	i := 0
	for ch := it(); ch != nil; ch = it() {
		if i == n {
			return ch
		}
		i++
	}
	return nil
}

func (c *Capture) NumChildren() int {
	nchild := 0
	for _, ch := range c.children {
		if ch.Dummy() {
			nchild += ch.NumChildren()
		} else {
			nchild++
		}
	}
	return nchild
}

func (c *Capture) Start() int {
	if c.ment != nil {
		return c.ment.pos.Pos() + c.off
	}
	return c.off
}

func (c *Capture) Len() int {
	return c.length
}

func (c *Capture) End() int {
	return c.Start() + c.length
}

func (c *Capture) Dummy() bool {
	return c.typ == tDummy
}

func (c *Capture) Id() int {
	return int(c.id)
}

func (c *Capture) setMEnt(e *Entry) {
	if c.ment != nil {
		return
	}

	c.ment = e
	c.off = c.off - e.pos.Pos()

	for _, c := range c.children {
		c.setMEnt(e)
	}
}

// String returns a readable string representation of this node, showing the ID
// of this node and its children.
func (c *Capture) String() string {
	buf := &bytes.Buffer{}
	for i, c := range c.children {
		buf.WriteString(c.String())
		if i != len(c.children)-1 {
			buf.WriteString(", ")
		}
	}
	return fmt.Sprintf("{%d, [%s]}", c.id, buf.String())
}
