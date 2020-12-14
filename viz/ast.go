package viz

import (
	"fmt"
	"strconv"

	"github.com/awalterschulze/gographviz"
	"github.com/zyedidia/gpeg/ast"
)

func text(n *ast.Node, data []byte) string {
	str := data[n.Start:n.End]
	return strconv.QuoteToASCII(fmt.Sprintf("'%s'", string(str)))
}

func uniqueID(n *ast.Node) string {
	return fmt.Sprintf("%p", n)[2:]
}

func exploreNode(n *ast.Node, data []byte, ids map[int16]string, graph *gographviz.Graph) {
	graph.AddNode("AST", uniqueID(n), map[string]string{
		"label": fmt.Sprintf("%s", ids[n.Id]),
		"shape": "Mrecord",
		"color": "black",
	})

	for _, c := range n.Children {
		exploreNode(c, data, ids, graph)
		graph.AddEdge(uniqueID(n), uniqueID(c), false, map[string]string{
			"color": "black",
		})
	}

	if len(n.Children) == 0 {
		textID := uniqueID(n) + "_text"
		graph.AddNode("AST", textID, map[string]string{
			"label": text(n, data),
			"shape": "Mrecord",
			"color": "black",
		})
		graph.AddEdge(uniqueID(n), textID, false, map[string]string{
			"color": "black",
		})
	}
}

func GraphAST(root []*ast.Node, data []byte, ids map[string]int16) string {
	graph := gographviz.NewGraph()
	graph.SetName("AST")
	graph.SetDir(false)

	revids := make(map[int16]string)
	for k, v := range ids {
		revids[v] = k
	}

	for _, n := range root {
		exploreNode(n, data, revids, graph)
	}

	return graph.String()
}
