// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	logger "github.com/VeritasOS/plugin-manager/utils/log"
	"gopkg.in/yaml.v3"
)

// Config is Plugin Manager's configuration information.
type Config struct {
	// PluginManager configuration information.
	PluginManager struct {
		// Library is the path where plugin directories containing plugin files are present.
		Library  string `yaml:"library"`
		LogDir   string `yaml:"log dir"`
		LogFile  string `yaml:"log file"`
		LogLevel string `yaml:"log level"`
	}
}

var myConfig Config

var (
	// EnvConfFile is environment variable containing the config file path.
	EnvConfFile string
	// DefaultConfigPath is default path for config file used when EnvConfFile is not set.
	DefaultConfigPath string
	// DefaultLogPath is default path for log file.
	DefaultLogPath string
)

// GetLogDir provides location for storing logs.
func GetLogDir() string {
	// TODO: Move log parameters one level up in config as it's common to asum,
	// and not specific to PM.
	return filepath.FromSlash(filepath.Clean(myConfig.PluginManager.LogDir) +
		string(os.PathSeparator))
}

// GetLogFile provides name of logfile.
func GetLogFile() string {
	// TODO: Move log parameters one level up in config as it's common to asum,
	// and not specific to PM.
	return myConfig.PluginManager.LogFile
}

// GetLogLevel provides name of loglevel.
func GetLogLevel() string {
	return myConfig.PluginManager.LogLevel
}

// GetPluginsLibrary gets location of plugins library.
func GetPluginsLibrary() string {
	return filepath.FromSlash(filepath.Clean(myConfig.PluginManager.Library) +
		string(os.PathSeparator))
}

// GetPMLogDir provides location for storing Plugin Manager logs.
//
//	NOTE: The plugin logs would be stored "plugins" directory under the
//	same path, and use GetPluginsLogDir() to get that path.
func GetPMLogDir() string {
	return filepath.FromSlash(filepath.Clean(myConfig.PluginManager.LogDir) +
		string(os.PathSeparator))
}

// GetPMLogFile gets the file for storing Plugin Manager logs.
func GetPMLogFile() string {
	return myConfig.PluginManager.LogFile
}

// GetPluginsLogDir provides location for storing individual plugins execution logs.
func GetPluginsLogDir() string {
	return GetPMLogDir() + "plugins" + string(os.PathSeparator)
}

// Load config information
func Load() error {
	logger.Debug.Println("Entering config.Load()")
	defer logger.Debug.Println("Exiting config.Load()")

	myConfigFile := os.Getenv(EnvConfFile)
	if myConfigFile == "" {
		logger.Info.Printf("%s env is not set. Using default config file.", EnvConfFile)
		myConfigFile = DefaultConfigPath
	}
	myConfigFile = filepath.FromSlash(myConfigFile)
	logger.Debug.Printf("config file: %s", myConfigFile)
	var err error
	myConfig, err = readConfigFile(myConfigFile)
	logger.Debug.Printf("Plugin Manager Config: %+v", myConfig)
	return err
}

func readConfigFile(confFilePath string) (Config, error) {
	logger.Debug.Printf("Entering readConfigFile(%s)", confFilePath)
	defer logger.Debug.Println("Exiting readConfigFile")

	var conf Config
	bFileContents, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return conf, logger.ConsoleError.PrintNReturnError("Failed to read \"" +
			confFilePath + "\" file.")
	}

	err = yaml.Unmarshal(bFileContents, &conf)
	if err != nil {
		logger.Error.Printf("Failed to call yaml.Unmarshal(%s, %s); err=%s",
			bFileContents, &conf, err.Error())
		return conf, logger.ConsoleError.PrintNReturnError("Failed to parse %s config file.", confFilePath)
	}

	logger.Debug.Printf("Config: %+v", conf)
	return conf, nil
}

// SetLogDir sets the location for storing Plugin Manager logs.
//
//	Use GetPMLogDir() to obtain this location from config.
//	NOTE: The plugin logs would be stored "plugins" directory under the
//	same path, and use GetPluginsLogDir() to get that path.
func SetLogDir(logDir string) {
	// TODO: Move log parameters one level up in config as it's common to asum,
	// and not specific to PM.
	myConfig.PluginManager.LogDir = filepath.FromSlash(
		filepath.Clean(logDir) + string(os.PathSeparator))
}

// SetLogFile sets the log file to use.
//
//	Use GetLogDir() to obtain this location from config.
//	NOTE: The plugin logs would be stored "plugins" directory under the
//	same path, and use Get/SetPluginsLogDir() to get/set that path.
func SetLogFile(logFile string) {
	// TODO: Move log parameters one level up in config as it's common to asum,
	// and not specific to PM.
	myConfig.PluginManager.LogFile = logFile
}

// SetLogLevel sets the log level.
func SetLogLevel(logLevel string) {
	myConfig.PluginManager.LogLevel = logLevel
}

// SetPluginsLibrary sets the plugins library location.
func SetPluginsLibrary(library string) {
	myConfig.PluginManager.Library = filepath.FromSlash(
		filepath.Clean(library) + string(os.PathSeparator))
}

// SetPMLogFile sets the file for storing Plugin Manager logs.
func SetPMLogFile(logfile string) {
	// add .log suffix if it doesn't exist.
	if !strings.HasSuffix(logfile, ".log") {
		logfile += ".log"
	}
	myConfig.PluginManager.LogFile = logfile
}

// SetPMLogDir sets the location for storing Plugin Manager logs.
//
//	Use GetPMLogDir() to obtain this location from config.
//	NOTE: The plugin logs would be stored "plugins" directory under the
//	same path, and use GetPluginsLogDir() to get that path.
func SetPMLogDir(logDir string) {
	myConfig.PluginManager.LogDir = filepath.FromSlash(
		filepath.Clean(logDir) + string(os.PathSeparator))
}
