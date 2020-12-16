package ast

import (
	"bytes"
	"fmt"

	"github.com/zyedidia/gpeg/input"
)

type Node struct {
	Id         int16
	Start, End input.Pos
	Children   []*Node
}

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

func (n *Node) Each(fn func(*Node)) {
	fn(n)
	for _, c := range n.Children {
		c.Each(fn)
	}
}
