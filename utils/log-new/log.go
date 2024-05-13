// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package logNew

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Log info such as log directory and log file name.
type Log struct {
	// Dir indicates the location for writing log file.
	Dir string

	// File indicates the log file name to write to in the logDir location.
	File string

	// Log handler
	Logger *log.Logger
}

// cmdOptions contains commandline parameters/options for generating output in specified format.
var cmdOptions Log

var curLogFile struct {
	absLogFileNoTimestamp string
	timestamp             string
}

// GetCurLogFile provides file path with option to include or not include  timestamp and file extension.
func (l *Log) GetCurLogFile(timestamp bool, extension bool) string {
	return l.File
}

// GetLogDir provides location for storing logs.
func (l *Log) GetLogDir() string {
	return l.Dir
}

// GetLogFile provides name of logfile.
func (l *Log) GetLogFile() string {
	return cmdOptions.File
}

func GetLogDir() string {
	return filepath.FromSlash(filepath.Clean(cmdOptions.Dir) + string(os.PathSeparator))
}

// GetLogFile provides name of logfile.
func GetLogFile() string {
	return cmdOptions.File
}

// PrintNLog prints to console as well as logs to file.
func PrintNLog(format string, a ...interface{}) {
	cmdOptions.PrintNLog(format, a...)
}

// PrintNLog prints to console as well as logs to file.
func (l *Log) PrintNLog(format string, a ...interface{}) {
	l.Logger.Printf(format, a...)
	fmt.Printf(format, a...)
}

func PrintNLogError(format string, a ...interface{}) error {
	return cmdOptions.PrintNLogError(format, a...)
}

// PrintNLogError prints to console and log file along with returning the
// message in error format.
func (l *Log) PrintNLogError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	l.Logger.Printf("ERROR: %v", err.Error())
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err.Error())
	return err
}

func PrintNLogWarning(format string, a ...interface{}) error {
	return cmdOptions.PrintNLogWarning(format, a...)
}

// PrintNLogWarning prints warning message to console and log file along with
// returning the message in error format.
func (l *Log) PrintNLogWarning(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	l.Logger.Printf("WARNING: %v", err.Error())
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
		&cmdOptions.Dir,
		"log-dir",
		defaultLogDir,
		"Directory for the log file.",
	)
	f.StringVar(
		&cmdOptions.File,
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
func (l *Log) SetLogging(myLogFile string) error {
	// setNewTimeStampForLogFile()
	// curLogFile.absLogFileNoTimestamp = filepath.Clean(myLogFile)
	// tLogFile := GetCurLogFile(true, true)
	l.File = myLogFile

	fh, err := os.OpenFile(l.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("os.OpenFile(%s) Error: %s", l.File, err.Error())
		return PrintNLogError("Failed to open log file: %s", l.File)
	}
	// defer fh.Close()
	l.Logger = log.New(fh, "", log.LstdFlags)
	l.Logger.SetOutput(fh)
	// l.Logger.Println("TASK:")

	return nil
}
