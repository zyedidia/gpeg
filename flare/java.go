package main

import (
	"github.com/zyedidia/gpeg/charset"
	p "github.com/zyedidia/gpeg/pattern"
)

var java = CreateHighlighter(map[string]p.Pattern{
	"whitespace": p.CapId(p.Plus(space), Whitespace),

	"line_comment": p.Concat(
		p.Literal("//"),
		p.Star(p.Concat(p.Not(p.Literal("\n")), p.Any(1))),
	),
	"block_comment": p.Concat(
		p.Literal("/*"),
		p.Star(
			p.Concat(p.Not(p.Literal("*/")), p.Any(1)),
		),
		p.Optional(p.Literal("*/")),
	),
	"comment": p.CapId(p.Or(
		p.NonTerm("line_comment"),
		p.NonTerm("block_comment"),
	), Comment),

	"sq_str": p.Concat(
		p.Literal("'"),
		p.Star(p.Concat(
			p.Concat(p.Not(p.Literal("'")), p.Any(1)),
			p.Optional(p.NonTerm("escape")),
		)),
		// TODO: optional?
		p.Literal("'"),
	),
	"dq_str": p.Concat(
		p.Literal("\""),
		p.Star(p.Concat(
			p.Concat(p.Not(p.Literal("\"")), p.Any(1)),
			p.Optional(p.NonTerm("escape")),
		)),
		// TODO: optional?
		p.Literal("\""),
	),
	"escape": p.CapId(p.Concat(
		p.Literal("\\"),
		p.Set(charset.New([]byte{'\'', '"', 'n', 'b', 'f', 'r', '\\'})),
	), Special),
	"string": p.CapId(p.Or(
		p.NonTerm("sq_str"),
		p.NonTerm("dq_str"),
	), String),

	"number": p.CapId(p.Concat(
		p.Or(
			float,
			integer,
		),
		p.Optional(p.Set(charset.New([]byte{'L', 'l', 'F', 'f', 'D', 'd'}))),
	), Number),

	"keyword": p.Concat(p.CapId(wordMatch(
		"abstract", "assert", "break", "case", "catch", "class", "const",
		"continue", "default", "do", "else", "enum", "extends", "final", "for",
		"goto", "if", "implements", "import", "instanceof", "interface",
		"native", "new", "package", "private", "protected", "public", "return",
		"static", "strictfp", "super", "switch", "synchronized", "this",
		"throw", "throws", "transient", "try", "while", "volatile", "true",
		"false", "null",
	), Keyword)),

	"type": p.Concat(p.CapId(wordMatch(
		"boolean", "byte", "char", "double", "float", "int", "long", "short",
		"void", "Boolean", "Byte", "Character", "Double", "Float", "Integer",
		"Long", "Short", "String",
	), Type)),

	"identifier": p.CapId(word, Identifier),
	"operator": p.CapId(p.Set(charset.New([]byte{
		'+', '-', '/', '*', '%', '<', '>', '!', '=', '^', '&', '|', '?', '~',
		':', ';', '.', '(', ')', '[', ']', '{', '}',
	})), Operator),

	"annotation": p.CapId(p.Concat(
		p.Literal("@"),
		word,
	), Annotation),

	"func": p.Concat(
		p.CapId(word, Function),
		p.And(p.Literal("(")),
	),

	"class_sequence": p.Concat(
		p.CapId(p.Literal("class"), Keyword),
		p.Plus(p.NonTerm("whitespace")),
		p.CapId(word, Class),
	),
}, []string{
	"whitespace",
	"class_sequence",
	"keyword",
	"type",
	"func",
	"identifier",
	"string",
	"comment",
	"number",
	"annotation",
	"operator",
})
