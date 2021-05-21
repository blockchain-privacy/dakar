package graph

import (
	"backend/cmd/cliutil"

	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// RenderReversibleGraph renders the given graph as an interactive html page
func RenderReversibleGraph(g *ReversibleGraph, filePath string) error {
	page := components.NewPage()

	page.SetLayout("flex")
	page.AddCharts(
		buildReversibleGraph(g),
	)

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	err = page.Render(io.MultiWriter(f))
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// RenderUndirectedGraph renders the given graph as an interactive html page
func RenderUndirectedGraph(g *UndirectedGraph, filePath string) error {
	page := components.NewPage()

	page.SetLayout("flex")
	page.AddCharts(
		buildUndirectedGraph(g),
	)

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	err = page.Render(io.MultiWriter(f))
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

func buildReversibleGraph(g *ReversibleGraph) *charts.Graph {
	chartGraph := charts.NewGraph()
	chartGraph.SetGlobalOptions(
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

	chartGraph.AddSeries("graph", graphNodes, links,
		charts.WithGraphChartOpts(
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 250}, Roam: true},
		),
	)
	return chartGraph
}

func buildUndirectedGraph(g *UndirectedGraph) *charts.Graph {
	chartGraph := charts.NewGraph()
	chartGraph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Height: "900px", Width: "1900px"}),
		charts.WithTitleOpts(opts.Title{Title: "Dash Source Address Graph"}),
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

	chartGraph.AddSeries("graph", graphNodes, links,
		charts.WithGraphChartOpts(
			opts.GraphChart{Force: &opts.GraphForce{Repulsion: 250}, Roam: true},
		),
	)
	return chartGraph
}

// ExportReversibleGraphToGephi writes the relationships of the given graph to disk
func ExportReversibleGraphToGephi(g *ReversibleGraph, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Println(closeErr)
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		nid := n.ID()

		var line []string
		line = append(line, strconv.FormatInt(nid, 10))

		fromNodes := g.From(nid)
		if fromNodes.Len() > 0 {
			for fromNodes.Next() {
				line = append(line, strconv.FormatInt(fromNodes.Node().ID(), 10))
			}

			if writeErr := writer.Write(line); writeErr != nil {
				fmt.Println("Cannot write to file", writeErr)
			}
		}
	}

	return nil
}

// ExportUndirectedGraphToGephi writes the relationships of the given graph to disk
func ExportUndirectedGraphToGephi(g *UndirectedGraph, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Println(closeErr)
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		nid := n.ID()

		var line []string
		line = append(line, strconv.FormatInt(nid, 10))

		fromNodes := g.From(nid)
		if fromNodes.Len() > 0 {
			for fromNodes.Next() {
				line = append(line, strconv.FormatInt(fromNodes.Node().ID(), 10))
			}

			if writeErr := writer.Write(line); writeErr != nil {
				fmt.Println("Cannot write to file", writeErr)
			}
		}
	}

	return nil
}

// ExportClustersToGephi writes all address cluster to disk
func ExportClustersToGephi(g *UndirectedGraph, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Println(closeErr)
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, cluster := range GetAllClusters(g) {
		if len(cluster) < 2 {
			continue
		}

		var line []string

		for _, address := range cluster {
			line = append(line, address)
		}

		if writeErr := writer.Write(line); writeErr != nil {
			fmt.Println("Cannot write to file", writeErr)
		}
	}

	return nil
}
