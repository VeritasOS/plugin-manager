// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package graph is used for generating the plugins' graph image.
package graph

import (
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/VeritasOS/plugin-manager/config"
	"github.com/VeritasOS/plugin-manager/pluginmanager"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	graphviz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

const (
	// NodeLabelFontSize is font size of node labels.
	NodeLabelFontSize float64 = 9.0
	// EdgeLabelFontSize is font size of edge labels.
	EdgeLabelFontSize float64 = 2.0
)

// GetImagePath gets the path of the image file.
func GetImagePath() string {
	return logutil.GetCurLogFile(true, false) + ".svg"
}

// GetDotFilePath gets the path of the dot file.
func GetDotFilePath() string {
	return logutil.GetCurLogFile(true, false) + ".dot"
}

var gv = graphviz.New()
var graph1 *cgraph.Graph

// ResetGraph is mainly used for unit testing.
func ResetGraph() {
	graph1 = nil
}

// prepareSubGraphName creates subgraph name using pluginType.
func prepareSubGraphName(pluginType string) string {
	return "cluster-" + pluginType
}

// InitGraph initliazes the graph data structure and invokes generateGraph.
func InitGraph(pluginType string, pluginsInfo map[string]*pluginmanager.PluginAttributes) error {
	var err error
	if graph1 == nil {
		graph1, err = gv.Graph()
		if err != nil {
			log.Fatal(err)
		}
	}

	sb := graph1.SubGraph(prepareSubGraphName(pluginType), 1)
	sb.SetLabel(pluginType)
	sb.Attr(0, "cluster", "true")
	// sb.SetBackgroundColor("red")
	sb.SetStyle(cgraph.FilledGraphStyle + "," + cgraph.RoundedGraphStyle)
	sb.SetGradientAngle(270)

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
		pluginNode.SetShape(cgraph.BoxShape) // Box3DShape
		pluginNode.SetLabel(pluginsInfo[p].Description)
		pluginNode.SetFontSize(NodeLabelFontSize)
		pluginNode.SetURL(pURL)
		pluginNode.SetStyle(cgraph.FilledNodeStyle + "," + cgraph.RoundedNodeStyle)
		pluginNode.SetGradientAngle(270)
		pluginNode.SetFillColor("#f5f5f5:#b3b3b3") // gray
		// pluginNode.Set("strokeColor", "#82b366")

		for rby := range pluginsInfo[p].RequiredBy {
			reqbyNode, err := sb.CreateNode(pluginsInfo[p].RequiredBy[rby])
			if err != nil {
				log.Printf("SubGraph.CreateNode(%s) Error: %s", pluginsInfo[p].RequiredBy[rby], err.Error())
				continue
			}
			// 'A' is RequiredBy 'B' will appear as 'A' --> 'B' to indicate that
			// 	once 'A' is complete, 'B' will start.
			_, err = sb.CreateEdge("", pluginNode, reqbyNode)
			// rbyEdge, err := sb.CreateEdge("RequiredBy", pluginNode, reqbyNode)
			if err != nil {
				log.Printf("SubGraph.CreateEdge(%s, %s) Error: %s",
					p, pluginsInfo[p].RequiredBy[rby], err.Error())
				continue
			}
			// rbyEdge.SetLabel("RequiredBy")
			// rbyEdge.SetFontSize(EdgeLabelFontSize)
		}
		for rs := range pluginsInfo[p].Requires {
			rsNode, err := sb.CreateNode(pluginsInfo[p].Requires[rs])
			if err != nil {
				log.Printf("SubGraph.CreateNode(%s) Error: %s", pluginsInfo[p].Requires[rs], err.Error())
				continue
			}
			// 'A' Requires 'B' will appear as 'A' <-- 'B' to indicate that
			// 	once 'B' is complete, 'A' will start.
			_, err = sb.CreateEdge("", rsNode, pluginNode)
			// rsEdge, err := sb.CreateEdge("Requires", rsNode, pluginNode)
			if err != nil {
				log.Printf("SubGraph.CreateEdge(%s, %s) Error: %s",
					p, pluginsInfo[p].RequiredBy[rs], err.Error())
				continue
			}
			// rsEdge.SetLabel("Requires")
			// rsEdge.SetFontSize(EdgeLabelFontSize)
		}
	}

	return GenerateGraph()
}

// GenerateGraph generates an input `.dot` file based on the fileNoExt name,
// and then generates an `.svg` image output file as fileNoExt.svg.
func GenerateGraph() error {
	svgFile := GetImagePath()

	rendererr := gv.RenderFilename(graph1, graphviz.Format(graphviz.DOT), GetDotFilePath())
	if rendererr != nil {
		log.Printf("gv.RenderFilename( , DOT) Err: %s", rendererr.Error())
		// return rendererr
	}
	rendererr = gv.RenderFilename(graph1, graphviz.SVG, svgFile)
	if rendererr != nil {
		log.Printf("gv.RenderFilename( , SVG) Err: %s", rendererr.Error())
		return rendererr
	}

	return nil
}

// getStatusColor returns the color for a given result status.
func getStatusColor(status string) string {
	// Node color
	ncolor := "#dae8fc:#7ea6e0" // blue // dStatusStart by default
	if status == pluginmanager.DStatusFail {
		ncolor = "#f8cecc:#ea6b66" // "red"
	} else if status == pluginmanager.DStatusOk {
		ncolor = "#d5e8d4:#97d077" // "green"
	} else if status == pluginmanager.DStatusSkip {
		ncolor = "yellow"
	}
	return ncolor
}

// UpdateGraph updates the plugin node with the status and url.
func UpdateGraph(subgraphName, plugin, status, url string) error {
	sb := graph1.SubGraph(prepareSubGraphName(subgraphName), 0)
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
	if url != "" {
		node.SetURL(url)
	}
	//  TODO Graph: Commenting until concurrency is supported in RenderFilename() of GenerateGraph().
	// return GenerateGraph()
	return nil
}
