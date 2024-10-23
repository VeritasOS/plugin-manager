// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package main provides a commandline tool interface for interacting with
// Plugin Manager (PM).
package main

import (
	"os"
	"path/filepath"

	logger "github.com/VeritasOS/plugin-manager/utils/log"

	pm "github.com/VeritasOS/plugin-manager"
	"github.com/VeritasOS/plugin-manager/config"
)

var (
	buildDate string
	// Version of the Plugin Manager (PM) command.
	version = "4.4"
	// progname is name of my binary/program/executable.
	progname = filepath.Base(os.Args[0])
)

func init() {
	// EnvConfFile is environment variable containing PluginManager config file path.
	config.EnvConfFile = "PM_CONF_FILE"
	// DefaultConfigPath is default path for config file used when EnvConfFile is not set.
	config.DefaultConfigPath = "/opt/veritas/appliance/asum/pm.config.yaml"
	logger.InitLogging()
}

func main() {
	logger.Debug.Println("Entering main::main with", os.Args[:])
	defer logger.Debug.Println("Exiting main::main")

	cmd := os.Args[1]
	if cmd == "version" {
		logger.ConsoleInfo.Printf("%s version %s %s", progname, version, buildDate)
		os.Exit(0)
	}

	if err := config.Load(); err != nil {
		logger.ConsoleWarning.Printf("Failed to load config file. Using default values and proceeding with the operation")
	}

	pm.RegisterCommandOptions(progname)
	err := pm.ScanCommandOptions(nil)
	if err != nil {
		os.Exit(1)
	}
}
