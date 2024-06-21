// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package pm defines Plugin Manager (PM) functions like executing
// all plugins of a particular plugin type.
package pm

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/VeritasOS/plugin-manager/config"
	logger "github.com/VeritasOS/plugin-manager/utils/log"
	osutils "github.com/VeritasOS/plugin-manager/utils/os"
	"github.com/VeritasOS/plugin-manager/utils/output"
	"gopkg.in/yaml.v3"
)

var (
	// Version of the Plugin Manager (PM).
	version = "4.8"
)

// Status of plugin execution used for displaying to user on console.
const (
	dStatusFail  = "Failed"
	dStatusOk    = "Succeeded"
	dStatusSkip  = "Skipped"
	dStatusStart = "Starting"
)

// PluginAttributes that are supported in a plugin file.
type PluginAttributes struct {
	Name        string
	Description string
	ExecStart   string
	RequiredBy  []string
	Requires    []string
}

// pluginsMap is basically a map of file and its contents.
type pluginsMap map[string]*PluginAttributes

// Plugin is the plugin's info: attributes, status, stdouterr.
type Plugin struct {
	PluginAttributes `yaml:",inline"`
	Status           string
	StdOutErr        string
}

// Plugins is a list of plugins' info.
type Plugins []Plugin

// RunStatus is the pm run status.
type RunStatus struct {
	Type string
	// TODO: Add Percentage to get no. of pending vs. completed run of plugins.
	Plugins   Plugins `yaml:",omitempty"`
	Status    string
	StdOutErr string
}

// getPluginFiles retrieves the plugin files under each component matching
// the specified pluginType.
func getPluginFiles(pluginType, library string) ([]string, error) {
	logger.Debug.Println("Entering getPluginFiles")
	defer logger.Debug.Println("Exiting getPluginFiles")

	var pluginFiles []string
	if _, err := os.Stat(library); os.IsNotExist(err) {
		return pluginFiles, logger.ConsoleError.PrintNReturnError("Library '%s' doesn't exist. "+
			"A valid plugins library path must be specified.", library)
	}
	var files []string
	dirs, err := ioutil.ReadDir(library)
	if err != nil {
		logger.Error.Printf("Failed to call ioutil.ReadDir(%s), err=%s", library, err.Error())
		return pluginFiles, logger.ConsoleError.PrintNReturnError("Failed to get contents of %s plugins library.", library)
	}

	for _, dir := range dirs {
		compPluginDir := filepath.FromSlash(library + "/" + dir.Name())
		fi, err := os.Stat(compPluginDir)
		if err != nil {
			logger.Error.Printf("Unable to stat on %s directory, err=%s", dir, err.Error())
			continue
		}
		if !fi.IsDir() {
			logger.Error.Printf("%s is not a directory.", compPluginDir)
			continue
		}

		tfiles, err := ioutil.ReadDir(compPluginDir)
		if err != nil {
			logger.Error.Printf("Unable to read contents of %s directory, err=%s", compPluginDir, err.Error())
		}
		for _, tf := range tfiles {
			files = append(files, filepath.FromSlash(dir.Name()+"/"+tf.Name()))
		}
	}

	for _, file := range files {
		matched, err := regexp.MatchString("[.]"+pluginType+"$", file)
		if err != nil {
			logger.Error.Printf("Failed to call regexp.MatchString(%s, %s), err=%s", "[.]"+pluginType, file, err.Error())
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

func getPluginsInfoFromJSONStrOrFile(strOrFile string) (RunStatus, error) {
	var err error
	var pluginsInfo RunStatus
	rawData := strOrFile
	jsonFormat := true

	// If Plugins information is in file...
	fi, err := os.Stat(strOrFile)
	if err != nil {
		logger.Debug.Printf("Specified input is not a file. Err: %s",
			err.Error())
	} else {
		if fi.IsDir() {
			return pluginsInfo,
				logger.ConsoleError.PrintNReturnError(
					"Specified path %s is directory. Plugins info should be specified either as a json string or in a json file.",
					strOrFile)
		}

		pluginsFile := strOrFile
		fh, err := os.Open(pluginsFile)
		if err != nil {
			logger.ConsoleError.PrintNReturnError("%s", err)
			return pluginsInfo, err
		}
		defer fh.Close()

		rawData, err = readFile(filepath.FromSlash(pluginsFile))
		if err != nil {
			return pluginsInfo,
				logger.ConsoleError.PrintNReturnError(err.Error())
		}

		logger.Debug.Printf("Plugins file %v has ext %v", pluginsFile, path.Ext(pluginsFile))
		if path.Ext(pluginsFile) == ".yaml" || path.Ext(pluginsFile) == ".yml" {
			jsonFormat = false
		}
	}
	// INFO: Use RunStatus to unmarshal to keep input consistent with current
	//  output json, so that rerun failed could be done using result json.
	var pluginsData RunStatus
	if jsonFormat {
		err = json.Unmarshal([]byte(rawData), &pluginsData)
	} else {
		err = yaml.Unmarshal([]byte(rawData), &pluginsData)
	}
	if err != nil {
		logger.Error.Printf("Failed to call Unmarshal(%s, %v); err=%#v",
			rawData, &pluginsInfo, err)
		return pluginsInfo,
			logger.ConsoleError.PrintNReturnError(
				"Plugins is not in expected format. Error: %s", err.Error())
	}
	return pluginsData, nil
}

func getPluginsInfoFromLibrary(pluginType, library string) (Plugins, error) {
	var pluginsInfo Plugins
	pluginFiles, err := getPluginFiles(pluginType, library)
	if err != nil {
		return pluginsInfo, err
	}
	for file := range pluginFiles {
		fContents, rerr := readFile(filepath.FromSlash(
			library + pluginFiles[file]))
		if rerr != nil {
			return pluginsInfo, logger.ConsoleError.PrintNReturnError(rerr.Error())
		}
		logger.Debug.Printf("Plugin file %s contents: \n%s\n",
			pluginFiles[file], fContents)
		pInfo, perr := parseUnitFile(fContents)
		if perr != nil {
			return pluginsInfo, perr
		}
		logger.Info.Printf("Plugin %s info: %+v", pluginFiles[file], pInfo)
		pInfo.Name = pluginFiles[file]
		pluginsInfo = append(pluginsInfo, Plugin{PluginAttributes: pInfo})
	}
	return pluginsInfo, nil
}

func normalizePluginsInfo(pluginsInfo Plugins) pluginsMap {
	logger.Debug.Printf("Entering normalizePluginsInfo(%+v)...", pluginsInfo)
	defer logger.Debug.Println("Exiting normalizePluginsInfo")

	nPInfo := pluginsMap{}
	for _, pContents := range pluginsInfo {
		pName := pContents.PluginAttributes.Name
		nPInfo[pName] = &PluginAttributes{
			Name:        pName,
			Description: pContents.Description,
			ExecStart:   pContents.ExecStart,
		}
		nPInfo[pName].RequiredBy = append(nPInfo[pName].Requires, pContents.RequiredBy...)
		nPInfo[pName].Requires = append(nPInfo[pName].Requires, pContents.Requires...)
		logger.Info.Printf("%s plugin dependencies: %v", pName, nPInfo[pName])
	}
	for p := range nPInfo {
		logger.Debug.Printf("nPInfo key(%s): %v", p, nPInfo[p])
		for _, rs := range nPInfo[p].Requires {
			// Check whether it's already marked as RequiredBy dependency in `Requires` plugin.
			// logger.Info.Printf("Check whether `in` (%s) already marked as RequiredBy dependency in `Requires`(%s) plugin: %v",
			// p, rs, nPInfo[rs])
			present := false
			// If dependencies are missing, then nPInfo[rs] value will not be defined.
			if nPInfo[rs] != nil {
				logger.Info.Printf("PluginInfo for %s is present: %v", rs, nPInfo[rs])
				for _, rby := range nPInfo[rs].RequiredBy {
					logger.Debug.Printf("p(%s) == rby(%s)? %v", p, rby, p == rby)
					if p == rby {
						present = true
						break
					}
				}
				if !present {
					nPInfo[rs].RequiredBy = append(nPInfo[rs].RequiredBy, p)
					logger.Info.Printf("Added %s as RequiredBy dependency of %s: %+v", p, rs, nPInfo[rs])
				}
			}
		}

		// Check whether RequiredBy dependencies are also marked as Requires dependency on other plugin.
		logger.Info.Println("Check whether RequiredBy dependencies are also marked as Requires dependency on other plugin.")
		for _, rby := range nPInfo[p].RequiredBy {
			logger.Debug.Printf("RequiredBy of %s: %s", p, rby)
			logger.Debug.Printf("nPInfo of %s: %+v", rby, nPInfo[rby])
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
				logger.Info.Printf("Added %s as Requires dependency of %s: %+v", p, rby, nPInfo[rby])
			}
		}
	}
	logger.Info.Printf("Plugins info after normalizing: \n%+v\n", nPInfo)
	return nPInfo
}

// parseUnitFile parses the plugin file contents.
func parseUnitFile(fileContents string) (PluginAttributes, error) {
	logger.Debug.Println("Entering parseUnitFile")
	defer logger.Debug.Println("Exiting parseUnitFile")

	pluginInfo := PluginAttributes{}
	if len(fileContents) == 0 {
		return pluginInfo, nil
	}
	lines := strings.Split(fileContents, "\n")
	for l := range lines {
		line := strings.TrimSpace(lines[l])
		logger.Debug.Println("line...", line)
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			// No need to parse comments.
			logger.Debug.Println("Skipping comment line...", line)
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
			logger.Debug.Printf("Non-standard line found: %s", line)
			break
		}
	}

	return pluginInfo, nil
}

func validateDependencies(nPInfo pluginsMap) ([]string, error) {
	logger.Debug.Println("Entering validateDependencies")
	defer logger.Debug.Println("Exiting validateDependencies")

	var pluginOrder []string
	notPlacedPlugins := []string{}
	dependencyMet := map[string]bool{}
	// for pName, pContents := range pluginsInfo {}
	sortedpNames := []string{}
	for pName := range nPInfo {
		sortedpNames = append(sortedpNames, pName)
	}
	// NOTE: Sorting plugin files mainly to have a deterministic order,
	// though it's not required for solution to work.
	// (Sorting takes care of unit tests as maps return keys/values in random order).
	sort.Strings(sortedpNames)
	logger.Info.Printf("Plugin files in sorted order: %+v", sortedpNames)

	for pNameIndex := range sortedpNames {
		pName := sortedpNames[pNameIndex]
		pContents := nPInfo[pName]
		logger.Debug.Printf("\nFile: %s \n%+v \n\n", pName, pContents)
		if len(pContents.Requires) == 0 {
			dependencyMet[pName] = true
			pluginOrder = append(pluginOrder, pName)
		} else {
			dependencyMet[pName] = false
			notPlacedPlugins = append(notPlacedPlugins, pName)
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
		pName := notPlacedPlugins[0]
		notPlacedPlugins = notPlacedPlugins[1:]
		pDependencies := nPInfo[pName].Requires
		logger.Info.Printf("Plugin %s dependencies: %+v", pName, pDependencies)

		dependencyMet[pName] = true
		for w := range pDependencies {
			val := dependencyMet[pDependencies[w]]
			if false == val {
				// If dependency met is false, then process it later again after all dependencies are met.
				dependencyMet[pName] = false
				logger.Warning.Printf("Adding %s back to list %s to process as %s plugin dependency is not met.",
					pName, notPlacedPlugins, pDependencies[w])
				notPlacedPlugins = append(notPlacedPlugins, pName)
				break
			}
		}
		// If dependency met is not set to false, then it means all
		// dependencies are met. So, add it to pluginOrder
		if false != dependencyMet[pName] {
			logger.Info.Printf("Dependency met for %s: %v.", pName, dependencyMet[pName])
			pluginOrder = append(pluginOrder, pName)
		}

		elementsLeft--
		if elementsLeft == 0 {
			logger.Debug.Printf("PrevLen: %d; CurLen: %d.", prevLen, curLen)
			curLen = len(notPlacedPlugins)
			if prevLen == curLen {
				// INFO: Clear out the pluginOrder as we cannot run all the
				// 	plugins either due to missing dependencies or having
				// 	circular dependency.
				return []string{}, logger.ConsoleError.PrintNReturnError(
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

func executePluginCmd(statusCh chan<- map[string]*Plugin, p string, pInfo PluginAttributes, failedDependency bool, env map[string]string) {
	logger.Debug.Printf("Channel: Plugin %s info: \n%+v", p, pInfo)
	updateGraph(getPluginType(p), p, dStatusStart, "")
	logger.ConsoleInfo.Printf("%s: %s", pInfo.Description, dStatusStart)
	// Create chLog, by default it will use syslog, if user specified logFile, then use previous defined log generator
	var chLog *log.Logger
	if !logger.IsFileLogger() {
		var logTag string
		// Set log tag for
		logTag = logger.SyslogTagPrefix + "pm-" + *CmdOptions.logTagPtr
		logger.Debug.Printf("logTag = %s", logTag)
		syslogHandle, err := syslog.New(syslog.LOG_LOCAL0|syslog.LOG_INFO, logTag)
		if err != nil {
			logger.Error.Printf("Failed to call syslog.New, err=%s", err.Error())
		}
		defer syslogHandle.Close()
		chLog = log.New(syslogHandle, "", 0)
	} else {
		// Get relative path to plugins log file from PM log dir, so that linking
		// in plugin graph works even when the logs are copied to another system.
		pluginLogFile := strings.Replace(config.GetPluginsLogDir(), config.GetPMLogDir(), "", -1) +
			strings.Replace(p, string(os.PathSeparator), ":", -1) +
			"." + time.Now().Format(time.RFC3339Nano) + ".log"
		logFile := config.GetPMLogDir() + pluginLogFile
		fh, openerr := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if openerr != nil {
			logger.Error.Printf("Failed to call os.OpenFile(%s), err=%s", logFile, openerr.Error())
			// Ignore error and continue as plugin log file creation is not fatal.
		}
		defer fh.Close()
		// chLog is a channel logger
		chLog = log.New(fh, "", log.LstdFlags)
		chLog.SetOutput(fh)
	}

	chLog.Println("INFO: Plugin file:", p)

	// If already marked as failed/skipped due to dependency fail,
	// then just return that status.
	myStatus := ""
	myStatusMsg := ""
	if failedDependency {
		myStatusMsg = "Skipping as its dependency failed."
		myStatus = dStatusSkip
	} else if pInfo.ExecStart == "" {
		myStatusMsg = "Passing as ExecStart value is empty!"
		myStatus = dStatusOk
	}

	if myStatus != "" {
		chLog.Println("INFO: ", myStatusMsg)
		logger.Info.Printf("Plugin(%s): %s", p, myStatusMsg)
		updateGraph(getPluginType(p), p, myStatus, "")
		logger.ConsoleInfo.Printf("%s: %s", pInfo.Description, myStatus)
		statusCh <- map[string]*Plugin{p: {Status: myStatus}}
		return
	}

	// INFO: First initialize with existing OS env, and then overwrite any
	// 	existing keys with user specified values. I.e., Even if PM_LIBRARY
	//  env is set in shell, it'll be overwritten by Library parameter passed
	//  by user.
	envList := osutils.OsEnviron()
	envMap := osutils.EnvMap()
	for envKey, envValue := range env {
		envList = append(envList, envKey+"="+envValue)
		envMap[envKey] = envValue
	}

	getEnvVal := func(name string) string {
		// logger.Debug.Printf("In getEnvVal(%v)...", name)
		// logger.Debug.Printf("env: %+v", envMap)
		if val, ok := envMap[name]; ok {
			// logger.Debug.Printf("Key:%v = Value:%v", name, val)
			return val
		}
		return ""
	}

	logger.Info.Printf("Executing command, cmd=%s", pInfo.ExecStart)
	// INFO: Expand environment values like "PM_LIBRARY" so that ...
	// 	1. Binaries or scripts placed in the same directory as that of plugins
	// 		can be accessed as ${PM_LIBRARY}/<binary|script> path.
	// 	2. Envs that are set by caller of calling plugin manager gets expanded.
	cmdParam := strings.Split(os.Expand(pInfo.ExecStart, getEnvVal), " ")
	cmdStr := cmdParam[0]
	cmdParams := cmdParam[1:]
	cmd := exec.Command(cmdStr, cmdParams...)
	cmd.Env = envList
	stdOutErr, err := cmd.CombinedOutput()

	func() {
		chLog.Printf("INFO: Plugin(%s): Executing command: %s", p, pInfo.ExecStart)
		if err != nil {
			chLog.Printf("ERROR: Plugin(%s): Failed to execute command, err=%s", p, err.Error())
			updateGraph(getPluginType(p), p, dStatusFail, "")
		} else {
			chLog.Printf("INFO: Plugin(%s): Stdout & Stderr: %s", p, string(stdOutErr))
			updateGraph(getPluginType(p), p, dStatusOk, "")
		}
	}()

	logger.Debug.Println("Stdout & Stderr:", string(stdOutErr))
	pStatus := Plugin{StdOutErr: string(stdOutErr)}
	if err != nil {
		pStatus.Status = dStatusFail
		logger.Error.Printf("Failed to execute plugin %s. err=%s\n", p, err.Error())
		logger.ConsoleError.Printf("%s: %s\n", pInfo.Description, dStatusFail)
		statusCh <- map[string]*Plugin{p: &pStatus}
		return
	}
	pStatus.Status = dStatusOk
	logger.ConsoleInfo.Printf("%s: %s\n", pInfo.Description, dStatusOk)
	statusCh <- map[string]*Plugin{p: &pStatus}
}

func executePlugins(psStatus *Plugins, sequential bool, env map[string]string) bool {
	logger.Debug.Printf("Entering executePlugins(%+v, %v, %+v)...",
		psStatus, sequential, env)
	defer logger.Debug.Println("Exiting executePlugins")

	retStatus := true

	nPInfo := normalizePluginsInfo(*psStatus)

	_, err := validateDependencies(nPInfo)
	if err != nil {
		return false
	}

	waitCount := map[string]int{}
	for p := range nPInfo {
		waitCount[p] = len(nPInfo[p].Requires)
		logger.Info.Printf("%s plugin dependencies: %+v", p, nPInfo[p])
	}

	pluginIndexes := make(map[string]int)
	for pIdx, pInfo := range *psStatus {
		pluginIndexes[pInfo.Name] = pIdx
	}
	executingCnt := 0
	exeCh := make(chan map[string]*Plugin)
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
				logger.Info.Printf("Plugin %s is ready for execution: %v.", p, nPInfo[p])
				waitCount[p]--

				go executePluginCmd(exeCh, p, *nPInfo[p], failedDependency[p], env)
				executingCnt++
			}
		}
		// start other dependent ones as soon as one of the plugin completes.
		exeStatus := <-exeCh
		executingCnt--
		for plugin, pStatus := range exeStatus {
			logger.Info.Printf("%s status: %v", plugin, pStatus.Status)
			pIdx := pluginIndexes[plugin]
			ps := *psStatus
			ps[pIdx].Status = pStatus.Status
			ps[pIdx].StdOutErr = pStatus.StdOutErr
			if pStatus.Status == dStatusFail {
				retStatus = false
			}

			for _, rby := range nPInfo[plugin].RequiredBy {
				if pStatus.Status == dStatusFail ||
					pStatus.Status == dStatusSkip {
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

	// pluginsPtr specifies plugins Name and its Description, ExecStart and any dependencies (Requires, RequiredBy).
	// For input format, check 'Plugins' struct.
	pluginsPtr *string

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

	// logTagPtr indicates the log tag to write into syslog.
	logTagPtr *string
}

// ListOptions are optional parameters related to list function.
type ListOptions struct {
	Type string
}

// RunOptions are optional parameters related to run function.
type RunOptions struct {
	Library    string
	Type       string
	Sequential bool
}

// ListFromLibrary lists the plugin and its dependencies from the plugins
// library path.
func ListFromLibrary(pluginType, library string) error {
	pluginsInfo, err := getPluginsInfoFromLibrary(pluginType, library)
	if err != nil {
		return err
	}

	listOptions := ListOptions{
		Type: pluginType,
	}
	return list(pluginsInfo, listOptions)
}

// ListFromJSONStrOrFile lists the plugin and its dependencies from a json
// string or a json file.
func ListFromJSONStrOrFile(jsonStrOrFile string, listOptions ListOptions) error {
	pluginsInfo, err := getPluginsInfoFromJSONStrOrFile(jsonStrOrFile)
	if err != nil {
		return err
	}

	return list(pluginsInfo.Plugins, listOptions)
}

// List the plugin and its dependencies.
func list(pluginsInfo Plugins, listOptions ListOptions) error {
	pluginType := listOptions.Type

	var err error

	err = initGraph(pluginType, pluginsInfo)
	if err != nil {
		return err
	}

	logger.ConsoleInfo.Printf("The list of plugins are mapped in %s", getImagePath())
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
	logger.Debug.Println("Entering RegisterCommandOptions")
	defer logger.Debug.Println("Exiting RegisterCommandOptions")

	CmdOptions.versionCmd = flag.NewFlagSet(progname+" version", flag.ContinueOnError)
	CmdOptions.versionPtr = CmdOptions.versionCmd.Bool("version", false, "print Plugin Manager (PM) version.")

	CmdOptions.RunCmd = flag.NewFlagSet(progname+" run", flag.PanicOnError)
	CmdOptions.pluginsPtr = CmdOptions.RunCmd.String(
		"plugins",
		"",
		"Plugins and its dependencies in json format as a string or in a file (Ex: './plugins.json').\nWhen specified, plugin files are not looked up in specified -library path.",
	)
	CmdOptions.pluginTypePtr = CmdOptions.RunCmd.String(
		"type",
		"",
		"Type of plugin.",
	)
	CmdOptions.libraryPtr = CmdOptions.RunCmd.String(
		"library",
		"",
		"Path of the plugins library.\nSets PM_LIBRARY env value.\n"+
			"When '-plugins' is specified, only PM_LIBRARY env value is set. "+
			"The plugin files are not read from library path.",
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
		progname+".log",
		"Name of the log file.",
	)
	CmdOptions.logTagPtr = CmdOptions.RunCmd.String(
		"log-tag",
		"",
		"Syslog tag for plugin manager logs.",
	)
	output.RegisterCommandOptions(CmdOptions.RunCmd, map[string]string{})

	CmdOptions.ListCmd = flag.NewFlagSet(progname+" list", flag.PanicOnError)
	CmdOptions.ListCmd.StringVar(
		CmdOptions.pluginsPtr,
		"plugins",
		"",
		"Plugins and its dependencies in json format as a string or in a file (Ex: './plugins.json')",
	)
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
	CmdOptions.ListCmd.StringVar(
		CmdOptions.logTagPtr,
		"log-tag",
		"",
		"Syslog tag name.",
	)
}

// RunFromJSONStrOrFile runs the plugins based on dependencies specified in a
// json string or a json/yaml file.
func RunFromJSONStrOrFile(result *RunStatus, jsonStrOrFile string, runOptions RunOptions) error {
	pluginsInfo, err := getPluginsInfoFromJSONStrOrFile(jsonStrOrFile)
	if err != nil {
		result.Status = dStatusFail
		result.StdOutErr = err.Error()
		return err
	}
	result.Plugins = pluginsInfo.Plugins
	result.Type = pluginsInfo.Type

	return run(result, runOptions)
}

// RunFromLibrary runs the specified plugin type plugins from the library.
func RunFromLibrary(result *RunStatus, pluginType string, runOptions RunOptions) error {
	result.Type = pluginType

	var pluginsInfo, err = getPluginsInfoFromLibrary(pluginType, runOptions.Library)
	if err != nil {
		result.Status = dStatusFail
		result.StdOutErr = err.Error()
		return err
	}
	result.Plugins = pluginsInfo

	runOptions.Type = pluginType
	return run(result, runOptions)
}

// run the specified plugins.
func run(result *RunStatus, runOptions RunOptions) error {
	logger.Debug.Printf("Entering run(%+v, %+v)...", result, runOptions)
	defer logger.Debug.Println("Exiting run")
	pluginType := runOptions.Type
	sequential := runOptions.Sequential

	if err := osutils.OsMkdirAll(config.GetPluginsLogDir(), 0755); nil != err {
		err = logger.ConsoleError.PrintNReturnError(
			"Failed to create the plugins logs directory: %s. "+
				"Error: %s", config.GetPluginsLogDir(), err.Error())
		result.Status = dStatusFail
		result.StdOutErr = err.Error()
		return err
	}

	initGraph(pluginType, result.Plugins)

	env := map[string]string{}
	if runOptions.Library != "" {
		env["PM_LIBRARY"] = runOptions.Library
	}
	status := executePlugins(&result.Plugins, sequential, env)
	if status != true {
		result.Status = dStatusFail
		err := fmt.Errorf("Running %s plugins: %s", pluginType, dStatusFail)
		result.StdOutErr = err.Error()
		logger.ConsoleError.Printf("%s\n", err.Error())
		return err
	}
	result.Status = dStatusOk
	logger.ConsoleInfo.Printf("Running %s plugins: %s\n", pluginType, dStatusOk)
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
	logger.Debug.Printf("Entering ScanCommandOptions(%+v)...", options)
	defer logger.Debug.Println("Exiting ScanCommandOptions")

	progname := filepath.Base(os.Args[0])
	cmdIndex := 1
	if valI, ok := options["progname"]; ok {
		progname = valI.(string)
	}
	if valI, ok := options["cmd-index"]; ok {
		cmdIndex = valI.(int)
	}
	cmd := os.Args[cmdIndex]
	logger.Debug.Println("progname:", progname, "cmd with arguments: ", os.Args[cmdIndex:])

	switch cmd {
	case "version":
		logger.ConsoleInfo.Printf("Plugin Manager (PM) version %s", version)

	case "list":
		err := CmdOptions.ListCmd.Parse(os.Args[cmdIndex+1:])
		if err != nil {
			logger.Error.Printf("Command arguments parse error, cmd=%s, err=%s", cmd, err.Error())
		}

	case "run":
		err := CmdOptions.RunCmd.Parse(os.Args[cmdIndex+1:])
		if err != nil {
			logger.Error.Printf("Command arguments parse error, cmd=%s, err=%s", cmd, err.Error())
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
	logToNewFile := false
	if *CmdOptions.logDirPtr != "" {
		config.SetPMLogDir(*CmdOptions.logDirPtr)
		logToNewFile = true
	}
	// Info: Call set PM log-dir to clean extra slashes, and to append path
	// 	separator at the end.
	config.SetPMLogDir(config.GetPMLogDir())

	// Reinit logging if required.
	if *CmdOptions.logTagPtr != "" {
		logTag := *CmdOptions.logTagPtr
		// Use Syslog whenever logTagPtr is specified.
		err := logger.InitSysLogger(logTag, "INFO")
		if err != nil {
			fmt.Printf("Failed to initialize SysLog [%v]. Exiting...\n", err)
			os.Exit(-1)
		}
	} else {
		if *CmdOptions.logFilePtr != "" {
			config.SetPMLogFile(*CmdOptions.logFilePtr)
			logToNewFile = true
		}
		if logToNewFile {
			myLogFile := config.GetPMLogDir() + config.GetPMLogFile()
			logger.Info.Println("Logging to specified log file:", myLogFile)
			errList := logger.DeInitLogger()
			if len(errList) > 0 {
				fmt.Printf("Failed to deinitialize logger, err=[%v]", errList)
				os.Exit(-1)
			}
			err := logger.InitFileLogger(myLogFile, "INFO")
			if err != nil {
				fmt.Printf("Failed to initialize logger, err=[%v]", err)
				os.Exit(-1)
			}
		}
	}

	var err error
	pluginType := *CmdOptions.pluginTypePtr
	if *CmdOptions.pluginsPtr != "" {
		jsonStrOrFile := *CmdOptions.pluginsPtr
		switch cmd {
		case "list":
			err = ListFromJSONStrOrFile(jsonStrOrFile,
				ListOptions{Type: pluginType})

		case "run":
			pmstatus := RunStatus{}
			runOptions := RunOptions{
				Type:       pluginType,
				Sequential: *CmdOptions.sequential,
			}
			// NOTE: When '-plugins' info is passed as str or file, don't use
			// 	Library from config.
			// 	The config file is expected to have some library path, and that
			//  may not be applicable for this set of inputs. So, set/use
			//  "Library" value only if it's passed as cmdline argument.
			if *CmdOptions.libraryPtr != "" {
				runOptions.Library = config.GetPluginsLibrary()
			}
			err = RunFromJSONStrOrFile(&pmstatus, jsonStrOrFile, runOptions)
			output.Write(pmstatus)
		}
	} else if pluginType != "" {
		switch cmd {
		case "list":
			err = ListFromLibrary(pluginType, config.GetPluginsLibrary())

		case "run":
			pmstatus := RunStatus{}
			err = RunFromLibrary(&pmstatus, pluginType,
				RunOptions{Library: config.GetPluginsLibrary(),
					Sequential: *CmdOptions.sequential})
			output.Write(pmstatus)
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
