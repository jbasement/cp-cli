package resource

import (
	"fmt"
	"io"
	"os"

	"github.com/emicklei/dot"
)

type GraphPrinter struct {
	writer io.Writer
}

func NewGraphPrinter() *GraphPrinter {
	return &GraphPrinter{writer: os.Stdout}
}

func (p *GraphPrinter) Print(resources []Resource) error {
	g := dot.NewGraph(dot.Undirected)
	for _, r := range resources {
		p.printResourceGraph(g, r)
	}
	fmt.Fprintln(p.writer, g.String())
	return nil
}

func (p *GraphPrinter) printResourceGraph(g *dot.Graph, r Resource) {
	node := g.Node(getResourceID(r))
	node.Label(getResourceLabel(r))
	node.Attr("penwidth", "2")

	for _, child := range r.children {
		p.printResourceGraph(g, child)
		g.Edge(node, g.Node(getResourceID(child)))
	}
}

func getResourceID(r Resource) string {
	name := r.GetName()
	if len(name) > 24 {
		name = name[:12] + "..." + name[len(name)-12:]
	}
	kind := r.GetKind()
	return fmt.Sprintf("%s-%s", kind, name)
}

func getResourceLabel(r Resource) string {
	labelKind := r.GetKind()
	labelName := r.GetName()
	if len(labelName) > 24 {
		labelName = labelName[:12] + "..." + labelName[len(labelName)-12:]
	}
	return fmt.Sprintf("%s\n%s", labelKind, labelName)
}
