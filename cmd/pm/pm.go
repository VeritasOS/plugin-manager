// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package main provides a commandline tool interface for interacting with
// Plugin Manager (PM).
package main

import (
	"log"
	"os"
	"path/filepath"

	pm "github.com/VeritasOS/plugin-manager"
	"github.com/VeritasOS/plugin-manager/config"
	logutil "github.com/VeritasOS/plugin-manager/utils/log"
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
	config.DefaultConfigPath = "/etc/pm.config.yaml"
	// DefaultLogPath is default path for log file.
	config.DefaultLogPath = "/var/log/pm"

	// INFO: Use DefaultLogPath when it's available (until the config file is read).
	// 	If not, use basename of file.
	// NOTE: while running tests, the path of binary would be in `/tmp/<go-build*>`,
	// so, using relative logging path w.r.t. binary wouldn't be accessible on Jenkins.
	// So, use absolute path which also has write permissions (like current source directory).
	myLogFile := config.DefaultLogPath
	if _, err := os.Stat(filepath.Dir(myLogFile)); os.IsNotExist(err) {
		myLogFile = progname
	}
	logutil.SetLogging(myLogFile)
}

func main() {
	log.Println("Entering main::main with", os.Args[:])
	defer log.Println("Exiting main::main")

	cmd := os.Args[1]
	if cmd == "version" {
		logutil.PrintNLog("%s version %s %s\n",
			progname, version, buildDate)
		os.Exit(0)
	}

	if err := config.Load(); err != nil {
		logutil.PrintNLogWarning("Failed to load config file. Using default values " +
			"and proceeding with the operation")
	}
	if config.GetPMLogDir() != "" && config.GetPMLogFile() != "" {
		myLogFile := config.GetPMLogDir() + config.GetPMLogFile()
		if myLogFile != config.DefaultLogPath {
			logutil.SetLogging(myLogFile)
		}
	}

	pm.RegisterCommandOptions(progname)
	err := pm.ScanCommandOptions(nil)
	if err != nil {
		os.Exit(1)
	}
}
