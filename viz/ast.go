package viz

import (
	"github.com/awalterschulze/gographviz"
	"github.com/zyedidia/gpeg/ast"
)

func exploreNode(n *ast.Node, ids map[int16]string, graph *gographviz.Graph) {
	graph.AddNode("AST", ids[n.Id], map[string]string{
		"label": ids[n.Id],
		"shape": "Mrecord",
		"color": "black",
	})

	for _, c := range n.Children {
		exploreNode(n, ids, graph)
		graph.AddEdge(ids[n.Id], ids[c.Id], false, map[string]string{
			"color": "black",
		})
	}
}

func GraphAST(root []*ast.Node, ids map[string]int16) string {
	graph := gographviz.NewGraph()
	graph.SetName("AST")
	graph.SetDir(false)

	revids := make(map[int16]string)
	for k, v := range ids {
		revids[v] = k
	}

	for _, n := range root {
		exploreNode(n, revids, graph)
	}

	return graph.String()
}
