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
	parser = vm.Encode(pattern.MustCompile(pattern.Grammar("Pattern", grammar)))
}

func compile(root *memo.Capture, s string) pattern.Pattern {
	var p pattern.Pattern
	switch root.Id {
	case idPattern:
		p = compile(root.Children[0], s)
	case idGrammar:
		nonterms := make(map[string]pattern.Pattern)
		var first string
		for _, c := range root.Children {
			k, v := compileDef(c, s)
			if first == "" {
				first = k
			}
			nonterms[k] = v
		}
		p = pattern.Grammar(first, nonterms)
	case idExpression:
		alternations := make([]pattern.Pattern, 0, len(root.Children))
		for _, c := range root.Children {
			alternations = append(alternations, compile(c, s))
		}
		p = pattern.Or(alternations...)
	case idSequence:
		concats := make([]pattern.Pattern, 0, len(root.Children))
		for _, c := range root.Children {
			concats = append(concats, compile(c, s))
		}
		p = pattern.Concat(concats...)
	case idPrefix:
		c := root.Children[0]
		switch c.Id {
		case idAND:
			p = pattern.And(compile(root.Children[1], s))
		case idNOT:
			p = pattern.Not(compile(root.Children[1], s))
		default:
			p = compile(root.Children[0], s)
		}
	case idSuffix:
		if len(root.Children) == 2 {
			c := root.Children[1]
			switch c.Id {
			case idQUESTION:
				p = pattern.Optional(compile(root.Children[0], s))
			case idSTAR:
				p = pattern.Star(compile(root.Children[0], s))
			case idPLUS:
				p = pattern.Plus(compile(root.Children[0], s))
			}
		} else {
			p = compile(root.Children[0], s)
		}
	case idPrimary:
		switch root.Children[0].Id {
		case idIdentifier, idLiteral, idClass:
			p = compile(root.Children[0], s)
		case idOPEN:
			p = compile(root.Children[1], s)
		case idBRACEO:
			p = pattern.Cap(compile(root.Children[1], s))
		case idBRACEPO:
			p = pattern.Memo(compile(root.Children[1], s))
		case idDOT:
			p = pattern.Any(1)
		}
	case idLiteral:
		lit := &bytes.Buffer{}
		for _, c := range root.Children {
			lit.WriteByte(parseChar(s[c.Start():c.End()]))
		}
		p = pattern.Literal(lit.String())
	case idClass:
		var set charset.Set
		children := root.Children
		complement := false
		if children[0].Id == idCARAT {
			children = children[1:]
			complement = true
		}
		for _, c := range children {
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
	for _, c := range root.Children {
		ident.WriteString(s[c.Start():c.End()])
	}
	return ident.String()
}

func compileDef(root *memo.Capture, s string) (string, pattern.Pattern) {
	id := root.Children[0]
	exp := root.Children[1]
	return parseId(id, s), compile(exp, s)
}

func compileSet(root *memo.Capture, s string) charset.Set {
	switch len(root.Children) {
	case 1:
		c := root.Children[0]
		return charset.New([]byte{parseChar(s[c.Start():c.End()])})
	case 2:
		c1, c2 := root.Children[0], root.Children[1]
		return charset.Range(parseChar(s[c1.Start():c1.End()]), parseChar(s[c2.Start():c2.End()]))
	}
	return charset.Set{}
}

func CompilePatt(s string) (pattern.Pattern, error) {
	match, n, ast, errs := parser.Exec(strings.NewReader(s), memo.NoneTable{})
	if len(errs) != 0 {
		return nil, errs[0]
	}
	if !match {
		return nil, fmt.Errorf("Invalid PEG: failed at %d", n)
	}

	return compile(ast[0], s), nil
}

func MustCompilePatt(s string) pattern.Pattern {
	p, err := CompilePatt(s)
	if err != nil {
		panic(err)
	}
	return p
}
