// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package log

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// cmdOptions contains commandline parameters/options for generating output in specified format.
var cmdOptions struct {
	// logDir indicates the location for writing log file.
	logDir string

	// logFile indicates the log file name to write to in the logDir location.
	logFile string
}

var curLogFile struct {
	absLogFileNoTimestamp string
	timestamp             string
}

// GetCurLogFile provides file path with option to include or not include  timestamp and file extension.
func GetCurLogFile(timestamp bool, extension bool) string {
	logFile := curLogFile.absLogFileNoTimestamp
	if timestamp {
		logFile += "." + curLogFile.timestamp
	}
	if extension {
		logFile += ".log"
	}
	return logFile
}

func setNewTimeStampForLogFile() string {
	curLogFile.timestamp = time.Now().Format(time.RFC3339Nano)
	return curLogFile.timestamp
}

// GetLogDir provides location for storing logs.
func GetLogDir() string {
	return filepath.FromSlash(filepath.Clean(cmdOptions.logDir) +
		string(os.PathSeparator))
}

// GetLogFile provides name of logfile.
func GetLogFile() string {
	return cmdOptions.logFile
}

// PrintNLog prints to console as well as logs to file.
func PrintNLog(format string, a ...interface{}) {
	log.Printf(format, a...)
	fmt.Printf(format, a...)
}

// PrintNLogError prints to console and log file along with returning the
// message in error format.
func PrintNLogError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	log.Printf("ERROR: %v", err.Error())
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err.Error())
	return err
}

// PrintNLogWarning prints warning message to console and log file along with
// returning the message in error format.
func PrintNLogWarning(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	log.Printf("WARNING: %v", err.Error())
	fmt.Fprintf(os.Stderr, "WARNING: %v\n", err.Error())
	return err
}

// RegisterCommandOptions registers the command options related to the log options.
func RegisterCommandOptions(f *flag.FlagSet, defaultParams map[string]string) {
	defaultLogDir, ok := defaultParams["log-dir"]
	if !ok {
		defaultLogDir = ""
	}
	defaultLogFile, ok := defaultParams["log-file"]
	if !ok {
		defaultLogFile = ""
	}
	f.StringVar(
		&cmdOptions.logDir,
		"log-dir",
		defaultLogDir,
		"Directory for the log file.",
	)
	f.StringVar(
		&cmdOptions.logFile,
		"log-file",
		defaultLogFile,
		"Name of the log file.",
	)
}

// SetLogging sets the logfile for this program.
// To handle log rotation, the specified myLogFile would be suffixed with
// the current date before the log file extension.
// Ex: If myLogFile := "/log/asum/pm", then the logfile used for
// logging would be "/log/asum/pm.20181212235500.0000.log".
func SetLogging(myLogFile string) error {
	setNewTimeStampForLogFile()
	curLogFile.absLogFileNoTimestamp = filepath.Clean(myLogFile)
	tLogFile := GetCurLogFile(true, true)

	fh, err := os.OpenFile(tLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("os.OpenFile(%s) Error: %s", tLogFile, err.Error())
		return PrintNLogError("Failed to open log file: %s", tLogFile)
	}
	// defer fh.Close()
	log.SetOutput(fh)
	fmt.Println("Log:", tLogFile)
	log.Println("CMD:", os.Args[:])
	return nil
}
