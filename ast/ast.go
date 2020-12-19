package ast

import (
	"bytes"
	"fmt"

	"github.com/zyedidia/gpeg/input"
)

// A Node represents an AST capture node. It stores the ID of the capture, the
// start position and length, and any children.
type Node struct {
	Id       int16
	start    input.Pos
	length   int
	Children []*Node
}

// NewNode constructs a new AST node.
func NewNode(id int16, start input.Pos, length int, children []*Node) *Node {
	return &Node{
		Id:       id,
		start:    start,
		length:   length,
		Children: children,
	}
}

// String returns a readable string representation of this node, showing the ID
// of this node and its children.
func (n *Node) String() string {
	buf := &bytes.Buffer{}
	for i, c := range n.Children {
		buf.WriteString(c.String())
		if i != len(n.Children)-1 {
			buf.WriteString(", ")
		}
	}
	return fmt.Sprintf("{%d, [%s]}", n.Id, buf.String())
}

// Each applies a function to this node and all children.
func (n *Node) Each(fn func(*Node)) {
	fn(n)
	for _, c := range n.Children {
		c.Each(fn)
	}
}

// Start returns the start index of this AST capture.
func (n *Node) Start() input.Pos {
	return n.start
}

// End returns the end index of this AST capture.
func (n *Node) End() input.Pos {
	return n.start + input.Pos(n.length)
}

// Advance shifts this capture position by k.
func (n *Node) Advance(k int) {
	n.start += input.Pos(k)
}
