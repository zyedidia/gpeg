package re

import (
	"github.com/zyedidia/gpeg/charset"
	p "github.com/zyedidia/gpeg/pattern"
)

// Pattern    <- Spacing_ (Grammar / Expression) EndOfFile_
// Grammar    <- Definition+
// Definition <- Identifier LEFTARROW Expression
//
// Expression <- Sequence (SLASH Sequence)*
// Sequence   <- Prefix*
// Prefix     <- (AND / NOT)? Suffix
// Suffix     <- Primary (QUESTION / STAR / PLUS)?
// Primary    <- Identifier !LEFTARROW
// 			/ '(' Expression ')'
// 			/ Literal / Class
// 			/ BRACEPO Expression BRACEPC
// 			/ BRACEO Expression BRACEC
// 			/ DOT
//
// Identifier <- IdentStart IdentCont* Spacing_
// IdentStart <- [a-zA-Z_]
// IdentCont  <- IdentStart / [0-9]
//
// Literal    <- ['] (!['] Char)* ['] Spacing_
// 			/ ["] (!["] Char)* ["] Spacing_
// Class      <- '[' CARAT? (!']' Range)* ']' Spacing_
// Range      <- Char '-' Char / Char
// Char       <- '\\' [nrt'"\[\]\\\-]
// 			/ '\\' [0-2][0-7][0-7]
// 			/ '\\' [0-7][0-7]?
// 			/ !'\\' .
//
// AND        <- '&' Spacing_
// NOT        <- '!' Spacing_
// QUESTION   <- '?' Spacing_
// STAR       <- '*' Spacing_
// PLUS       <- '+' Spacing_
// DOT        <- '.' Spacing_
// CARAT      <- '^' Spacing_
// BRACEO     <- '{' Spacing_
// BRACEC     <- '}' Spacing_
// BRACEPO    <- '{{' Spacing_
// BRACEPC    <- '}}' Spacing_
// LEFTARROW  <- '<-' Spacing_
// OPEN       <- '(' Spacing_
// CLOSE      <- ')' Spacing_
// SLASH      <- '/' Spacing_
//
// Spacing_   <- (Space_ / Comment_)*
// Comment_   <- '#' (!EndOfLine_ .)* EndOfLine_
// Space_     <- ' ' / '\t' / EndOfLine_
// EndOfLine_ <- '\r\n' / '\n' / '\r'
// EndOfFile_ <- !.

const (
	idPattern = iota
	idGrammar
	idDefinition
	idExpression
	idSequence
	idPrefix
	idSuffix
	idPrimary
	idLiteral
	idRange
	idClass
	idIdentifier
	idIdentStart
	idIdentCont
	idChar
	idAND
	idNOT
	idQUESTION
	idSTAR
	idPLUS
	idDOT
	idCARAT
	idOPEN
	idBRACEO
	idBRACEPO
)

var grammar = map[string]p.Pattern{
	"Pattern": p.Cap(p.Concat(
		p.NonTerm("Spacing"),
		p.Or(
			p.NonTerm("Grammar"),
			p.NonTerm("Expression"),
		),
		p.NonTerm("EndOfFile"),
	), idPattern),
	"Grammar": p.Cap(p.Plus(p.NonTerm("Definition")), idGrammar),
	"Definition": p.Cap(p.Concat(
		p.NonTerm("Identifier"),
		p.NonTerm("LEFTARROW"),
		p.NonTerm("Expression"),
	), idDefinition),

	"Expression": p.Cap(p.Concat(
		p.NonTerm("Sequence"),
		p.Star(p.Concat(
			p.NonTerm("SLASH"),
			p.NonTerm("Sequence"),
		)),
	), idExpression),
	"Sequence": p.Cap(p.Star(p.NonTerm("Prefix")), idSequence),
	"Prefix": p.Cap(p.Concat(
		p.Optional(p.Or(
			p.NonTerm("AND"),
			p.NonTerm("NOT"),
		)),
		p.NonTerm("Suffix"),
	), idPrefix),
	"Suffix": p.Cap(p.Concat(
		p.NonTerm("Primary"),
		p.Optional(p.Or(
			p.NonTerm("QUESTION"),
			p.NonTerm("STAR"),
			p.NonTerm("PLUS"),
		)),
	), idSuffix),
	"Primary": p.Cap(p.Or(
		p.Concat(
			p.NonTerm("Identifier"),
			p.Not(p.NonTerm("LEFTARROW")),
		),
		p.Concat(
			p.NonTerm("OPEN"),
			p.NonTerm("Expression"),
			p.NonTerm("CLOSE"),
		),
		p.Concat(
			p.NonTerm("BRACEPO"),
			p.NonTerm("Expression"),
			p.NonTerm("BRACEPC"),
		),
		p.Concat(
			p.NonTerm("BRACEO"),
			p.NonTerm("Expression"),
			p.NonTerm("BRACEC"),
		),
		p.NonTerm("Literal"),
		p.NonTerm("Class"),
		p.NonTerm("DOT"),
	), idPrimary),

	"Identifier": p.Cap(p.Concat(
		p.NonTerm("IdentStart"),
		p.Star(p.NonTerm("IdentCont")),
		p.NonTerm("Spacing"),
	), idIdentifier),
	"IdentStart": p.Cap(
		p.Set(charset.Range('a', 'z').
			Add(charset.Range('A', 'Z')).
			Add(charset.New([]byte{'_'})),
		), idIdentStart),
	"IdentCont": p.Cap(p.Or(
		p.NonTerm("IdentStart"),
		p.Set(charset.Range('0', '9')),
	), idIdentCont),

	"Literal": p.Cap(p.Or(
		p.Concat(
			p.Literal("'"),
			p.Star(p.Concat(
				p.Not(p.Literal("'")),
				p.NonTerm("Char"),
			)),
			p.Literal("'"),
			p.NonTerm("Spacing"),
		),
		p.Concat(
			p.Literal("\""),
			p.Star(p.Concat(
				p.Not(p.Literal("\"")),
				p.NonTerm("Char"),
			)),
			p.Literal("\""),
			p.NonTerm("Spacing"),
		),
	), idLiteral),
	"Class": p.Cap(p.Concat(
		p.Literal("["),
		p.Optional(p.NonTerm("CARAT")),
		p.Star(p.Concat(
			p.Not(p.Literal("]")),
			p.NonTerm("Range"),
		)),
		p.Literal("]"),
		p.NonTerm("Spacing"),
	), idClass),
	"Range": p.Cap(p.Or(
		p.Concat(
			p.NonTerm("Char"),
			p.Literal("-"),
			p.NonTerm("Char"),
		),
		p.NonTerm("Char"),
	), idRange),
	"Char": p.Cap(p.Or(
		p.Concat(
			p.Literal("\\"),
			p.Set(charset.New([]byte{'n', 'r', 't', '\'', '"', '[', ']', '\\', '-'})),
		),
		p.Concat(
			p.Literal("\\"),
			p.Set(charset.Range('0', '2')),
			p.Set(charset.Range('0', '7')),
			p.Set(charset.Range('0', '7')),
		),
		p.Concat(
			p.Literal("\\"),
			p.Set(charset.Range('0', '7')),
			p.Optional(p.Set(charset.Range('0', '7'))),
		),
		p.Concat(
			p.Not(p.Literal("\\")),
			p.Any(1),
		),
	), idChar),

	"AND": p.Cap(p.Concat(
		p.Literal("&"),
		p.NonTerm("Spacing"),
	), idAND),
	"NOT": p.Cap(p.Concat(
		p.Literal("!"),
		p.NonTerm("Spacing"),
	), idNOT),
	"QUESTION": p.Cap(p.Concat(
		p.Literal("?"),
		p.NonTerm("Spacing"),
	), idQUESTION),
	"STAR": p.Cap(p.Concat(
		p.Literal("*"),
		p.NonTerm("Spacing"),
	), idSTAR),
	"PLUS": p.Cap(p.Concat(
		p.Literal("+"),
		p.NonTerm("Spacing"),
	), idPLUS),
	"DOT": p.Cap(p.Concat(
		p.Literal("."),
		p.NonTerm("Spacing"),
	), idDOT),
	"CARAT": p.Cap(p.Concat(
		p.Literal("^"),
		p.NonTerm("Spacing"),
	), idCARAT),
	"OPEN": p.Cap(p.Concat(
		p.Literal("("),
		p.NonTerm("Spacing"),
	), idOPEN),
	"CLOSE": p.Concat(
		p.Literal(")"),
		p.NonTerm("Spacing"),
	),
	"BRACEO": p.Cap(p.Concat(
		p.Literal("{"),
		p.NonTerm("Spacing"),
	), idBRACEO),
	"BRACEC": p.Concat(
		p.Literal("}"),
		p.NonTerm("Spacing"),
	),
	"BRACEPO": p.Cap(p.Concat(
		p.Literal("{{"),
		p.NonTerm("Spacing"),
	), idBRACEPO),
	"BRACEPC": p.Concat(
		p.Literal("}}"),
		p.NonTerm("Spacing"),
	),
	"SLASH": p.Concat(
		p.Literal("/"),
		p.NonTerm("Spacing"),
	),
	"LEFTARROW": p.Concat(
		p.Literal("<-"),
		p.NonTerm("Spacing"),
	),

	"Spacing": p.Star(p.Or(
		p.NonTerm("Space"),
		p.NonTerm("Comment"),
	)),
	"Comment": p.Concat(
		p.Literal("#"),
		p.Star(p.Concat(
			p.Not(p.NonTerm("EndOfLine")),
			p.Any(1),
		)),
		p.NonTerm("EndOfLine"),
	),
	"Space": p.Or(
		p.Set(charset.New([]byte{' ', '\t'})),
		p.NonTerm("EndOfLine"),
	),
	"EndOfLine": p.Or(
		p.Literal("\r\n"),
		p.Literal("\n"),
		p.Literal("\r"),
	),
	"EndOfFile": p.Not(p.Any(1)),
}
