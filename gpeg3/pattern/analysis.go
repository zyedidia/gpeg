package pattern

import (
	"fmt"
	"reflect"

	"github.com/awalterschulze/gographviz"
	"github.com/zyedidia/gpeg/isa"
)

// Histogram returns the number of times each instruction occurs in the given
// parsing program.
func Histogram(p isa.Program) map[reflect.Type]int {
	hist := make(map[reflect.Type]int)
	for _, insn := range p {
		t := reflect.TypeOf(insn)
		hist[t]++
	}
	return hist
}

func exploreCalls(def string, defs map[string]Pattern, explored map[string]bool, graph *gographviz.Graph) {
	explored[def] = true
	p := defs[def]
	WalkPattern(p, false, func(sub Pattern) {
		switch t := sub.(type) {
		case *NonTermNode:
			color := "black"
			if t.inlined != nil {
				color = "blue"
			}
			graph.AddEdge(def, t.Name, true, map[string]string{
				"color": color,
			})
			if !explored[t.Name] {
				exploreCalls(t.Name, defs, explored, graph)
			}
		}
	})
}

// Graph returns the string form of a GraphViz Dot graph displaying the
// call-structure of a grammar.
func Graph(g *GrammarNode) string {
	graph := gographviz.NewGraph()
	graph.SetName("Grammar")
	graph.SetDir(true)
	explored := make(map[string]bool)

	for d, p := range g.Defs {
		prog, _ := p.Compile()
		sz := prog.Size()
		astsz := CountSubPatterns(p)
		var color string
		if astsz >= 100 {
			color = "red"
		} else if astsz >= 10 {
			color = "black"
		} else {
			color = "green"
		}
		graph.AddNode("Grammar", d, map[string]string{
			"label": fmt.Sprintf("\"%v/%d\"", d, sz),
			"shape": "Mrecord",
			"color": color,
		})
	}

	exploreCalls(g.Start, g.Defs, explored, graph)

	return graph.String()
}
