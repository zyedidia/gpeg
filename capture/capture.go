// Package capture provides interfaces for describing values captured during a
// GPeg parse.
package capture

import (
	"bytes"
	"fmt"

	"github.com/zyedidia/gpeg/input"
)

type Locator interface {
	Start() input.Pos
	End() input.Pos
}

// A Node represents an AST capture node. It stores the ID of the capture, the
// start position and length, and any children.
type Node struct {
	Id       int16
	Loc      Locator
	Children []*Node
}

// NewNode constructs a new AST node.
func NewNode(id int16, Loc Locator, children []*Node) *Node {
	return &Node{
		Id:       id,
		Loc:      Loc,
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

// Start returns the start index of this AST capture.
func (n *Node) Start() input.Pos {
	return n.Loc.Start()
}

// End returns the end index of this AST capture.
func (n *Node) End() input.Pos {
	return n.Loc.End()
}
