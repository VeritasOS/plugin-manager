// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package main provides a commandline tool interface for interacting with
// Plugin Manager (PM).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logger "github.com/VeritasOS/plugin-manager/utils/log"

	pm "github.com/VeritasOS/plugin-manager"
	"github.com/VeritasOS/plugin-manager/config"
)

var (
	buildDate string
	// Version of the Plugin Manager (PM) command.
	version = "4.2"
	// progname is name of my binary/program/executable.
	progname = filepath.Base(os.Args[0])
)

func init() {
	// EnvConfFile is environment variable containing PluginManager config file path.
	config.EnvConfFile = "PM_CONF_FILE"
	// DefaultConfigPath is default path for config file used when EnvConfFile is not set.
	config.DefaultConfigPath = "/opt/veritas/appliance/asum/pm.config.yaml"
	// DefaultLogPath is default path for log file.
	config.DefaultLogPath = "./" + progname
	// config.DefaultLogPath = "/var/log/asum/pm.log"

	// NOTE: while running tests, the path of binary would be in `/tmp/<go-build*>`,
	// so, using relative logging path w.r.t. binary wouldn't be accessible on Jenkins.
	// So, use path which also has write permissions (like current source directory).
	// Use file logging by default, unless we see log-tag is specified to use syslog.
	err := logger.InitFileLogger(config.DefaultLogPath, "INFO")
	if err != nil {
		fmt.Printf("Failed to initialize logger [%v]. Exiting...\n", err)
		os.Exit(-1)
	}
}

func main() {
	cmd := os.Args[1]
	if cmd == "version" {
		logger.ConsoleInfo.Printf("%s version %s %s", progname, version, buildDate)
		os.Exit(0)
	}

	if err := config.Load(); err != nil {
		logger.ConsoleWarning.Printf("Failed to load config file. Using default values and proceeding with the operation")
	}
	if config.GetPMLogDir() != "" && config.GetPMLogFile() != "" {
		myLogFile := config.GetPMLogDir() + config.GetPMLogFile()
		if !strings.HasSuffix(myLogFile, ".log") {
			myLogFile += ".log"
		}
		if myLogFile != config.DefaultLogPath {
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

	pm.RegisterCommandOptions(progname)
	err := pm.ScanCommandOptions(nil)
	if err != nil {
		os.Exit(1)
	}
}
