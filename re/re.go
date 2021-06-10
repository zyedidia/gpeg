// Package re provides functions for compiling 're' patterns (given as strings)
// into standard patterns.
package re

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/memo"
	"github.com/zyedidia/gpeg/pattern"
	"github.com/zyedidia/gpeg/vm"
)

var parser vm.VMCode

func init() {
	prog := pattern.MustCompile(pattern.Grammar("Pattern", grammar))
	parser = vm.Encode(prog)
}

func compile(root *memo.Capture, s string) pattern.Pattern {
	var p pattern.Pattern
	switch root.Id() {
	case idPattern:
		p = compile(root.Child(0), s)
	case idGrammar:
		nonterms := make(map[string]pattern.Pattern)
		var first string
		it := root.ChildIterator(0)
		for c := it(); c != nil; c = it() {
			k, v := compileDef(c, s)
			if first == "" {
				first = k
			}
			nonterms[k] = v
		}
		p = pattern.Grammar(first, nonterms)
	case idExpression:
		alternations := make([]pattern.Pattern, 0, root.NumChildren())
		it := root.ChildIterator(0)
		for c := it(); c != nil; c = it() {
			alternations = append(alternations, compile(c, s))
		}
		p = pattern.Or(alternations...)
	case idSequence:
		concats := make([]pattern.Pattern, 0, root.NumChildren())
		it := root.ChildIterator(0)
		for c := it(); c != nil; c = it() {
			concats = append(concats, compile(c, s))
		}
		p = pattern.Concat(concats...)
	case idPrefix:
		c := root.Child(0)
		switch c.Id() {
		case idAND:
			p = pattern.And(compile(root.Child(1), s))
		case idNOT:
			p = pattern.Not(compile(root.Child(1), s))
		default:
			p = compile(root.Child(0), s)
		}
	case idSuffix:
		if root.NumChildren() == 2 {
			c := root.Child(1)
			switch c.Id() {
			case idQUESTION:
				p = pattern.Optional(compile(root.Child(0), s))
			case idSTAR:
				p = pattern.Star(compile(root.Child(0), s))
			case idPLUS:
				p = pattern.Plus(compile(root.Child(0), s))
			}
		} else {
			p = compile(root.Child(0), s)
		}
	case idPrimary:
		switch root.Child(0).Id() {
		case idIdentifier, idLiteral, idClass:
			p = compile(root.Child(0), s)
		case idOPEN:
			p = compile(root.Child(1), s)
		case idBRACEPO:
			p = pattern.Memo(compile(root.Child(1), s))
		case idDOT:
			p = pattern.Any(1)
		}
	case idLiteral:
		lit := &bytes.Buffer{}
		it := root.ChildIterator(0)
		for c := it(); c != nil; c = it() {
			lit.WriteByte(parseChar(s[c.Start():c.End()]))
		}
		p = pattern.Literal(lit.String())
	case idClass:
		var set charset.Set
		if root.NumChildren() <= 0 {
			break
		}
		complement := false
		if root.Child(0).Id() == idCARAT {
			complement = true
		}
		it := root.ChildIterator(0)
		i := 0
		for c := it(); c != nil; c = it() {
			if i == 0 && complement {
				i++
				continue
			}
			set = set.Add(compileSet(c, s))
		}
		if complement {
			set = set.Complement()
		}
		p = pattern.Set(set)
	case idIdentifier:
		p = pattern.NonTerm(parseId(root, s))
	}
	return p
}

var special = map[byte]byte{
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'\'': '\'',
	'"':  '"',
	'[':  '[',
	']':  ']',
	'\\': '\\',
	'-':  '-',
}

func parseChar(char string) byte {
	switch char[0] {
	case '\\':
		for k, v := range special {
			if char[1] == k {
				return v
			}
		}

		i, _ := strconv.ParseInt(string(char[1:]), 8, 8)
		return byte(i)
	default:
		return char[0]
	}
}

func parseId(root *memo.Capture, s string) string {
	ident := &bytes.Buffer{}
	it := root.ChildIterator(0)
	for c := it(); c != nil; c = it() {
		ident.WriteString(s[c.Start():c.End()])
	}
	return ident.String()
}

func compileDef(root *memo.Capture, s string) (string, pattern.Pattern) {
	id := root.Child(0)
	exp := root.Child(1)
	return parseId(id, s), compile(exp, s)
}

func compileSet(root *memo.Capture, s string) charset.Set {
	switch root.NumChildren() {
	case 1:
		c := root.Child(0)
		return charset.New([]byte{parseChar(s[c.Start():c.End()])})
	case 2:
		c1, c2 := root.Child(0), root.Child(1)
		return charset.Range(parseChar(s[c1.Start():c1.End()]), parseChar(s[c2.Start():c2.End()]))
	}
	return charset.Set{}
}

func Compile(s string) (pattern.Pattern, error) {
	match, n, ast, errs := parser.Exec(strings.NewReader(s), memo.NoneTable{})
	if len(errs) != 0 {
		return nil, errs[0]
	}
	if !match {
		return nil, fmt.Errorf("Invalid PEG: failed at %d", n)
	}

	return compile(ast.Child(0), s), nil
}

func MustCompile(s string) pattern.Pattern {
	p, err := Compile(s)
	if err != nil {
		panic(err)
	}
	return p
}
