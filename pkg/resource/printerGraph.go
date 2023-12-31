package resource

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/emicklei/dot"
	"github.com/goccy/go-graphviz"
)

type GraphPrinter struct {
	writer io.Writer
}

// Initialize a new graph printer
func NewGraphPrinter() *GraphPrinter {
	return &GraphPrinter{writer: os.Stdout}
}

// Set a new graph. Gets all the nodes and then prints the graph to a file.
func (p *GraphPrinter) Print(resource Resource, fields []string, path string) error {
	g := dot.NewGraph(dot.Undirected)
	p.printResourceGraph(g, resource, fields)

	// Save graph to file
	g1 := graphviz.New()
	dotBytes := []byte(g.String())
	graph, err := graphviz.ParseBytes(dotBytes)
	if err != nil {
		return fmt.Errorf("Couldn't create PNG -> %w", err)
	}

	if err := g1.RenderFilename(graph, graphviz.PNG, path); err != nil {
		return fmt.Errorf("Couldn't save PNG to path %s -> %w", path, err)
	}
	return nil
}

// Iteratre over resources and set ID and label(content) of each node
func (p *GraphPrinter) printResourceGraph(g *dot.Graph, r Resource, fields []string) {
	node := g.Node(getResourceID(r))
	node.Label(getResourceLabel(r, fields))
	node.Attr("penwidth", "2")

	for _, child := range r.children {
		p.printResourceGraph(g, child, fields)
		g.Edge(node, g.Node(getResourceID(child)))
	}
}

// Set individual resourceID for node
func getResourceID(r Resource) string {
	name := r.GetName()
	if len(name) > 24 {
		name = name[:12] + "..." + name[len(name)-12:]
	}
	kind := r.GetKind()
	return fmt.Sprintf("%s-%s", kind, name)
}

// This functions sets the label (the actual content) of the nodes in a graph.
// Fields are defined by the fields string.
func getResourceLabel(r Resource, fields []string) string {

	var label = make([]string, len(fields))
	for i, field := range fields {
		if field == "name" {
			label[i] = field + ": " + r.GetName()
		}
		if field == "kind" {
			label[i] = field + ": " + r.GetKind()
		}
		if field == "namespace" {
			label[i] = field + ": " + r.GetNamespace()
		}
		if field == "apiversion" {
			label[i] = field + ": " + r.GetApiVersion()
		}
		if field == "synced" {
			label[i] = field + ": " + r.GetConditionStatus("Synced")
		}
		if field == "ready" {
			label[i] = field + ": " + r.GetConditionStatus("Ready")
		}
		if field == "message" {
			label[i] = field + ": " + r.GetConditionMessage()
		}
		if field == "event" {
			label[i] = field + ": " + r.GetEvent()
		}
	}

	return strings.Join(label, "\n")
}
