package viz

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/zyedidia/gpeg/input"
	"github.com/zyedidia/gpeg/memo"
)

func text(n *memo.Capture, data *input.Input) string {
	str := string(data.Slice(n.Start(), n.End()))
	str = strings.ReplaceAll(str, ">", "&gt;")
	str = strings.ReplaceAll(str, "<", "&lt;")
	return strconv.Quote(strconv.QuoteToASCII(str))
}

func uniqueID(n *memo.Capture) string {
	return fmt.Sprintf("%p", n)[2:]
}

func exploreNode(n *memo.Capture, data *input.Input, ids map[int]string, graph *gographviz.Graph) {
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

// GraphAST renders the given AST to a graphviz dot graph, returned as a
// string.
func GraphAST(root []*memo.Capture, data *input.Input, ids map[string]int) string {
	graph := gographviz.NewGraph()
	graph.SetName("AST")
	graph.SetDir(false)

	revids := make(map[int]string)
	for k, v := range ids {
		revids[v] = k
	}

	for _, n := range root {
		exploreNode(n, data, revids, graph)
	}

	return graph.String()
}
