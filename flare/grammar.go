package main

import (
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
	p "github.com/zyedidia/gpeg/pattern"
)

const (
	Whitespace = iota
	Class
	Keyword
	Type
	Function
	Identifier
	String
	Comment
	Number
	Annotation
	Operator
	Special
	Other
)

type TokenType byte

func (t TokenType) String() string {
	switch t {
	case Whitespace:
		return "Whitespace"
	case Class:
		return "Class"
	case Keyword:
		return "Keyword"
	case Type:
		return "Type"
	case Function:
		return "Function"
	case Identifier:
		return "Identifier"
	case String:
		return "String"
	case Comment:
		return "Comment"
	case Number:
		return "Number"
	case Annotation:
		return "Annotation"
	case Operator:
		return "Operator"
	}
	return "Other"
}

var (
	alpha  = p.Set(charset.Range('A', 'Z').Add(charset.Range('a', 'z')))
	alnum  = p.Set(charset.Range('A', 'Z').Add(charset.Range('a', 'z')).Add(charset.Range('0', '9')))
	digit  = p.Set(charset.Range('0', '9'))
	xdigit = p.Set(charset.Range('0', '9').Add(charset.Range('A', 'F')).Add(charset.Range('a', 'f')))
	space  = p.Set(charset.New([]byte{9, 10, 11, 12, 13, ' '}))

	dec_num = p.Plus(digit)
	hex_num = p.Concat(
		p.Literal("0"),
		p.Set(charset.New([]byte{'x', 'X'})),
		p.Plus(xdigit),
	)
	oct_num = p.Concat(
		p.Literal("0"),
		p.Plus(p.Set(charset.New([]byte{'0', '7'}))),
	)

	integer = p.Concat(
		p.Optional(p.Set(charset.New([]byte{'+', '-'}))),
		p.Or(
			hex_num,
			oct_num,
			dec_num,
		),
	)
	float = p.Concat(
		p.Optional(p.Set(charset.New([]byte{'+', '-'}))),
		p.Or(
			p.Concat(
				p.Concat(
					p.Or(
						p.Concat(
							p.Star(digit),
							p.Literal("."),
							p.Plus(digit),
						),
						p.Concat(
							p.Plus(digit),
							p.Literal("."),
							p.Star(digit),
						),
					),
				),
				p.Optional(p.Concat(
					p.Set(charset.New([]byte{'e', 'E'})),
					p.Optional(p.Set(charset.New([]byte{'+', '-'}))),
					p.Plus(digit),
				)),
			),
			p.Concat(
				p.Plus(digit),
				p.Set(charset.New([]byte{'e', 'E'})),
				p.Optional(p.Set(charset.New([]byte{'+', '-'}))),
				p.Plus(digit),
			),
		),
	)

	word = p.Concat(
		p.Or(alpha, p.Literal("_")),
		p.Star(p.Or(alnum, p.Literal("_"))),
	)
)

func wordMatch(words ...string) p.Pattern {
	m := make(map[string]struct{})
	var chars []byte

	for _, w := range words {
		for _, c := range []byte(w) {
			chars = append(chars, c)
		}

		m[w] = struct{}{}
	}

	return p.Check(p.Plus(p.Set(charset.New(chars))), isa.MapChecker(m))
}

func CreateHighlighter(grammar map[string]p.Pattern, names []string) p.Pattern {
	var tokens []p.Pattern
	for _, k := range names {
		tokens = append(tokens, p.NonTerm(k))
	}

	grammar["top"] = p.Star(p.Memo(p.Or(
		p.NonTerm("token"),
		p.CapId(p.Concat(
			p.Any(1),
			p.Star(p.Concat(
				p.Not(p.NonTerm("token")),
				p.Any(1),
			)),
		), Other),
	)))
	grammar["token"] = p.Or(tokens...)

	return p.Grammar("top", grammar)
}
