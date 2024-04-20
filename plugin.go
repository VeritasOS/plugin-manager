// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package pm defines Plugin Manager (PM) functions like executing
// all plugins of a particular plugin type.
package pm

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/VeritasOS/plugin-manager/config"
	"github.com/VeritasOS/plugin-manager/graph"
	"github.com/VeritasOS/plugin-manager/pluginmanager"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	osutils "github.com/VeritasOS/plugin-manager/utils/os"
	"github.com/VeritasOS/plugin-manager/utils/output"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
)

var (
	// Version of the Plugin Manager (PM).
	version = "4.6"
)

// CmdOptions contains subcommands and parameters of the pm command.
var CmdOptions struct {
	ServerCmd  *flag.FlagSet
	RunCmd     *flag.FlagSet
	ListCmd    *flag.FlagSet
	versionCmd *flag.FlagSet
	versionPtr *bool

	// portPtr indicates port number of http server to run on.
	portPtr *int

	// sequential enforces execution of plugins in sequence mode.
	// (If sequential is disabled, plugins whose dependencies are met would be executed in parallel).
	sequential *bool

	// pluginTypePtr indicates type of the plugin to run.
	pluginTypePtr *string

	// libraryPtr indicates the path of the plugins library.
	libraryPtr *string

	// pluginDirPtr indicates the location of the plugins.
	// 	NOTE: `pluginDir` is deprecated, use `library` instead.
	pluginDirPtr *string

	// workflowPtr indicates action and rollback plugin types to be run.
	workflowPtr *string
}

// getPluginFiles retrieves the plugin files under each component matching
// the specified pluginType.
func getPluginFiles(pluginType string) ([]string, error) {
	log.Println("Entering getPluginFiles")
	defer log.Println("Exiting getPluginFiles")

	var pluginFiles []string

	library := config.GetPluginsLibrary()
	if _, err := os.Stat(library); os.IsNotExist(err) {
		return pluginFiles, logutil.PrintNLogError("Library '%s' doesn't exist. "+
			"A valid plugins library path must be specified.", library)
	}
	var files []string
	dirs, err := os.ReadDir(library)
	if err != nil {
		log.Printf("os.ReadDir(%s); Error: %s", library, err.Error())
		return pluginFiles, logutil.PrintNLogError("Failed to get contents of %s plugins library.", library)
	}

	for _, dir := range dirs {
		compPluginDir := filepath.FromSlash(library + "/" + dir.Name())
		fi, err := os.Stat(compPluginDir)
		if err != nil {
			log.Printf("Unable to stat on %s directory. Error: %s\n",
				dir, err.Error())
			continue
		}
		if !fi.IsDir() {
			log.Printf("%s is not a directory.\n", compPluginDir)
			continue
		}

		tfiles, err := os.ReadDir(compPluginDir)
		if err != nil {
			log.Printf("Unable to read contents of %s directory. Error: %s\n",
				compPluginDir, err.Error())
		}
		for _, tf := range tfiles {
			files = append(files, filepath.FromSlash(dir.Name()+"/"+tf.Name()))
		}
	}

	for _, file := range files {
		matched, err := regexp.MatchString("[.]"+pluginType+"$", file)
		if err != nil {
			log.Printf("regexp.MatchString(%s, %s); Error: %s", "[.]"+pluginType, file, err.Error())
			continue
		}
		if matched {
			pluginFiles = append(pluginFiles, file)
		}
	}
	return pluginFiles, nil
}

// getPluginType returns the plugin type of the specified plugin file.
func getPluginType(file string) string {
	return strings.Replace(path.Ext(file), ".", ``, -1)
}

func getPluginsInfo(pluginType string) (*pluginmanager.Plugins, error) {
	var plugins = &pluginmanager.Plugins{}
	plugins.Attributes = make(map[string]*pluginmanager.PluginAttributes)
	pluginFiles, err := getPluginFiles(pluginType)
	if err != nil {
		return plugins, err
	}
	for _, file := range pluginFiles {
		fContents, rerr := readFile(filepath.FromSlash(
			config.GetPluginsLibrary() + file))
		if rerr != nil {
			return plugins, logutil.PrintNLogError(rerr.Error())
		}
		// log.Printf("Plugin file %s contents: \n%s\n", file, fContents)
		pInfo, perr := parseUnitFile(fContents)
		if perr != nil {
			return plugins, perr
		}
		log.Printf("Plugin %s info: %+v\n", file, pInfo)
		plugins.Attributes[file] = pInfo
	}
	return plugins, nil
}

func normalizePluginsInfo(pluginsInfo *pluginmanager.Plugins) *pluginmanager.Plugins {
	log.Println("Entering normalizePluginsInfo")
	defer log.Println("Exiting normalizePluginsInfo")

	nPInfo := &pluginmanager.Plugins{}
	nPInfo.Attributes = make(map[string]*pluginmanager.PluginAttributes)
	for pFile, pFContents := range pluginsInfo.Attributes {
		nPInfo.Attributes[pFile] = &pluginmanager.PluginAttributes{
			Description: pFContents.Description,
			ExecStart:   pFContents.ExecStart,
			FileName:    pFile,
		}
		nPInfo.Attributes[pFile].RequiredBy = append(nPInfo.Attributes[pFile].Requires, pFContents.RequiredBy...)
		nPInfo.Attributes[pFile].Requires = append(nPInfo.Attributes[pFile].Requires, pFContents.Requires...)
		log.Printf("%s plugin dependencies: %v", pFile, nPInfo.Attributes[pFile])
	}
	for p := range nPInfo.Attributes {
		log.Printf("nPInfo key(%s): %v", p, nPInfo.Attributes[p])
		for _, rs := range nPInfo.Attributes[p].Requires {
			// Check whether it's already marked as RequiredBy dependency in `Requires` plugin.
			// log.Printf("Check whether `in` (%s) already marked as RequiredBy dependency in `Requires`(%s) plugin: %v",
			// p, rs, nPInfo[rs])
			present := false
			// If dependencies are missing, then nPInfo[rs] value will not be defined.
			if nPInfo.Attributes[rs] != nil {
				log.Printf("PluginInfo for %s is present: %v", rs, nPInfo.Attributes[rs])
				for _, rby := range nPInfo.Attributes[rs].RequiredBy {
					log.Printf("p(%s) == rby(%s)? %v", p, rby, p == rby)
					if p == rby {
						present = true
						break
					}
				}
				if !present {
					nPInfo.Attributes[rs].RequiredBy = append(nPInfo.Attributes[rs].RequiredBy, p)
					log.Printf("Added %s as RequiredBy dependency of %s: %+v", p, rs, nPInfo.Attributes[rs])
				}
			}
		}

		// Check whether RequiredBy dependencies are also marked as Requires dependency on other plugin.
		log.Printf("Check whether RequiredBy dependencies are also marked as Requires dependency on other plugin.")
		for _, rby := range nPInfo.Attributes[p].RequiredBy {
			log.Printf("RequiredBy of %s: %s", p, rby)
			log.Printf("nPInfo.Attributes of %s: %+v", rby, nPInfo.Attributes[rby])
			// INFO: If one plugin type is added as dependent on another by
			// any chance, then skip checking its contents as the other
			// plugin type files were not parsed.
			if nPInfo.Attributes[rby] == nil {
				// NOTE: Add the missing plugin in Requires, So that the issue
				// gets caught during validation.
				nPInfo.Attributes[p].Requires = append(nPInfo.Attributes[p].Requires, rby)
				continue
			}
			present := false
			for _, rs := range nPInfo.Attributes[rby].Requires {
				if p == rs {
					present = true
					break
				}
			}
			if !present {
				nPInfo.Attributes[rby].Requires = append(nPInfo.Attributes[rby].Requires, p)
				log.Printf("Added %s as Requires dependency of %s: %+v", p, rby, nPInfo.Attributes[rby])
			}
		}
	}
	log.Printf("Plugins info after normalizing: \n%+v\n", nPInfo.Attributes)
	return nPInfo
}

// parseUnitFile parses the plugin file contents.
func parseUnitFile(fileContents string) (*pluginmanager.PluginAttributes, error) {
	log.Println("Entering parseUnitFile")
	defer log.Println("Exiting parseUnitFile")

	pluginInfo := &pluginmanager.PluginAttributes{}
	if len(fileContents) == 0 {
		return pluginInfo, nil
	}
	lines := strings.Split(fileContents, "\n")
	for l := range lines {
		line := strings.TrimSpace(lines[l])
		// TODO: Log as debug message
		// log.Println("line...", line)
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			// No need to parse comments.
			// TODO: Log as debug message
			// log.Println("Skipping comment line...", line)
			continue
		}

		fields := strings.Split(line, "=")
		if len(fields) == 0 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		val := strings.TrimSpace(strings.Join(fields[1:], "="))
		switch key {
		case "Description":
			pluginInfo.Description = val
			break
		case "ExecStart":
			pluginInfo.ExecStart = val
			break
		case "RequiredBy":
			pluginInfo.RequiredBy = strings.Split(val, " ")
			break
		case "Requires":
			pluginInfo.Requires = strings.Split(val, " ")
			break
		default:
			// TODO: Log as debug message
			// log.Printf("Non-standard line found: %s", line)
			break
		}
	}

	return pluginInfo, nil
}

func validateDependencies(nPInfo *pluginmanager.Plugins) ([]string, error) {
	log.Println("Entering validateDependencies")
	defer log.Println("Exiting validateDependencies")

	var pluginOrder []string
	notPlacedPlugins := []string{}
	dependencyMet := map[string]bool{}
	// for pFile, pFContents := range pluginsInfo {}
	sortedPFiles := []string{}
	for pFile := range nPInfo.Attributes {
		sortedPFiles = append(sortedPFiles, pFile)
	}
	// NOTE: Sorting plugin files mainly to have a deterministic order,
	// though it's not required for solution to work.
	// (Sorting takes care of unit tests as maps return keys/values in random order).
	sort.Strings(sortedPFiles)
	log.Printf("Plugin files in sorted order: %+v\n", sortedPFiles)

	for pFileIndex := range sortedPFiles {
		pFile := sortedPFiles[pFileIndex]
		pFContents := nPInfo.Attributes[pFile]
		log.Printf("\nFile: %s \n%+v \n\n", pFile, pFContents)
		if len(pFContents.Requires) == 0 {
			dependencyMet[pFile] = true
			pluginOrder = append(pluginOrder, pFile)
		} else {
			dependencyMet[pFile] = false
			notPlacedPlugins = append(notPlacedPlugins, pFile)
		}
	}

	curLen := len(notPlacedPlugins)
	// elementsLeft to process in the notPlacedPlugins queue!
	elementsLeft := curLen
	prevLen := curLen
	// INFO:
	// 	When all the elements are processed in the queue
	//	(i.e., `elementsLeft` becomes 0), check whether at least one of the
	//	plugin's dependency has been met (i.e., prevLen != curLen). If not,
	//	then there is a circular dependency, or plugins are missing dependencies.
	for curLen != 0 {
		pFile := notPlacedPlugins[0]
		notPlacedPlugins = notPlacedPlugins[1:]
		pDependencies := nPInfo.Attributes[pFile].Requires
		log.Printf("Plugin %s dependencies: %+v\n", pFile, pDependencies)

		dependencyMet[pFile] = true
		for w := range pDependencies {
			val := dependencyMet[pDependencies[w]]
			if !val {
				// If dependency met is false, then process it later again after all dependencies are met.
				dependencyMet[pFile] = false
				log.Printf("Adding %s back to list %s to process as %s plugin dependency is not met.\n",
					pFile, notPlacedPlugins, pDependencies[w])
				notPlacedPlugins = append(notPlacedPlugins, pFile)
				break
			}
		}
		// If dependency met is not set to false, then it means all
		// dependencies are met. So, add it to pluginOrder
		if false != dependencyMet[pFile] {
			log.Printf("Dependency met for %s: %v\n", pFile, dependencyMet[pFile])
			pluginOrder = append(pluginOrder, pFile)
		}

		elementsLeft--
		if elementsLeft == 0 {
			log.Printf("PrevLen: %d; CurLen: %d\n", prevLen, curLen)
			curLen = len(notPlacedPlugins)
			if prevLen == curLen {
				// INFO: Clear out the pluginOrder as we cannot run all the
				// 	plugins either due to missing dependencies or having
				// 	circular dependency.
				return []string{}, logutil.PrintNLogError(
					"There is either a circular dependency between plugins, "+
						"or some dependencies are missing in these plugins: %+v",
					notPlacedPlugins)
			}
			prevLen = curLen
			elementsLeft = curLen
		}
	}

	return pluginOrder, nil
}

func executePluginCmd(statusCh chan<- map[string]*pluginmanager.PluginStatus, p string, pluginsInfo *pluginmanager.Plugins, failedDependency bool, pluginLogFile string) {
	pInfo := pluginsInfo.Attributes[p]
	log.Printf("\nChannel: Plugin %s info: \n%+v\n", p, pInfo)
	start := &timestamp.Timestamp{Nanos: int32(time.Second.Nanoseconds())}

	// TODO: Uncomment below UpdateGraph() once concurrency issue is
	//  taken care, and remove the one from where executePluginCmd().
	//  is called. Refer "TODO Graph" for more details.
	// graph.UpdateGraph(getPluginType(p), p, pluginmanager.DStatusStart, "")
	logutil.PrintNLog("\n%s: %s\n", pInfo.Description, pluginmanager.DStatusStart)
	// TODO: Uncomment below logFile once UpdateGraph() concurrency issue is
	//  resolved.
	// Get relative path to plugins log file from PM log dir, so that linking
	// in plugin graph works even when the logs are copied to another system.
	// pluginLogFile := strings.Replace(config.GetPluginsLogDir(),
	// 	config.GetPMLogDir(), "", -1) +
	// 	strings.Replace(p, string(os.PathSeparator), ":", -1) +
	// 	"." + time.Now().Format(time.RFC3339Nano) + ".log"
	logFile := config.GetPMLogDir() + pluginLogFile
	fh, openerr := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if openerr != nil {
		log.Printf("os.OpenFile(%s) Error: %s", logFile, openerr.Error())
		// Ignore error and continue as plugin log file creation is not fatal.
	}
	defer fh.Close()
	// chLog is a channel logger
	chLog := log.New(fh, "", log.LstdFlags)
	chLog.SetOutput(fh)
	chLog.Println("Plugin file:", p)

	// If already marked as failed/skipped due to dependency fail,
	// then just return that status.
	myStatus := ""
	myStatusMsg := ""
	if failedDependency {
		myStatusMsg = "Skipping as its dependency failed."
		myStatus = pluginmanager.DStatusSkip
	} else if pInfo.ExecStart == "" {
		myStatusMsg = "Passing as ExecStart value is empty!"
		myStatus = pluginmanager.DStatusOk
	}

	if myStatus != "" {
		log.Println(myStatusMsg)
		chLog.Println(myStatusMsg)
		// graph.UpdateGraph(getPluginType(p), p, myStatus, "")
		logutil.PrintNLog("%s: %s\n", pInfo.Description, myStatus)
		end := &timestamp.Timestamp{Nanos: int32(time.Second.Nanoseconds())}
		statusCh <- map[string]*pluginmanager.PluginStatus{
			p: {
				Status: myStatus,
				RunTime: &pluginmanager.RunTime{
					StartTime: start,
					EndTime:   end,
					Duration:  &duration.Duration{Nanos: end.Nanos - start.Nanos},
				},
			},
		}
		return
	}

	log.Printf("\nExecuting command: %s\n", pInfo.ExecStart)
	cmdParam := strings.Split(pInfo.ExecStart, " ")
	cmdStr := cmdParam[0]
	cmdParams := os.ExpandEnv(strings.Join(cmdParam[1:], " "))
	cmdParamsExpanded := strings.Split(cmdParams, " ")

	cmd := exec.Command(os.ExpandEnv(cmdStr), cmdParamsExpanded...)
	// INFO: https://stackoverflow.com/questions/69954944/capture-stdout-from-exec-command-line-by-line-and-also-pipe-to-os-stdout
	iostdout, err := cmd.StdoutPipe()
	if err != nil {
		pStatus := pluginmanager.PluginStatus{Status: pluginmanager.DStatusFail}
		log.Printf("Failed to execute plugin %s. Error: %s\n", p, err.Error())
		logutil.PrintNLog("%s: %s\n", pInfo.Description, pluginmanager.DStatusFail)
		statusCh <- map[string]*pluginmanager.PluginStatus{p: &pStatus}
		return
	}
	cmd.Stderr = cmd.Stdout

	chLog.Println("Executing command:", pInfo.ExecStart)
	err = cmd.Start()
	var stdOutErr []string
	if err == nil {
		scanner := bufio.NewScanner(iostdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			iobytes := scanner.Text()
			chLog.Println(string(iobytes))
			stdOutErr = append(stdOutErr, iobytes)
		}
		err = cmd.Wait()
		// chLog.Printf("command exited with code: %+v", err)
	}

	func() {
		if err != nil {
			chLog.Println("Error:", err.Error())
			// graph.UpdateGraph(getPluginType(p), p, pluginmanager.DStatusFail, pluginLogFile)
		} else {
			// graph.UpdateGraph(getPluginType(p), p, pluginmanager.DStatusOk, pluginLogFile)
		}
	}()
	log.Println("Stdout & Stderr:", stdOutErr)
	pStatus := pluginmanager.PluginStatus{StdOutErr: stdOutErr}
	if err != nil {
		pStatus.Status = pluginmanager.DStatusFail
		log.Printf("Failed to execute plugin %s. Error: %s\n", p, err.Error())
		logutil.PrintNLog("%s: %s\n", pInfo.Description, pluginmanager.DStatusFail)
		statusCh <- map[string]*pluginmanager.PluginStatus{p: &pStatus}
		return
	}
	pStatus.Status = pluginmanager.DStatusOk
	logutil.PrintNLog("%s: %s\n", pInfo.Description, pluginmanager.DStatusOk)
	end := &timestamp.Timestamp{Nanos: int32(time.Second.Nanoseconds())}
	pStatus.RunTime = &pluginmanager.RunTime{
		StartTime: start,
		EndTime:   end,
		Duration:  &duration.Duration{Nanos: end.Nanos - start.Nanos},
	}
	statusCh <- map[string]*pluginmanager.PluginStatus{p: &pStatus}
}

func executePlugins(psStatus *pluginmanager.PluginTypeStatus, nPInfo *pluginmanager.Plugins, sequential bool) bool {
	log.Println("Entering executePlugins")
	defer log.Println("Exiting executePlugins")

	retStatus := true

	_, err := validateDependencies(nPInfo)
	if err != nil {
		return false
	}

	// INFO: Set the PM_LIBRARY env variable so that binaries/scripts placed
	// in the same directory as that of plugins can be accessed using
	// ${PM_LIBRARY}/<binary|script> path.
	os.Setenv("PM_LIBRARY", config.GetPluginsLibrary())

	// INFO: PM_PLUGIN_DIR is deprecated. Use PM_LIBRARY instead.
	// Keeping it for backward compatibility for couple of releases.
	// 	(Currently Px is 1.x, and we can keep say until Px 3.x).
	// If older versions of PM is out there, then PM_PLUGIN_DIR env value
	// 	will be expected in plugins.
	os.Setenv("PM_PLUGIN_DIR", config.GetPluginsLibrary())

	waitCount := map[string]int{}
	for p := range nPInfo.Attributes {
		waitCount[p] = len(nPInfo.Attributes[p].Requires)
		log.Printf("%s plugin dependencies: %+v", p, nPInfo.Attributes[p])
	}

	executingCnt := 0
	exeCh := make(chan map[string]*pluginmanager.PluginStatus)
	pluginIndexes := make(map[string]int)
	failedDependency := make(map[string]bool)
	for len(nPInfo.Attributes) > 0 || executingCnt != 0 {
		for p := range nPInfo.Attributes {
			// INFO: When all dependencies are met, plugin waitCount would be 0.
			// 	When sequential execution is enforced, even if a plugin is ready
			// 	 to run, make sure that only one plugin is running at time, by
			// 	 checking executing count is 0.
			// 	When sequential execution is not enforced, run plugins that are ready.
			if waitCount[p] == 0 && (!sequential ||
				(sequential && executingCnt == 0)) {
				log.Printf("Plugin %s is ready for execution: %v.", p, nPInfo.Attributes[p])
				waitCount[p]--

				ps := &pluginmanager.PluginStatus{}
				ps.Attributes = nPInfo.Attributes[p]
				psStatus.Plugins = append(psStatus.Plugins, ps)
				pluginIndexes[p] = len(psStatus.Plugins) - 1

				// TODO: Remove below UpdateGraph() once concurrency issue is
				//  taken care, and keep the one inside executePluginCmd().
				//  Refer "TODO Graph" for more details.
				pluginLogFile := strings.Replace(config.GetPluginsLogDir(),
					config.GetPMLogDir(), "", -1) +
					strings.Replace(p, string(os.PathSeparator), ":", -1) +
					"." + time.Now().Format(time.RFC3339Nano) + ".log"
				graph.UpdateGraph(getPluginType(p), p, pluginmanager.DStatusStart, pluginLogFile)
				go executePluginCmd(exeCh, p, nPInfo, failedDependency[p], pluginLogFile)
				executingCnt++
			}
		}
		// TODO: Remove below GenerateGraph() once concurrency issue is taken
		//  care, and keep the one inside executePluginCmd().
		//  Refer "TODO Graph" for more details.
		// INFO: Call generate graph before and waiting so as to update and
		//  display in-progress and done status in graph.
		graph.GenerateGraph()
		// start other dependent ones as soon as one of the plugin completes.
		exeStatus := <-exeCh
		executingCnt--
		for plugin, pStatus := range exeStatus {
			log.Printf("%s status: %v", plugin, pStatus.Status)
			pIdx := pluginIndexes[plugin]
			ps := psStatus.Plugins
			ps[pIdx].Status = pStatus.Status
			ps[pIdx].StdOutErr = pStatus.StdOutErr
			ps[pIdx].RunTime = &pluginmanager.RunTime{
				StartTime: pStatus.RunTime.StartTime,
				EndTime:   pStatus.RunTime.EndTime,
				Duration:  pStatus.RunTime.Duration}
			graph.UpdateGraph(getPluginType(plugin), plugin, pStatus.Status, "")
			if pStatus.Status == pluginmanager.DStatusFail {
				retStatus = false
			}

			for _, rby := range nPInfo.Attributes[plugin].RequiredBy {
				if pStatus.Status == pluginmanager.DStatusFail ||
					pStatus.Status == pluginmanager.DStatusSkip {
					// TODO: When "Wants" and "WantedBy" options are supported similar to
					// 	"Requires" and "RequiredBy", the failedDependency flag should be
					// 	checked in conjunction with if its required dependency is failed,
					// 	and not the wanted dependency.
					failedDependency[rby] = true
				}
				waitCount[rby]--
			}
			delete(nPInfo.Attributes, plugin)
		}
		// TODO: Remove below GenerateGraph() once concurrency issue is taken
		//  care. Refer "TODO Graph" for more details.
		graph.GenerateGraph()
	}
	return retStatus
}

// List the plugin and its dependencies.
func List(pluginType string, options map[string]string) error {
	var pluginsInfo, err = getPluginsInfo(pluginType)
	if err != nil {
		return err
	}
	nPInfo := normalizePluginsInfo(pluginsInfo)

	err = graph.InitGraph(pluginType, nPInfo.Attributes, options)
	if err != nil {
		return err
	}

	return nil
}

func readFile(filePath string) (string, error) {
	bFileContents, err := os.ReadFile(filePath)
	if err != nil {
		message := "Failed to read " + filePath + " file."
		err = errors.New(message)
		return "", err
	}

	return string(bFileContents), nil
}

// RegisterCommandOptions registers the command options that are supported
func RegisterCommandOptions(progname string) {
	log.Println("Entering RegisterCommandOptions")
	defer log.Println("Exiting RegisterCommandOptions")

	CmdOptions.versionCmd = flag.NewFlagSet(progname+" version", flag.ContinueOnError)
	CmdOptions.versionPtr = CmdOptions.versionCmd.Bool("version", false, "print Plugin Manager (PM) version.")

	CmdOptions.ServerCmd = flag.NewFlagSet(progname+" server", flag.PanicOnError)
	CmdOptions.portPtr = CmdOptions.ServerCmd.Int(
		"port",
		8080,
		"Port number",
	)
	logutil.RegisterCommandOptions(CmdOptions.ServerCmd, map[string]string{
		"defaultLogDir":  "./",
		"defaultLogFile": progname,
	})

	CmdOptions.RunCmd = flag.NewFlagSet(progname+" run", flag.PanicOnError)
	CmdOptions.pluginTypePtr = CmdOptions.RunCmd.String(
		"type",
		"",
		"Type of plugin.",
	)
	CmdOptions.libraryPtr = CmdOptions.RunCmd.String(
		"library",
		"",
		"Path of the plugins library.",
	)
	CmdOptions.sequential = CmdOptions.RunCmd.Bool(
		"sequential",
		false,
		"Enforce running plugins in sequential.",
	)
	CmdOptions.workflowPtr = CmdOptions.RunCmd.String(
		"workflow",
		"",
		"List of action and optional rollback plugin types in JSON format.",
	)

	logutil.RegisterCommandOptions(CmdOptions.RunCmd, map[string]string{})
	output.RegisterCommandOptions(CmdOptions.RunCmd, map[string]string{})

	CmdOptions.ListCmd = flag.NewFlagSet(progname+" list", flag.PanicOnError)
	CmdOptions.ListCmd.StringVar(
		CmdOptions.pluginTypePtr,
		"type",
		"",
		"Type of plugin.",
	)
	CmdOptions.ListCmd.StringVar(
		CmdOptions.libraryPtr,
		"library",
		"",
		"Path of the plugins library.",
	)
	CmdOptions.ListCmd.StringVar(
		CmdOptions.workflowPtr,
		"workflow",
		"",
		"List of action and rollback plugin types.",
	)
	logutil.RegisterCommandOptions(CmdOptions.ListCmd, map[string]string{})

}

// Run the specified plugin type plugins.
func Run(result *pluginmanager.PluginTypeStatus, pluginType string, options map[string]string) error {
	result.Type = pluginType
	status := true

	if err := osutils.OsMkdirAll(config.GetPluginsLogDir(), 0755); nil != err {
		err = logutil.PrintNLogError(
			"Failed to create the plugins logs directory: %s. "+
				"Error: %s", config.GetPluginsLogDir(), err.Error())
		result.Status = pluginmanager.DStatusFail
		result.StdOutErr = err.Error()
		return err
	}

	var pluginsInfo, err = getPluginsInfo(pluginType)
	if err != nil {
		result.Status = pluginmanager.DStatusFail
		result.StdOutErr = err.Error()
		return err
	}
	nPInfo := normalizePluginsInfo(pluginsInfo)
	graph.InitGraph(pluginType, nPInfo.Attributes, options)

	status = executePlugins(result, nPInfo, *CmdOptions.sequential)
	if !status {
		result.Status = pluginmanager.DStatusFail
		err = fmt.Errorf("Running %s plugins: %s", pluginType, pluginmanager.DStatusFail)
		result.StdOutErr = err.Error()
		logutil.PrintNLog("%s\n", err.Error())
		return err
	}
	result.Status = pluginmanager.DStatusOk
	logutil.PrintNLog("Running %s plugins: %s\n", pluginType, pluginmanager.DStatusOk)
	return nil
}

// RegisterHandlers defines http handlers.
func RegisterHandlers(port int) {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/run", runHandler)

	// Loading web/js files
	webPath := os.Getenv("PM_WEB")
	if webPath == "" {
		webPath = "web"
	}
	http.Handle("/web/",
		http.StripPrefix("/web/",
			http.FileServer(http.Dir(webPath))))

	// Enable viewing of overall log file and plugin logs.
	http.Handle("/log/",
		http.StripPrefix("/log/",
			http.FileServer(http.Dir(logutil.GetLogDir()))))
	http.Handle("/plugins/",
		http.StripPrefix("/plugins/",
			http.FileServer(http.Dir(config.GetPluginsLogDir()))))

	fmt.Println("Starting server on port", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		logutil.PrintNLogError(err.Error())
		log.Fatalln(err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering homePage")
	defer log.Println("Exiting homePage")

	if r.RequestURI != "/" {
		return
	}

	webPath := os.Getenv("PM_WEB")
	if webPath == "" {
		webPath = "web"
	}
	tmpl := template.Must(template.ParseFiles(webPath + "/pm.html"))
	tmpl.Execute(w, nil)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering listHandler")
	defer log.Println("Exiting listHandler")

	// Create new log file with same name but with new timestamp.
	logutil.SetLogging(logutil.GetCurLogFile(false, false))
	graph.ResetGraph()

	// queryParams := r.URL.Query()
	// fmt.Println("Query Params: ", queryParams)
	// pluginType := queryParams["type"]
	// library := queryParams["library"]
	// pluginType := r.PostFormValue("type")
	library := r.PostFormValue("library")
	fmt.Println("Library:", library)

	config.SetPluginsLibrary(library)

	var err error
	r.ParseForm()
	// INFO: pluginTypes could be either a single element of comma or space separated list, or multiple elements - all in the array.
	userPluginTypes := r.PostForm["type"]
	if len(userPluginTypes) > 0 {
		// INFO: pluginTypes could be either a single element of comma or space separated list, or multiple elements - all in the array.
		seps := " ,"
		splitter := func(r rune) bool {
			return strings.ContainsRune(seps, r)
		}
		pluginTypes := []string{}
		for _, pt := range userPluginTypes {
			pluginTypes = append(pluginTypes, strings.FieldsFunc(pt, splitter)...)
		}
		fmt.Printf("Plugin Types(%d): %+v\n", len(pluginTypes), pluginTypes)

		listFunc := func() {
			for idx, pluginType := range pluginTypes {
				fmt.Printf("\nListing %v plugins...\n", pluginType)
				err = List(pluginType, nil)
				if err != nil {
					fmt.Fprintf(w, "Error: %s", err.Error())
					return
				}
				if idx > 0 {
					graph.ConnectGraph(pluginTypes[idx-1], pluginType)
				}
			}
		}
		go listFunc()
	}

	userWorkflow := r.PostForm["workflow"]
	if len(userWorkflow) > 0 {
		fmt.Printf("User Workflow (%v): %+v\n", len(userWorkflow), userWorkflow)

		var workflow pluginmanager.Workflow
		for _, ar := range userWorkflow {
			ar = strings.TrimSpace(ar)
			if ar == "" {
				continue
			}
			var pAR pluginmanager.ActionRollback
			json.Unmarshal([]byte(ar), &pAR)
			fmt.Printf("Unmarshal(%+v) = %+v\n", userWorkflow, &pAR)
			workflow.ActionRollbacks = append(workflow.ActionRollbacks, &pAR)
		}
		if len(workflow.ActionRollbacks) > 0 {
			json.Unmarshal([]byte(userWorkflow[0]), &workflow.ActionRollbacks)
			fmt.Printf("Received workflow request: %+v\n", &workflow)

			workflowFunc := func() {
				err = triggerWorkflow("list", &workflow)
			}
			go workflowFunc()
		}
	}

	if err != nil {
		fmt.Fprintf(w, "Error: \n%v", err.Error())
	} else {
		webPath := os.Getenv("PM_WEB")
		if webPath == "" {
			webPath = "web"
		}
		tmpl := template.Must(template.ParseFiles(webPath + "/run-response.html"))
		// Get relative path of log file from log dir, so that the handler can
		// 	server it under "/log" path.
		tmpl.Execute(w, "/log/"+filepath.Base(logutil.GetCurLogFile(true, false)))
	}
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering runHandler")
	defer log.Println("Exiting runHandler")

	// Create new log file with same name but with new timestamp.
	logutil.SetLogging(logutil.GetCurLogFile(false, false))
	graph.ResetGraph()

	// pluginType := r.PostFormValue("type")
	// fmt.Println("Type:", pluginType)
	r.ParseForm()
	library := r.PostFormValue("library")
	fmt.Println("Library:", library)

	config.SetPluginsLibrary(library)

	userEnvs := r.PostForm["env"]
	fmt.Printf("User Parameters (ENV) %d: %+v\n", len(userEnvs), userEnvs)
	if len(userEnvs) > 0 {
		for _, userEnv := range userEnvs {
			userEnv = strings.TrimSpace(userEnv)
			fmt.Println("userEnv: ", userEnv)
			lines := strings.Split(userEnv, "\n")
			fmt.Printf("Lines (%+v): %v\n", len(lines), lines)
			for _, line := range lines {
				line := strings.TrimSpace(line)
				fmt.Printf("Line (%+v): %v\n", len(line), line)
				envVar, envVal, status := strings.Cut(line, "=")
				if status {
					fmt.Println("ENV: ", envVar, "=", envVal)
					os.Setenv(envVar, envVal)
				}
			}
		}
	}

	userPluginTypes := r.PostForm["type"]
	if len(userPluginTypes) > 0 {
		// INFO: pluginTypes could be either a single element of comma or space separated list, or multiple elements - all in the array.
		seps := " ,"
		splitter := func(r rune) bool {
			return strings.ContainsRune(seps, r)
		}
		pluginTypes := []string{}
		for _, pt := range userPluginTypes {
			pluginTypes = append(pluginTypes, strings.FieldsFunc(pt, splitter)...)
		}
		fmt.Printf("Plugin Types(%d): %+v\n", len(pluginTypes), pluginTypes)

		pmstatus := pluginmanager.PluginTypeStatus{}

		runFunc := func() {
			// fmt.Println("Inside runFunc routine...")
			for idx, pluginType := range pluginTypes {
				fmt.Printf("\nRunning %v plugins...\n", pluginType)
				err := Run(&pmstatus, pluginType, nil)
				if err != nil {
					fmt.Fprintf(w, "Error: %s", err.Error())
					return
				}
				if idx > 0 {
					graph.ConnectGraph(pluginTypes[idx-1], pluginType)
				}
			}
		}
		go runFunc()

	}

	userWorkflow := r.PostForm["workflow"]
	if len(userWorkflow) > 0 {
		fmt.Printf("User Workflow (%v): %+v\n", len(userWorkflow), userWorkflow)

		var workflow pluginmanager.Workflow
		for _, ar := range userWorkflow {
			ar = strings.TrimSpace(ar)
			if ar == "" {
				continue
			}
			var pAR pluginmanager.ActionRollback
			json.Unmarshal([]byte(ar), &pAR)
			// pAR := &pluginmanager.ActionRollback{}
			// json.Unmarshal([]byte(ar), pAR)
			fmt.Printf("Unmarshal(%+v) = %+v\n", userWorkflow, &pAR)
			workflow.ActionRollbacks = append(workflow.ActionRollbacks, &pAR)
		}
		if len(workflow.ActionRollbacks) > 0 {
			json.Unmarshal([]byte(userWorkflow[0]), &workflow.ActionRollbacks)
			fmt.Printf("Received workflow request: %+v\n", workflow.ActionRollbacks)

			workflowFunc := func() {
				triggerWorkflow("run", &workflow)
			}
			go workflowFunc()
		}
	}

	webPath := os.Getenv("PM_WEB")
	if webPath == "" {
		webPath = "web"
	}
	tmpl := template.Must(template.ParseFiles(webPath + "/run-response.html"))
	// Get relative path of log file from log dir, so that the handler can
	// 	server it under "/log" path.
	tmpl.Execute(w, "/log/"+filepath.Base(logutil.GetCurLogFile(true, false)))

}

func triggerWorkflow(cmd string, workflow *pluginmanager.Workflow) error {
	log.Println("Entering triggerWorkflow")
	defer log.Println("Exiting triggerWorkflow")

	for idx, actionRollback := range workflow.ActionRollbacks {
		pluginType := actionRollback.Action
		err := List(pluginType, map[string]string{"TYPE": "ACTION"})
		if err != nil {
			logutil.PrintNLogError("Error: %s", err.Error())
		}
		if idx > 0 {
			graph.ConnectGraph(workflow.ActionRollbacks[idx-1].Action, pluginType)
		}

		rollbackPluginType := actionRollback.Rollback
		if rollbackPluginType != "" {
			err := List(rollbackPluginType, map[string]string{"TYPE": "ROLLBACK"})
			if err != nil {
				logutil.PrintNLogError("Error: %s", err.Error())
			}
			graph.ConnectGraph(pluginType, rollbackPluginType)
			// NOTE: Some actions may not have rollback plugin-type. In those cases, instead of connecting current rollback to its immediate previous rollback plugin, connect to the next available previous rollback plugin-type.
			for rIdx := idx; rIdx > 0; rIdx-- {
				if workflow.ActionRollbacks[rIdx-1].Rollback != "" {
					graph.ConnectGraph(rollbackPluginType, workflow.ActionRollbacks[rIdx-1].Rollback)
					break
				}
			}
		}

	}

	switch cmd {
	case "list":
		// List already done above.
		logutil.PrintNLog("The list of plugins are mapped in %s\n",
			graph.GetImagePath())

	case "run":
		runRollback := false
		workflowCnt := len(workflow.ActionRollbacks)
		log.Printf("Number of actions to run: %+v\n", workflowCnt)
		workflowStatus := pluginmanager.WorkflowStatus{}
		defer output.Write(&workflowStatus)
		workflowStatus.Action = make([]*pluginmanager.PluginTypeStatus, workflowCnt)
		workflowStatus.Rollback = make([]*pluginmanager.PluginTypeStatus, workflowCnt)
		workflowStatus.Status = pluginmanager.DStatusStart
		idx := 0
		var actionRollback *pluginmanager.ActionRollback
		for idx, actionRollback = range workflow.ActionRollbacks {
			pluginType := actionRollback.Action
			fmt.Printf("\nRunning action plugins: %v [%d/%d]...\n", pluginType, idx+1, workflowCnt)
			workflowStatus.Action[idx] = &pluginmanager.PluginTypeStatus{}
			err := Run(workflowStatus.Action[idx], pluginType, map[string]string{"TYPE": "ACTION"})
			if err != nil {
				logutil.PrintNLogError("%s", err.Error())
				workflowStatus.Status = pluginmanager.DStatusFail
				workflowStatus.Action[idx].Status = pluginmanager.DStatusFail
				workflowStatus.Action[idx].StdOutErr = err.Error()
				runRollback = true
				break
			}
		}

		fmt.Println()
		if runRollback {
			logutil.PrintNLog("Starting rollback...")
			totalRollbackPluginTypes2Run := idx + 1
			log.Printf("Number of rollback plugin-types to run: %+v\n", totalRollbackPluginTypes2Run)
			for ; idx >= 0; idx-- {
				rollbackPluginType := workflow.ActionRollbacks[idx].Rollback
				fmt.Printf("\nRunning rollback plugins: %v [%d/%d]...\n", rollbackPluginType, totalRollbackPluginTypes2Run-idx, totalRollbackPluginTypes2Run)
				workflowStatus.Rollback[idx] = &pluginmanager.PluginTypeStatus{}
				err := Run(workflowStatus.Rollback[idx], rollbackPluginType, map[string]string{"TYPE": "ROLLBACK"})
				if err != nil {
					logutil.PrintNLogError("Error: %s", err.Error())
					workflowStatus.Rollback[idx].Status = pluginmanager.DStatusFail
					workflowStatus.Rollback[idx].StdOutErr = err.Error()
				}
			}
			return logutil.PrintNLogError("Running Workflow: %v",
				pluginmanager.DStatusFail)
		}

		workflowStatus.Status = pluginmanager.DStatusOk
		logutil.PrintNLog("Running Workflow: %v\n", workflowStatus.Status)
	}

	return nil
}

// ScanCommandOptions scans for the command line options and makes appropriate
// function call.
// Input:
//  1. map[string]interface{}
//     where, the options could be following:
//     "progname":  Name of the program along with any cmds (ex: asum pm)
//     "cmd-index": Index to the cmd (ex: run)
func ScanCommandOptions(options map[string]interface{}) error {
	log.Printf("Entering ScanCommandOptions(%+v)...", options)
	defer log.Println("Exiting ScanCommandOptions")

	progname := filepath.Base(os.Args[0])
	cmdIndex := 1
	if valI, ok := options["progname"]; ok {
		progname = valI.(string)
	}
	if valI, ok := options["cmd-index"]; ok {
		cmdIndex = valI.(int)
	}
	cmd := os.Args[cmdIndex]
	log.Println("progname:", progname, "cmd with arguments: ", os.Args[cmdIndex:])

	switch cmd {
	case "version":
		logutil.PrintNLog("Plugin Manager (PM) version %s\n", version)

	case "list":
		err := CmdOptions.ListCmd.Parse(os.Args[cmdIndex+1:])
		if err != nil {
			log.Fatalln(cmd, "command arguments parse error:", err.Error())
		}

	case "run":
		err := CmdOptions.RunCmd.Parse(os.Args[cmdIndex+1:])
		if err != nil {
			log.Fatalln(cmd, "command arguments parse error:", err.Error())
		}

	case "server":
		err := CmdOptions.ServerCmd.Parse(os.Args[cmdIndex+1:])
		if err != nil {
			log.Fatalln(cmd, "command arguments parse error:", err.Error())
		}

	case "help":
		subcmd := ""
		if len(os.Args) == cmdIndex+2 {
			subcmd = os.Args[cmdIndex+1]
		} else if len(os.Args) > cmdIndex+2 {
			fmt.Fprintf(os.Stderr, "usage: %s help command\n\nToo many arguments (%d) given.\n", progname, len(os.Args))
			os.Exit(2)
		}
		usage(progname, subcmd)

	default:
		fmt.Fprintf(os.Stderr, "%s: unknown command \"%s\"\n", progname, os.Args[1])
		fmt.Fprintf(os.Stderr, "Run '%s help [command]' for usage.\n", progname)
		os.Exit(2)
	}

	// Override `pm.config.yaml` value with command-line arguments.
	if *CmdOptions.libraryPtr != "" {
		config.SetPluginsLibrary(*CmdOptions.libraryPtr)
	}
	myLogFile := "./"
	if logutil.GetLogDir() != "" {
		config.SetPMLogDir(logutil.GetLogDir())
		myLogFile = config.GetPMLogDir()
	}
	// Info: Call set PM log-dir to clean extra slashes, and to append path
	// 	separator at the end.
	config.SetPMLogDir(config.GetPMLogDir())
	tLogFile := progname
	if logutil.GetLogFile() != "" {
		tLogFile = logutil.GetLogFile()
	}
	// NOTE: Even when no log file is specified, and we're using default log
	//  file name, we still need to call SetPMLogFile() as SVG image file name
	//  is based on this. Otherwise image and dot files will not have any names
	//  but only extensions (i.e., they get created as hidden files).
	config.SetPMLogFile(tLogFile)
	myLogFile += config.GetPMLogFile()
	if myLogFile != config.DefaultLogPath {
		myLogFile = filepath.Clean(myLogFile)
		log.Println("Logging to specified log file:", myLogFile)
		err := logutil.SetLogging(myLogFile)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if cmd == "server" {
		RegisterHandlers(*CmdOptions.portPtr)
	}

	log.Printf("plugin-type: %+v, workflow: %+v\n",
		*CmdOptions.pluginTypePtr, *CmdOptions.workflowPtr)
	if *CmdOptions.pluginTypePtr != "" && *CmdOptions.workflowPtr != "" {
		log.Fatalln("Only one of either 'plugin-type' or 'workflow' argument can be specified. Check usage for details...")
		CmdOptions.RunCmd.Usage()
	}

	var err error
	if *CmdOptions.workflowPtr != "" {
		var workflow pluginmanager.Workflow
		json.Unmarshal([]byte(*CmdOptions.workflowPtr), &workflow.ActionRollbacks)
		log.Printf("Received workflow request: %+v\n", &workflow)

		err = triggerWorkflow(cmd, &workflow)
	}

	if *CmdOptions.pluginTypePtr != "" {
		pluginType := *CmdOptions.pluginTypePtr
		switch cmd {
		case "list":
			err = List(pluginType, nil)
			logutil.PrintNLog("The list of plugins are mapped in %s\n",
				graph.GetImagePath())

		case "run":
			pmstatus := pluginmanager.PluginTypeStatus{}
			err = Run(&pmstatus, pluginType, nil)
			output.Write(&pmstatus)
		}
	}
	return err
}

// Usage of Plugin Manager (pm) command.
func usage(progname, subcmd string) {
	switch subcmd {
	case "", "pm":
		var usageStr = `
Plugin Manager ( PROGNAME ` + subcmd + `) is a tool for managing ASUM plugins.

Usage:

	PROGNAME ` + subcmd + ` command [arguments]

The commands are:

	server 	    Starts server runs Plugin Manager to serve API requests.
	list 		lists plugins and its dependencies of specified type in an image.
	run 		run plugins of specified type.
	version		print Plugin Manager version.

Use "PROGNAME ` + subcmd + ` help [command]" for more information about a command.
		
`
		fmt.Fprint(os.Stderr, strings.Replace(usageStr, "PROGNAME", progname, -1))
	case "version":
		CmdOptions.versionCmd.Usage()
	case "list":
		CmdOptions.ListCmd.Usage()
	case "run":
		CmdOptions.RunCmd.Usage()
	case "server":
		CmdOptions.ServerCmd.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown help topic `%s`. Run '%s'.", subcmd, progname+" help")
		fmt.Println()
		os.Exit(2)
	}
}
