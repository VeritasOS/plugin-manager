// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package graph is used for generating the graph image.
package graph

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VeritasOS/plugin-manager/config"
	"github.com/VeritasOS/plugin-manager/types"
	"github.com/VeritasOS/plugin-manager/types/status"
	logger "github.com/VeritasOS/plugin-manager/utils/log"
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

// Plugin is of type types.Plugin
type Plugin = types.Plugin

// Plugins is of type types.Plugins
type Plugins = types.Plugins

// InitGraphConfig initliazes output file names.
func InitGraphConfig(imgNamePrefix string) {
	// Initialization should be done only once.
	if g.fileNoExt == "" {
		// Remove imgNamePrefix if it's end with ".log"
		imgNamePrefix = strings.TrimSuffix(imgNamePrefix, ".log")
		g.fileNoExt = imgNamePrefix + "." + time.Now().Format(time.RFC3339Nano)
	}
}

// GetImagePath gets the path of the image file.
func GetImagePath() string {
	return config.GetPMLogDir() + g.fileNoExt + ".svg"
}

// GetDotFilePath gets the path of the dot file.
func GetDotFilePath() string {
	return config.GetPMLogDir() + g.fileNoExt + ".dot"
}

// InitGraph initliazes the graph data structure and invokes generateGraph.
func InitGraph(pluginType string, pluginsInfo Plugins) error {
	InitGraphConfig(config.GetPMLogFile())

	// DOT guide: https://graphviz.gitlab.io/_pages/pdf/dotguide.pdf

	for pIdx, p := range pluginsInfo {
		pFileString := "\"" + p.Name + "\""
		absLogPath, _ := filepath.Abs(config.GetPMLogDir())
		absLibraryPath, _ := filepath.Abs(config.GetPluginsLibrary())
		relPath, _ := filepath.Rel(absLogPath, absLibraryPath)
		pURL := "\"" + filepath.FromSlash(relPath+string(os.PathSeparator)+p.Name) + "\""
		rows := []string{}
		rowsInterface, ok := g.subgraph.Load(pluginType)
		if ok {
			rows = rowsInterface.([]string)
		}
		rows = append(rows, pFileString+" [label=\""+
			strings.Replace(pluginsInfo[pIdx].Description, "\"", `\"`, -1)+
			"\",style=filled,fillcolor=lightgrey,URL="+pURL+"]")
		rows = append(rows, "\""+p.Name+"\"")
		rbyLen := len(pluginsInfo[pIdx].RequiredBy)
		if rbyLen != 0 {
			graphRow := "\"" + p.Name + "\" -> "
			for rby := range pluginsInfo[pIdx].RequiredBy {
				graphRow += "\"" + pluginsInfo[pIdx].RequiredBy[rby] + "\""
				if rby != rbyLen-1 {
					graphRow += ", "
				}
			}
			rows = append(rows, graphRow)
		}
		rsLen := len(pluginsInfo[pIdx].Requires)
		if rsLen != 0 {
			graphRow := ""
			for rs := range pluginsInfo[pIdx].Requires {
				graphRow += "\"" + pluginsInfo[pIdx].Requires[rs] + "\""
				if rs != rsLen-1 {
					graphRow += ", "
				}
			}
			graphRow += " -> \"" + p.Name + "\""
			rows = append(rows, graphRow)
		}
		g.subgraph.Store(pluginType, rows)
	}

	return generateGraph()
}

// generateGraph generates an input `.dot` file based on the fileNoExt name,
// and then generates an `.svg` image output file as fileNoExt.svg.
func generateGraph() error {
	dotFile := GetDotFilePath()
	svgFile := GetImagePath()

	fhDigraph, openerr := osutils.OsOpenFile(dotFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if openerr != nil {
		abspath, _ := filepath.Abs(dotFile)
		logger.Error.Printf("OsOpenFile(%s) Abs path: %v, err=%s", dotFile, abspath, openerr.Error())
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
		logger.Error.Printf("fhDigraph.WriteString(%s), err=%s", graphContent, writeerr.Error())
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
		logger.Error.Printf("osutils.ExecCommand(%v, %v), err=%s", cmd, cmdParams, err.Error())
	}
	if len(stdOutErr) != 0 {
		logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	}

	return err
}

// getStatusColor returns the color for a given result status.
func getStatusColor(myStatus string) string {
	// Node color
	ncolor := "blue" // status.Start by default
	if myStatus == status.Fail {
		ncolor = "red"
	} else if myStatus == status.Ok {
		ncolor = "green"
	} else if myStatus == status.Skip {
		ncolor = "yellow"
	}
	return ncolor
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

	return generateGraph()
}
