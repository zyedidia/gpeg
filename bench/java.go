package bench

import (
	"github.com/zyedidia/gpeg/charset"
	"github.com/zyedidia/gpeg/isa"
	p "github.com/zyedidia/gpeg/pattern"
)

var (
	alpha = p.Set(charset.Range('A', 'Z').Add(charset.Range('a', 'z')))
	alnum = p.Set(charset.Range('A', 'Z').Add(charset.Range('a', 'z')).Add(charset.Range('0', '9')))

	word = p.Concat(
		p.Or(alpha, p.Literal("_")),
		p.Star(p.Or(alnum, p.Literal("_"))),
	)
)

func BlockPatt(start, end string, escape p.Pattern) p.Pattern {
	if escape != nil {
		return p.Concat(
			p.Literal(start),
			p.Star(
				p.Or(
					escape,
					p.Concat(
						p.Not(p.Literal(end)),
						p.Any(1),
					),
				),
			),
			p.Literal(end),
		)
	}

	return p.Concat(
		p.Literal(start),
		p.Star(p.Concat(
			p.Not(p.Literal(end)),
			p.Any(1),
		)),
		p.Literal(end),
	)
}

func WordMatch(words ...string) p.Pattern {
	m := make(map[string]struct{})

	for _, w := range words {
		m[w] = struct{}{}
	}

	return p.Check(word, isa.MapChecker(m))
}

// mini java grammar for picking out important pieces

var grammar = p.Grammar("S", map[string]p.Pattern{
	"S": p.Star(p.Memo(p.Or(
		p.NonTerm("Token"),
		p.Concat(
			p.Any(1),
			p.Star(p.Concat(
				p.Not(p.NonTerm("Token")),
				p.Any(1),
			)),
		),
	))),
	"Token": p.Or(
		p.NonTerm("Comment"),
		p.NonTerm("FuncQual"),
		p.NonTerm("FuncName"),
		p.NonTerm("String"),
		p.NonTerm("Newline"),
	),
	"Comment":     p.Or(p.NonTerm("LineComment"), p.NonTerm("LongComment")),
	"LineComment": p.Cap(BlockPatt("//", "\n", nil), capLineComment),
	"LongComment": BlockPatt("/*", "*/", nil),

	"FuncQual": p.Cap(WordMatch("public", "protected", "private"), capFuncQual),

	"FuncName": p.Concat(
		p.Cap(p.NonTerm("Identifier"), capFuncName),
		p.Literal("("),
	),
	"Identifier": word,

	"String": p.Cap(
		BlockPatt("\"", "\"", p.NonTerm("Escape")),
		capString,
	),
	"Escape": p.Concat(
		p.Literal("\\"),
		p.Set(charset.New([]byte{'\'', '"', 't', 'n', 'b', 'f', 'r', '\\'})),
	),

	"Newline": p.Cap(p.Literal("\n"), capNewline),
})

const (
	capLineComment = iota
	capFuncName
	capFuncQual
	capString
	capNewline
)
