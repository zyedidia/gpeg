// Package rxconv provides functions to convert a Go regexp into a PEG so that
// it can be used for incremental parsing.
package rxconv

import (
	"fmt"
	"regexp/syntax"
	"strconv"

	"github.com/zyedidia/gpeg/charset"
	p "github.com/zyedidia/gpeg/pattern"
)

var num = 0

func uniq() string {
	num++
	return "a" + strconv.Itoa(num)
}

func star(r *syntax.Regexp, k p.Pattern) p.Pattern {
	nterm := uniq()
	nonterms := make(map[string]p.Pattern)
	nonterms[nterm] = p.Or(pi(r, p.NonTerm(nterm)), k)
	return p.Grammar(nterm, nonterms)
}

// continuation-based conversion
func pi(e *syntax.Regexp, k p.Pattern) p.Pattern {
	switch e.Op {
	case syntax.OpEmptyMatch:
		return k
	case syntax.OpLiteral:
		return p.Concat(p.Literal(string(e.Rune)), k)
	case syntax.OpCharClass:
		lits := make([]p.Pattern, 0, len(e.Rune))
		for i := 0; i < len(e.Rune); i += 2 {
			start := e.Rune[i]
			end := e.Rune[i+1]
			var patt p.Pattern
			if start < 256 && end < 256 {
				patt = p.Set(charset.Range(byte(start), byte(end)))
				lits = append(lits, p.Concat(patt, k))
			} else {
				for ; start <= end; start++ {
					lits = append(lits, p.Concat(p.Literal(string(start)), k))
				}
			}
		}
		return p.Or(lits...)
	case syntax.OpAnyChar:
		// TODO: unicode
		return p.Concat(p.Any(1), k)
	case syntax.OpAnyCharNotNL:
		return p.Concat(p.Set(charset.New([]byte{'\n'}).Complement()), k)
	case syntax.OpConcat:
		patt := k
		for i := len(e.Sub) - 1; i >= 0; i-- {
			patt = pi(e.Sub[i], patt)
		}
		return patt
	case syntax.OpAlternate:
		alts := make([]p.Pattern, 0, len(e.Sub))
		for _, s := range e.Sub {
			alts = append(alts, pi(s, k))
		}
		return p.Or(alts...)
	case syntax.OpCapture:
		return pi(e.Sub[0], k)
	case syntax.OpStar:
		return star(e.Sub[0], k)
	case syntax.OpPlus:
		return pi(e.Sub[0], star(e.Sub[0], k))
	case syntax.OpQuest:
		return p.Or(pi(e.Sub[0], k), k)
	case syntax.OpBeginLine:
		return p.Concat(p.EmptyOp(syntax.EmptyBeginLine), k)
	case syntax.OpEndLine:
		return p.Concat(p.EmptyOp(syntax.EmptyEndLine), k)
	case syntax.OpBeginText:
		return p.Concat(p.EmptyOp(syntax.EmptyBeginText), k)
	case syntax.OpEndText:
		return p.Concat(p.EmptyOp(syntax.EmptyEndText), k)
	case syntax.OpWordBoundary:
		return p.Concat(p.EmptyOp(syntax.EmptyWordBoundary), k)
	case syntax.OpNoWordBoundary:
		return p.Concat(p.EmptyOp(syntax.EmptyNoWordBoundary), k)
	}
	panic(fmt.Sprintf("unimplemented %s", e.Op))
}

func convert(r *syntax.Regexp) p.Pattern {
	return pi(r, &p.EmptyNode{})
}

func FromRegexp(s string, flags syntax.Flags) (p.Pattern, error) {
	re, err := syntax.Parse(s, flags)
	if err != nil {
		return nil, err
	}
	return p.Search(p.Cap(convert(re), 0)), nil
}
