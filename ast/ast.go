package ast

import (
	"bytes"
	"fmt"

	"github.com/zyedidia/gpeg/input"
)

type Node struct {
	Id       int16
	start    input.Pos
	length   int
	Children []*Node
}

func NewNode(id int16, start input.Pos, length int, children []*Node) *Node {
	return &Node{
		Id:       id,
		start:    start,
		length:   length,
		Children: children,
	}
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

func (n *Node) Start() input.Pos {
	return n.start
}

func (n *Node) End() input.Pos {
	return n.start + input.Pos(n.length)
}

func (n *Node) Advance(k int) {
	n.start += input.Pos(k)
}
