package pattern

import (
	"fmt"
	"strconv"
)

func Prettify(p Pattern) string {
	switch t := Get(p).(type) {
	case *LiteralNode:
		return strconv.Quote(t.Str)
	case *ClassNode:
		return fmt.Sprintf("[%s]", t.Chars.String())
	case *DotNode:
		return "."
	case *EmptyNode:
		return "\"\""
	case *AltNode:
		return fmt.Sprintf("(%s / %s)", Prettify(Get(t.Left)), Prettify(Get(t.Right)))
	case *SeqNode:
		return fmt.Sprintf("(%s %s)", Prettify(Get(t.Left)), Prettify(Get(t.Right)))
	case *StarNode:
		return fmt.Sprintf("%s*", Prettify(Get(t.Patt)))
	case *PlusNode:
		return fmt.Sprintf("%s+", Prettify(Get(t.Patt)))
	case *OptionalNode:
		return fmt.Sprintf("%s?", Prettify(Get(t.Patt)))
	case *NotNode:
		return fmt.Sprintf("!%s", Prettify(Get(t.Patt)))
	case *AndNode:
		return fmt.Sprintf("&%s", Prettify(Get(t.Patt)))
	case *CapNode:
		return fmt.Sprintf("{ %s }", Prettify(Get(t.Patt)))
	case *MemoNode:
		return fmt.Sprintf("{{ %s }}", Prettify(Get(t.Patt)))
	case *SearchNode:
		return fmt.Sprintf("search(%s)", Prettify(Get(t.Patt)))
	case *CheckNode:
		return fmt.Sprintf("check(%s)", Prettify(Get(t.Patt)))
	case *ErrorNode:
		return fmt.Sprintf("err(%s, %s)", t.Message, Prettify(Get(t.Recover)))
	case *EmptyOpNode:
		return fmt.Sprintf("empty(%v)", t.Op)
	case *GrammarNode:
		s := fmt.Sprintf("%s\n", t.Start)
		t.Inline()
		for name, patt := range t.Defs {
			s += fmt.Sprintf("%s <- %s\n", name, Prettify(Get(patt)))
		}
		return s
	case *NonTermNode:
		if t.Inlined != nil {
			return Prettify(Get(t.Inlined))
		}
		return t.Name
	}

	return "<invalid>"
}
