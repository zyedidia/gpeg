package main

import "github.com/fatih/color"

type Theme map[TokenType]*color.Color

var theme = Theme{
	Whitespace: color.New(),
	Class:      color.New(color.FgBlue),
	Keyword:    color.New(color.FgRed),
	Type:       color.New(color.FgBlue),
	Function:   color.New(color.FgGreen),
	Identifier: color.New(),
	String:     color.New(color.FgYellow),
	Comment:    color.New(color.FgBlack),
	Number:     color.New(color.FgMagenta),
	Annotation: color.New(color.FgCyan),
	Operator:   color.New(),
	Special:    color.New(color.FgMagenta),
	Other:      color.New(),
}
