package analytics

import (
	"encoding/csv"
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"log"
	"os"
	"strconv"
)

func ExportToGephi(filePath string, g *simple.DirectedGraph) error {
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
