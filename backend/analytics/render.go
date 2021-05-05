package analytics

import (
	"backend/cmd/cliutil"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"gonum.org/v1/gonum/graph/simple"
	"io"
	"os"
	"strconv"
)

func RenderGraph(g *simple.DirectedGraph) error {
	page := components.NewPage()

	page.SetLayout("flex")
	page.AddCharts(
		graphBase(g),
	)

	f, err := os.Create("/home/dark/Downloads/graph.html")
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	err = page.Render(io.MultiWriter(f))
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

func graphBase(g *simple.DirectedGraph) *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Height: "900px", Width: "1900px"}),
		charts.WithTitleOpts(opts.Title{Title: "Dash Mixing Graph"}),
	)

	var graphNodes []opts.GraphNode
	var links []opts.GraphLink
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		nid := n.ID()
		graphNodes = append(graphNodes, opts.GraphNode{
			Name: strconv.FormatInt(nid, 10),
		})

		fromNodes := g.From(nid)
		for fromNodes.Next() {
			links = append(links, opts.GraphLink{
				Source: strconv.FormatInt(nid, 10),
				Target: strconv.FormatInt(fromNodes.Node().ID(), 10),
			})
		}
	}

	graph.AddSeries("graph", graphNodes, links,
		charts.WithGraphChartOpts(
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 250}, Roam: true},
		),
	)
	return graph
}
