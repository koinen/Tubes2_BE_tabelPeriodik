package main

import (
	"bytes"
	"context"
	"io/fs"
	"io/ioutil"

	gv "github.com/goccy/go-graphviz"
)

func render() {
	ctx := context.Background()
	g, _ := gv.New(ctx)
	graph, _ := g.Graph()
	a, _ := graph.CreateNodeByName("a")
	b, _ := graph.CreateNodeByName("b")
	c, _ := graph.CreateNodeByName("c")
	d, _ := graph.CreateNodeByName("d")
	e, _ := graph.CreateNodeByName("e")
	f, _ := graph.CreateNodeByName("f")
	a.Root().CreateEdgeByName("a-b", a, b)
	a.Root().CreateEdgeByName("a-c", a, c)
	b.Root().CreateEdgeByName("b-c", b, c)
	a.Root().CreateEdgeByName("a-d", a, d)
	d.Root().CreateEdgeByName("d-e", d, e)
	e.Root().CreateEdgeByName("e-f", e, f)
	buff := new(bytes.Buffer)
	g.Render(ctx, graph, "svg", buff)
	ioutil.WriteFile("./data/sample.svg", buff.Bytes(), fs.ModePerm)
}
