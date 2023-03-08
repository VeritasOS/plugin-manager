// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package pg (plugin graph) is used for generating the graph image.
package pg

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VeritasOS/plugin-manager/config"
	"github.com/VeritasOS/plugin-manager/pluginmanager"
	osutils "github.com/VeritasOS/plugin-manager/utils/os"
)

// graph of plugin and its dependencies.
type graph struct {
	// fileNoExt is the name of the graph artifacts without extension.
	// 	Extensions could be added to generate input `.dot` file or output
	// 	`.svg` images.
	fileNoExt string
	// subgraph contains subgraph name (i.e., cluster name) and its contents.
	//  I.e., each subgraph name is the key, and their contents would be in
	// 	an array.
	subgraph sync.Map
}

var g graph
var dotCmdPresent = true

// initGraphConfig: Initialize graph configurations
func initGraphConfig(imgNamePrefix string) {
	// Initialization should be done only once.
	if g.fileNoExt == "" {
		g.fileNoExt = imgNamePrefix + "." + time.Now().Format(time.RFC3339Nano)
	}
}

// GetImagePath returns the SVG image location.
func GetImagePath() string {
	return g.fileNoExt + ".svg"
}

func getDotFilePath() string {
	return g.fileNoExt + ".dot"
}

// InitGraph initliazes the graph data structure and invokes generateGraph.
func InitGraph(pluginType string, pluginsInfo map[string]*pluginmanager.PluginAttributes) error {
	initGraphConfig(config.GetPMLogDir() + config.GetPMLogFile())

	// DOT guide: https://graphviz.gitlab.io/_pages/pdf/dotguide.pdf

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
		pFileString := "\"" + p + "\""
		absLogPath, _ := filepath.Abs(config.GetPMLogDir())
		absLibraryPath, _ := filepath.Abs(config.GetPluginsLibrary())
		relPath, _ := filepath.Rel(absLogPath, absLibraryPath)
		pURL := "\"" + filepath.FromSlash(relPath+string(os.PathSeparator)+p) + "\""
		rows := []string{}
		rowsInterface, ok := g.subgraph.Load(pluginType)
		if ok {
			rows = rowsInterface.([]string)
		}
		rows = append(rows, pFileString+" [label=\""+
			strings.Replace(pluginsInfo[p].Description, "\"", `\"`, -1)+
			"\",style=filled,fillcolor=lightgrey,URL="+pURL+"]")
		rows = append(rows, "\""+p+"\"")
		rbyLen := len(pluginsInfo[p].RequiredBy)
		if rbyLen != 0 {
			graphRow := "\"" + p + "\" -> "
			for rby := range pluginsInfo[p].RequiredBy {
				graphRow += "\"" + pluginsInfo[p].RequiredBy[rby] + "\""
				if rby != rbyLen-1 {
					graphRow += ", "
				}
			}
			rows = append(rows, graphRow)
		}
		rsLen := len(pluginsInfo[p].Requires)
		if rsLen != 0 {
			graphRow := ""
			for rs := range pluginsInfo[p].Requires {
				graphRow += "\"" + pluginsInfo[p].Requires[rs] + "\""
				if rs != rsLen-1 {
					graphRow += ", "
				}
			}
			graphRow += " -> \"" + p + "\""
			rows = append(rows, graphRow)
		}
		g.subgraph.Store(pluginType, rows)
	}

	return GenerateGraph()
}

// GenerateGraph generates an input `.dot` file based on the fileNoExt name,
// and then generates an `.svg` image output file as fileNoExt.svg.
func GenerateGraph( /*dotFile, svgFile string*/ ) error {
	dotFile := getDotFilePath()
	svgFile := GetImagePath()

	fhDigraph, openerr := osutils.OsOpenFile(dotFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if openerr != nil {
		abspath, _ := filepath.Abs(dotFile)
		log.Printf("OsOpenFile(%s) Abs path: %v, Error: %s",
			dotFile, abspath, openerr.Error())
		return openerr
	}
	defer fhDigraph.Close()
	clusterCnt := 0
	graphContent := "digraph {\n"
	g.subgraph.Range(func(name interface{}, rows interface{}) bool {
		graphContent += "\nsubgraph cluster_" + strconv.Itoa(clusterCnt) + " {\n" +
			"label=\"" + name.(string) + " plugins\"\nlabelloc=t\nfontsize=24\n" +
			"node [shape=polygon,sides=6,style=filled,fillcolor=red]\n" +
			strings.Join(rows.([]string), "\n") + "\n}\n"
		clusterCnt++
		return true
	})
	graphContent += "\n}\n"

	_, writeerr := fhDigraph.WriteString(graphContent)
	if writeerr != nil {
		log.Printf("fhDigraph.WriteString(%s) Err: %s", graphContent, writeerr.Error())
		return writeerr
	}

	// https://graphviz.gitlab.io/_pages/doc/info/command.html
	cmdStr := "dot"
	// If cmdStr is not installed on system, then just return.
	if !dotCmdPresent {
		return nil
	}
	cmdParams := []string{"-Tsvg", dotFile, "-o", svgFile}

	cmd := osutils.ExecCommand(os.ExpandEnv(cmdStr), cmdParams...)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found in $PATH") {
			dotCmdPresent = false
			return nil
		}
		log.Printf("osutils.ExecCommand(%v, %v) Error: %s", cmd, cmdParams, err.Error())
	}
	if len(stdOutErr) != 0 {
		log.Println("Stdout & Stderr:", string(stdOutErr))
	}

	return err
}

// UpdateGraph updates the plugin node with the status and url.
func UpdateGraph(subgraphName, plugin, status, url string) error {
	ncolor := getStatusColor(status)
	gContents := []string{}
	gContentsInterface, ok := g.subgraph.Load(subgraphName)
	if ok {
		gContents = gContentsInterface.([]string)
	}
	gContents = append(gContents,
		"\""+plugin+"\" [style=filled,fillcolor="+ncolor+",URL=\""+url+"\"]")
	g.subgraph.Store(subgraphName, gContents)

	return GenerateGraph()
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
