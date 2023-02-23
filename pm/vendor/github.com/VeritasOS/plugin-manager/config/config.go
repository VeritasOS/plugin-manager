// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	logutil "github.com/VeritasOS/plugin-manager/utils/log"
)

// Config is Plugin Manager's configuration information.
type Config struct {
	// PluginManager configuration information.
	PluginManager struct {
		// Library is the path where plugin directories containing plugin files are present.
		Library string `yaml:"library"`
		LogDir  string `yaml:"log dir"`
		LogFile string `yaml:"log file"`
		// PluginDir is deprecated. Use Library instead.
		PluginDir string `yaml:"plugin dir"`
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

// GetPluginsLibrary gets location of plugins library.
func GetPluginsLibrary() string {
	return filepath.FromSlash(filepath.Clean(myConfig.PluginManager.Library) +
		string(os.PathSeparator))
}

// GetPluginsDir gets location of plugins directory.
//
//	NOTE: This is deprecated, Use GetPluginsLibrary() instead.
func GetPluginsDir() string {
	return filepath.FromSlash(filepath.Clean(myConfig.PluginManager.PluginDir) +
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
	log.Println("Entering config.Load()")
	defer log.Println("Exiting config.Load()")

	myConfigFile := os.Getenv(EnvConfFile)
	if myConfigFile == "" {
		log.Printf("%s env is not set. Using default config file.\n",
			EnvConfFile)
		myConfigFile = DefaultConfigPath
	}
	myConfigFile = filepath.FromSlash(myConfigFile)
	log.Printf("config file: %s\n", myConfigFile)
	var err error
	myConfig, err = readConfigFile(myConfigFile)

	// INFO: Library replaces PluginDir.
	// 	Keeping PluginDir for backward compatibility for couple of releases.
	// 	(Currently Px is 1.x, and we can keep say until Px 3.x).
	// 	If older versions of PM is out there, then PluginDir value will be read from config file,
	// 	 and assigned to Library variable.
	log.Printf("Plugin Manager Config: %+v", myConfig)
	if myConfig.PluginManager.Library == "" &&
		myConfig.PluginManager.PluginDir != "" {
		myConfig.PluginManager.Library = myConfig.PluginManager.PluginDir
	}

	return err
}

func readConfigFile(confFilePath string) (Config, error) {
	log.Printf("Entering readConfigFile(%s)", confFilePath)
	defer log.Println("Exiting readConfigFile")

	var conf Config
	bFileContents, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return conf, logutil.PrintNLogWarning("Failed to read \"" +
			confFilePath + "\" file.")
	}

	err = yaml.Unmarshal(bFileContents, &conf)
	if err != nil {
		log.Printf("yaml.Unmarshal(%s, %s); Error: %s",
			bFileContents, &conf, err.Error())
		return conf, logutil.PrintNLogError("Failed to parse %s config file.", confFilePath)
	}

	log.Printf("Config: %+v\n", conf)
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

// SetPluginsLibrary sets the plugins library location.
func SetPluginsLibrary(library string) {
	myConfig.PluginManager.Library = filepath.FromSlash(
		filepath.Clean(library) + string(os.PathSeparator))
}

// SetPluginsDir sets location of plugins directory.
//
//	NOTE: This is deprecated, Use SetPluginsLibrary() instead.
func SetPluginsDir(library string) {
	myConfig.PluginManager.PluginDir = filepath.FromSlash(filepath.Clean(library) +
		string(os.PathSeparator))
}

// SetPMLogFile sets the file for storing Plugin Manager logs.
func SetPMLogFile(logfile string) {
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
