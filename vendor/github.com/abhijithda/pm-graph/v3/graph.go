// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package pm graph is used for generating the graph image.
package pg

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/VeritasOS/plugin-manager/config"
	"github.com/VeritasOS/plugin-manager/pluginmanager"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	graphviz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// myGraph of plugin and its dependencies.
type myGraph struct {
	// fileNoExt is the name of the graph artifacts without extension.
	// 	Extensions could be added to generate input `.dot` file or output
	// 	`.svg` images.
	fileNoExt string
}

var mg myGraph

func initGraphConfig(imgNamePrefix string) {
	// Initialization should be done only once.
	if mg.fileNoExt == "" {
		mg.fileNoExt = imgNamePrefix + "." + time.Now().Format(time.RFC3339Nano)
	}
}

func GetImagePath() string {
	return config.GetPMLogDir() + mg.fileNoExt + ".svg"
}

var gv = graphviz.New()
var graph1 *cgraph.Graph

// ResetGraph is mainly used for unit testing.
func ResetGraph() {
	graph1 = nil
}

// InitGraph initliazes the graph data structure and invokes generateGraph.
func InitGraph(pluginType string, pluginsInfo map[string]*pluginmanager.PluginAttributes) error {
	initGraphConfig(config.GetPMLogFile())

	var err error
	if graph1 == nil {
		graph1, err = gv.Graph()
		if err != nil {
			log.Fatal(err)
		}
	}

	sb := graph1.SubGraph(pluginType, 1)
	sb.SetLabel(pluginType)
	// INFO: Sort the plugins so that list of dependencies generated
	// (used by documentation) doesn't change.
	// NOTE: If not sorted, then even without addition of any new plugin,
	//  the dependency file generated will keep changing and appears in
	// 	git staged list.
	orderedPluginsList := []string{}
	for p := range pluginsInfo {
		orderedPluginsList = append(orderedPluginsList, p)
	}
	sort.Strings(orderedPluginsList)
	for _, p := range orderedPluginsList {
		pluginNode, err := sb.CreateNode(p)
		if err != nil {
			log.Printf("SubGraph.CreateNode(%s) Error: %s", p, err.Error())
			continue
		}

		absLogPath, _ := filepath.Abs(config.GetPMLogDir())
		absLibraryPath, _ := filepath.Abs(config.GetPluginsLibrary())
		relPath, _ := filepath.Rel(absLogPath, absLibraryPath)
		pURL := filepath.FromSlash(relPath + string(os.PathSeparator) + p)
		pluginNode.SetLabel(pluginsInfo[p].Description)
		pluginNode.SetURL(pURL)
		pluginNode.SetStyle(cgraph.FilledNodeStyle)
		pluginNode.SetFillColor("lightgrey")

		for rby := range pluginsInfo[p].RequiredBy {
			reqbyNode, err := sb.CreateNode(pluginsInfo[p].RequiredBy[rby])
			if err != nil {
				log.Printf("SubGraph.CreateNode(%s) Error: %s", pluginsInfo[p].RequiredBy[rby], err.Error())
				continue
			}
			sb.CreateEdge("RequiredBy", pluginNode, reqbyNode)
		}
		for rs := range pluginsInfo[p].Requires {
			rsNode, err := sb.CreateNode(pluginsInfo[p].Requires[rs])
			if err != nil {
				log.Printf("SubGraph.CreateNode(%s) Error: %s", pluginsInfo[p].Requires[rs], err.Error())
				continue
			}
			sb.CreateEdge("Requires", pluginNode, rsNode)
		}
	}

	return GenerateGraph(graph1)
}

// GenerateGraph generates an input `.dot` file based on the fileNoExt name,
// and then generates an `.svg` image output file as fileNoExt.svg.
func GenerateGraph(g *cgraph.Graph) error {
	svgFile := GetImagePath()

	rendererr := gv.RenderFilename(g, graphviz.SVG, svgFile)
	if rendererr != nil {
		log.Printf("gv.RenderFilename() Err: %s", rendererr.Error())
		return rendererr
	}

	return nil
}

// getStatusColor returns the color for a given result status.
func getStatusColor(status string) string {
	// Node color
	ncolor := "blue" // dStatusStart by default
	if status == pluginmanager.DStatusFail {
		ncolor = "red"
	} else if status == pluginmanager.DStatusOk {
		ncolor = "green"
	} else if status == pluginmanager.DStatusSkip {
		ncolor = "yellow"
	}
	return ncolor
}

// UpdateGraph updates the plugin node with the status and url.
func UpdateGraph(subgraphName, plugin, status, url string) error {
	sb := graph1.SubGraph(subgraphName, 0)
	if sb == nil {
		err := logutil.PrintNLogError("Graph.SubGraph(%s, 0) returns nil. Error: Subgraph not found!", subgraphName)
		return err
	}

	node, err := sb.Node(plugin)
	if err != nil {
		err := logutil.PrintNLogError("Graph.Node(%s) Error: %s",
			plugin, err.Error())
		return err
	}
	node.SetStyle("filled")
	node.SetFillColor(getStatusColor(status))
	node.SetURL(url)

	svgFile := GetImagePath()

	// var buf bytes.Buffer
	// if err := gv.Render(graph1, "dot", &buf); err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(buf.String())

	rendererr := gv.RenderFilename(graph1, graphviz.SVG, svgFile)
	if rendererr != nil {
		log.Printf("gv.RenderFilename() Err: %s", rendererr.Error())
		return rendererr
	}
	return nil
}
