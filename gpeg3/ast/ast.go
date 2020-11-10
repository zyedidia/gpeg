package ast

import "github.com/zyedidia/gpeg/input"

type Node struct {
	Id         int16
	Start, End input.Pos
	Children   []*Node
}
