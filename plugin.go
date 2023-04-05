// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package pm defines Plugin Manager (PM) functions like executing
// all plugins of a particular plugin type.
package pm

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/VeritasOS/plugin-manager/config"
	"github.com/VeritasOS/plugin-manager/graph"
	"github.com/VeritasOS/plugin-manager/pluginmanager"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
	osutils "github.com/VeritasOS/plugin-manager/utils/os"
	"github.com/VeritasOS/plugin-manager/utils/output"
)

var (
	// Version of the Plugin Manager (PM).
	version = "4.6"
)

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
	dirs, err := ioutil.ReadDir(library)
	if err != nil {
		log.Printf("ioutil.ReadDir(%s); Error: %s", library, err.Error())
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

		tfiles, err := ioutil.ReadDir(compPluginDir)
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
		if matched == true {
			pluginFiles = append(pluginFiles, file)
		}
	}
	return pluginFiles, nil
}

// getPluginType returns the plugin type of the specified plugin file.
func getPluginType(file string) string {
	return strings.Replace(path.Ext(file), ".", ``, -1)
}

func getPluginsInfo(pluginType string) (pluginmanager.Plugins, error) {
	var pluginsInfo = make(pluginmanager.Plugins)
	pluginFiles, err := getPluginFiles(pluginType)
	if err != nil {
		return pluginsInfo, err
	}
	for file := range pluginFiles {
		fContents, rerr := readFile(filepath.FromSlash(
			config.GetPluginsLibrary() + pluginFiles[file]))
		if rerr != nil {
			return pluginsInfo, logutil.PrintNLogError(rerr.Error())
		}
		// log.Printf("Plugin file %s contents: \n%s\n", pluginFiles[file], fContents)
		pInfo, perr := parseUnitFile(fContents)
		if perr != nil {
			return pluginsInfo, perr
		}
		log.Printf("Plugin %s info: %+v\n", pluginFiles[file], pInfo)
		pluginsInfo[pluginFiles[file]] = &pInfo
	}
	return pluginsInfo, nil
}

func normalizePluginsInfo(pluginsInfo pluginmanager.Plugins) pluginmanager.Plugins {
	log.Println("Entering normalizePluginsInfo")
	defer log.Println("Exiting normalizePluginsInfo")

	nPInfo := pluginmanager.Plugins{}
	for pFile, pFContents := range pluginsInfo {
		nPInfo[pFile] = &pluginmanager.PluginAttributes{
			Description: pFContents.Description,
			ExecStart:   pFContents.ExecStart,
			FileName:    pFile,
		}
		nPInfo[pFile].RequiredBy = append(nPInfo[pFile].Requires, pFContents.RequiredBy...)
		nPInfo[pFile].Requires = append(nPInfo[pFile].Requires, pFContents.Requires...)
		log.Printf("%s plugin dependencies: %v", pFile, nPInfo[pFile])
	}
	for p := range nPInfo {
		log.Printf("nPInfo key(%s): %v", p, nPInfo[p])
		for _, rs := range nPInfo[p].Requires {
			// Check whether it's already marked as RequiredBy dependency in `Requires` plugin.
			// log.Printf("Check whether `in` (%s) already marked as RequiredBy dependency in `Requires`(%s) plugin: %v",
			// p, rs, nPInfo[rs])
			present := false
			// If dependencies are missing, then nPInfo[rs] value will not be defined.
			if nPInfo[rs] != nil {
				log.Printf("PluginInfo for %s is present: %v", rs, nPInfo[rs])
				for _, rby := range nPInfo[rs].RequiredBy {
					log.Printf("p(%s) == rby(%s)? %v", p, rby, p == rby)
					if p == rby {
						present = true
						break
					}
				}
				if !present {
					nPInfo[rs].RequiredBy = append(nPInfo[rs].RequiredBy, p)
					log.Printf("Added %s as RequiredBy dependency of %s: %+v", p, rs, nPInfo[rs])
				}
			}
		}

		// Check whether RequiredBy dependencies are also marked as Requires dependency on other plugin.
		log.Printf("Check whether RequiredBy dependencies are also marked as Requires dependency on other plugin.")
		for _, rby := range nPInfo[p].RequiredBy {
			log.Printf("RequiredBy of %s: %s", p, rby)
			log.Printf("nPInfo of %s: %+v", rby, nPInfo[rby])
			// INFO: If one plugin type is added as dependent on another by
			// any chance, then skip checking its contents as the other
			// plugin type files were not parsed.
			if nPInfo[rby] == nil {
				// NOTE: Add the missing plugin in Requires, So that the issue
				// gets caught during validation.
				nPInfo[p].Requires = append(nPInfo[p].Requires, rby)
				continue
			}
			present := false
			for _, rs := range nPInfo[rby].Requires {
				if p == rs {
					present = true
					break
				}
			}
			if !present {
				nPInfo[rby].Requires = append(nPInfo[rby].Requires, p)
				log.Printf("Added %s as Requires dependency of %s: %+v", p, rby, nPInfo[rby])
			}
		}
	}
	log.Printf("Plugins info after normalizing: \n%+v\n", nPInfo)
	return nPInfo
}

// parseUnitFile parses the plugin file contents.
func parseUnitFile(fileContents string) (pluginmanager.PluginAttributes, error) {
	log.Println("Entering parseUnitFile")
	defer log.Println("Exiting parseUnitFile")

	pluginInfo := pluginmanager.PluginAttributes{}
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

func validateDependencies(nPInfo pluginmanager.Plugins) ([]string, error) {
	log.Println("Entering validateDependencies")
	defer log.Println("Exiting validateDependencies")

	var pluginOrder []string
	notPlacedPlugins := []string{}
	dependencyMet := map[string]bool{}
	// for pFile, pFContents := range pluginsInfo {}
	sortedPFiles := []string{}
	for pFile := range nPInfo {
		sortedPFiles = append(sortedPFiles, pFile)
	}
	// NOTE: Sorting plugin files mainly to have a deterministic order,
	// though it's not required for solution to work.
	// (Sorting takes care of unit tests as maps return keys/values in random order).
	sort.Strings(sortedPFiles)
	log.Printf("Plugin files in sorted order: %+v\n", sortedPFiles)

	for pFileIndex := range sortedPFiles {
		pFile := sortedPFiles[pFileIndex]
		pFContents := nPInfo[pFile]
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
		pDependencies := nPInfo[pFile].Requires
		log.Printf("Plugin %s dependencies: %+v\n", pFile, pDependencies)

		dependencyMet[pFile] = true
		for w := range pDependencies {
			val := dependencyMet[pDependencies[w]]
			if false == val {
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

func executePluginCmd(statusCh chan<- map[string]*pluginmanager.RunStatus, p string, pluginsInfo pluginmanager.Plugins, failedDependency bool) {
	pInfo := pluginsInfo[p]
	log.Printf("\nChannel: Plugin %s info: \n%+v\n", p, pInfo)
	graph.UpdateGraph(getPluginType(p), p, pluginmanager.DStatusStart, "")
	logutil.PrintNLog("\n%s: %s\n", pInfo.Description, pluginmanager.DStatusStart)
	// Get relative path to plugins log file from PM log dir, so that linking
	// in plugin graph works even when the logs are copied to another system.
	pluginLogFile := strings.Replace(config.GetPluginsLogDir(),
		config.GetPMLogDir(), "", -1) +
		strings.Replace(p, string(os.PathSeparator), ":", -1) +
		"." + time.Now().Format(time.RFC3339Nano) + ".log"
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
		graph.UpdateGraph(getPluginType(p), p, myStatus, "")
		logutil.PrintNLog("%s: %s\n", pInfo.Description, myStatus)
		statusCh <- map[string]*pluginmanager.RunStatus{p: {Status: myStatus}}
		return
	}

	log.Printf("\nExecuting command: %s\n", pInfo.ExecStart)
	cmdParam := strings.Split(pInfo.ExecStart, " ")
	cmdStr := cmdParam[0]
	cmdParams := os.ExpandEnv(strings.Join(cmdParam[1:], " "))
	cmdParamsExpanded := strings.Split(cmdParams, " ")

	cmd := exec.Command(os.ExpandEnv(cmdStr), cmdParamsExpanded...)
	stdOutErr, err := cmd.CombinedOutput()

	func() {
		chLog.Println("Executing command:", pInfo.ExecStart)
		if err != nil {
			chLog.Println("Error:", err.Error())
			graph.UpdateGraph(getPluginType(p), p, pluginmanager.DStatusFail, pluginLogFile)
		} else {
			chLog.Println("Stdout & Stderr:", string(stdOutErr))
			graph.UpdateGraph(getPluginType(p), p, pluginmanager.DStatusOk, pluginLogFile)
		}
	}()

	log.Println("Stdout & Stderr:", string(stdOutErr))
	pStatus := pluginmanager.RunStatus{StdOutErr: string(stdOutErr)}
	if err != nil {
		pStatus.Status = pluginmanager.DStatusFail
		log.Printf("Failed to execute plugin %s. Error: %s\n", p, err.Error())
		logutil.PrintNLog("%s: %s\n", pInfo.Description, pluginmanager.DStatusFail)
		statusCh <- map[string]*pluginmanager.RunStatus{p: &pStatus}
		return
	}
	pStatus.Status = pluginmanager.DStatusOk
	logutil.PrintNLog("%s: %s\n", pInfo.Description, pluginmanager.DStatusOk)
	statusCh <- map[string]*pluginmanager.RunStatus{p: &pStatus}
}

func executePlugins(psStatus *pluginmanager.PluginsStatus, nPInfo pluginmanager.Plugins, sequential bool) bool {
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
	for p := range nPInfo {
		waitCount[p] = len(nPInfo[p].Requires)
		log.Printf("%s plugin dependencies: %+v", p, nPInfo[p])
	}

	executingCnt := 0
	exeCh := make(chan map[string]*pluginmanager.RunStatus)
	pluginIndexes := make(map[string]int)
	failedDependency := make(map[string]bool)
	for len(nPInfo) > 0 || executingCnt != 0 {
		for p := range nPInfo {
			// INFO: When all dependencies are met, plugin waitCount would be 0.
			// 	When sequential execution is enforced, even if a plugin is ready
			// 	 to run, make sure that only one plugin is running at time, by
			// 	 checking executing count is 0.
			// 	When sequential execution is not enforced, run plugins that are ready.
			if waitCount[p] == 0 && ((sequential == false) ||
				(sequential == true && executingCnt == 0)) {
				log.Printf("Plugin %s is ready for execution: %v.", p, nPInfo[p])
				waitCount[p]--

				ps := pluginmanager.RunStatus{}
				ps.PluginAttributes = *nPInfo[p]
				*psStatus = append(*psStatus, ps)
				pluginIndexes[p] = len(*psStatus) - 1

				go executePluginCmd(exeCh, p, nPInfo, failedDependency[p])
				executingCnt++
			}
		}
		// start other dependent ones as soon as one of the plugin completes.
		exeStatus := <-exeCh
		executingCnt--
		for plugin, pStatus := range exeStatus {
			log.Printf("%s status: %v", plugin, pStatus.Status)
			pIdx := pluginIndexes[plugin]
			ps := *psStatus
			ps[pIdx].Status = pStatus.Status
			ps[pIdx].StdOutErr = pStatus.StdOutErr
			if pStatus.Status == pluginmanager.DStatusFail {
				retStatus = false
			}

			for _, rby := range nPInfo[plugin].RequiredBy {
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
			delete(nPInfo, plugin)
		}
	}
	return retStatus
}

// CmdOptions contains subcommands and parameters of the pm command.
var CmdOptions struct {
	RunCmd     *flag.FlagSet
	ListCmd    *flag.FlagSet
	versionCmd *flag.FlagSet
	versionPtr *bool

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

	// logDirPtr indicates the location for writing log file.
	logDirPtr *string

	// logFilePtr indicates the log file name to write to in the logDirPtr location.
	logFilePtr *string
}

// List the plugin and its dependencies.
func List(pluginType string) error {
	var pluginsInfo, err = getPluginsInfo(pluginType)
	if err != nil {
		return err
	}
	nPInfo := normalizePluginsInfo(pluginsInfo)

	err = graph.InitGraph(pluginType, nPInfo)
	if err != nil {
		return err
	}

	logutil.PrintNLog("The list of plugins are mapped in %s\n",
		graph.GetImagePath())
	return nil
}

func readFile(filePath string) (string, error) {
	bFileContents, err := ioutil.ReadFile(filePath)
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
	CmdOptions.logDirPtr = CmdOptions.RunCmd.String(
		"log-dir",
		"",
		"Directory for the log file.",
	)
	CmdOptions.logFilePtr = CmdOptions.RunCmd.String(
		"log-file",
		"",
		"Name of the log file.",
	)
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
		CmdOptions.logDirPtr,
		"log-dir",
		"",
		"Directory for the log file.",
	)
	CmdOptions.ListCmd.StringVar(
		CmdOptions.logFilePtr,
		"log-file",
		"",
		"Name of the log file.",
	)
}

// Run the specified plugin type plugins.
func Run(result *pluginmanager.RunAllStatus, pluginType string) error {
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
	graph.InitGraph(pluginType, nPInfo)

	status = executePlugins(&result.Plugins, nPInfo, *CmdOptions.sequential)
	if status != true {
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
	if *CmdOptions.logDirPtr != "" {
		config.SetPMLogDir(*CmdOptions.logDirPtr)
		myLogFile = config.GetPMLogDir()
	}
	// Info: Call set PM log-dir to clean extra slashes, and to append path
	// 	separator at the end.
	config.SetPMLogDir(config.GetPMLogDir())
	if *CmdOptions.logFilePtr != "" {
		config.SetPMLogFile(*CmdOptions.logFilePtr)
		myLogFile += config.GetPMLogFile()
	} else {
		myLogFile += progname
	}
	if myLogFile != config.DefaultLogPath {
		myLogFile = filepath.Clean(myLogFile)
		log.Println("Logging to specified log file:", myLogFile)
		logutil.SetLogging(myLogFile)
	}

	if *CmdOptions.pluginTypePtr != "" {
		pluginType := *CmdOptions.pluginTypePtr
		var err error
		switch cmd {
		case "list":
			err = List(pluginType)

		case "run":
			pmstatus := pluginmanager.RunAllStatus{}
			err = Run(&pmstatus, pluginType)
			output.Write(pmstatus)
		}
		if err != nil {
			return err
		}
	}
	return nil
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

	list 		lists plugins and its dependencies of specified type in an image.
	run 		run plugins of specified type.
	version		print Plugin Manager version.

Use "PROGNAME ` + subcmd + ` help [command]" for more information about a command.
		
`
		fmt.Fprintf(os.Stderr, strings.Replace(usageStr, "PROGNAME", progname, -1))
	case "version":
		CmdOptions.versionCmd.Usage()
	case "list":
		CmdOptions.ListCmd.Usage()
	case "run":
		CmdOptions.RunCmd.Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown help topic `%s`. Run '%s'.", subcmd, progname+" help")
		fmt.Println()
		os.Exit(2)
	}
}
