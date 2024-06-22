// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package logger

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Type is used to track logger type
type Type int

// Below logger types are supported.
const (
	FileLog = 1 << iota
	SysLog
)

const syslogFacility = syslog.LOG_LOCAL0
const syslogConfig = "/etc/rsyslog.d/10-vxos-asum.conf"

// SyslogTagPrefix defines tag name for syslog.
const SyslogTagPrefix = "vxos-asum@"

// Config is used to track user defined configuration.
type Config struct {
	loggerType  Type
	level       string
	module      string
	logfilePath string
}

// FileLogConfig Setup FileLog Config
func FileLogConfig(level, file, module string) Config {
	return Config{loggerType: FileLog, level: level, module: module, logfilePath: file}
}

// SyslogConfig Setup SysLog Config
func SyslogConfig(level, module string) Config {
	return Config{loggerType: SysLog, level: level, module: module, logfilePath: ""}
}

// ConsoleLogger will print message on screen and write to log file
// The message will always be printed regardless the log level
// The writing action depends on the log level
type ConsoleLogger struct {
	prefix string
	logger *log.Logger
}

// Logger implements functions for all log levels
type Logger struct {
	//Logger for log Debug messages.
	debug *log.Logger
	//Logger for log Info messages.
	info *log.Logger
	//Logger for log Warning messages.
	warning *log.Logger
	//Logger for log Error messages.
	error *log.Logger
}

// Debug returns the debug logger
func (logger *Logger) Debug() *log.Logger {
	return logger.debug
}

// Info returns the info logger
func (logger *Logger) Info() *log.Logger {
	return logger.info
}

// Warning returns the warning logger
func (logger *Logger) Warning() *log.Logger {
	return logger.warning
}

// Error returns te error logger
func (logger *Logger) Error() *log.Logger {
	return logger.error
}

// Get returns the object of logger type
func Get() *Logger {
	return &Logger{debug: Debug, info: Info, warning: Warning, error: Error}
}

// Printf prints message in console and writes to log file
func (consoleLog *ConsoleLogger) Printf(msg string, args ...interface{}) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Printf(msg, args...)
	str := consoleLog.prefix + msg
	consoleLog.logger.Printf(str, args...)
}

// PrintNReturnError calls Printf and returns wrapped message as error
func (consoleLog *ConsoleLogger) PrintNReturnError(msg string, args ...interface{}) error {
	err := fmt.Errorf(msg, args...)
	consoleLog.Printf(msg, args...)
	return err
}

// LogHandle the log handle interface
type LogHandle interface {
	Write(p []byte) (n int, err error)
	Close() error
}

// CloseLogHandle closes log handler
func CloseLogHandle(handle LogHandle) error {
	if handle != nil {
		err := handle.Close()
		if err == nil {
			handle = nil
		}
		return err
	}
	return nil
}

// FileLogHandle the file log handler
type FileLogHandle struct {
	logFile  *os.File
	hostname *string
}

func (handle *FileLogHandle) Write(p []byte) (n int, err error) {
	t := time.Now()
	// required timestamp format is based on both rfc3339 and ISO 8601
	// YYYY-MM-DDThh:mm:ss.mmm+04:00
	prefix := t.Format(strings.Replace(time.RFC3339, "Z", ".000-", 1))
	if handle.hostname != nil {
		prefix = prefix + " " + *handle.hostname + " "
	}

	buf := []byte(prefix)
	buf = append(buf, p...)

	n, err = handle.logFile.Write(buf)
	if err != nil || n != len(buf) {
		return n, err
	}
	// io.MultiWriter will verify write count, return original bytes without prefixes
	return len(p), nil
}

// Close close the FileLog handler.
func (handle *FileLogHandle) Close() error {
	if handle.logFile != nil {
		return handle.logFile.Close()
	}
	return nil
}

var (
	//ConsoleDebug logger for console and log Debug messages.
	ConsoleDebug ConsoleLogger
	//ConsoleInfo logger for console and log Info messages.
	ConsoleInfo ConsoleLogger
	//ConsoleWarning logger for console and log Warning messages.
	ConsoleWarning ConsoleLogger
	//ConsoleError logger for console and log Error messages.
	ConsoleError ConsoleLogger
	//Debug logger for logging Debug messages.
	Debug *log.Logger
	//Info logger for logging Info messages.
	Info *log.Logger
	//Warning logger for logging Warning messages.
	Warning *log.Logger
	//Error logger for logging Error messages.
	Error             *log.Logger
	fileLogHandle     LogHandle
	syslogDebugHandle LogHandle
	syslogInfoHandle  LogHandle
	syslogWarnHandle  LogHandle
	syslogErrorHandle LogHandle

	fOpenFile       = os.OpenFile
	fCloseLogHandle = CloseLogHandle
	fHostname       = os.Hostname
	fMkdirAll       = os.MkdirAll
	fNewSyslogger   = func(priority syslog.Priority, tag string) (LogHandle, error) {
		return syslog.New(priority, tag)
	}

	// Lock to ensure thread safe behaviour when initializing and de-initializing the singleton logger object
	singleLogger = sync.Mutex{}
)

// IsFileLogger returns true if file log is initialized (and not syslog)
func IsFileLogger() bool {
	if fileLogHandle != nil {
		return true
	}
	return false
}

// initLogger sets logger for all supported log levels
func initLogger(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Debug = log.New(traceHandle,
		"DEBUG: ",
		0)

	Info = log.New(infoHandle,
		"INFO: ",
		0)

	Warning = log.New(warningHandle,
		"WARNING: ",
		0)

	Error = log.New(errorHandle,
		"ERROR: ",
		0)

	ConsoleDebug = ConsoleLogger{"[DEBUG] ", Debug}
	ConsoleInfo = ConsoleLogger{"[INFO] ", Info}
	ConsoleWarning = ConsoleLogger{"[WARNING] ", Warning}
	ConsoleError = ConsoleLogger{"[ERROR] ", Error}
}

func initFileLogHandle(myLogFile string) error {
	if fileLogHandle != nil {
		return nil
	}

	var err error
	myLogFile = filepath.Clean(myLogFile)
	myLogFileNoExt := myLogFile
	if strings.HasSuffix(myLogFile, ".log") {
		myLogFileNoExt = strings.TrimSuffix(myLogFile, filepath.Ext(myLogFile))
	}
	ts := time.Now().Format(time.RFC3339Nano)
	logFile := myLogFileNoExt + "." + ts + ".log"

	_, err = os.Stat(logFile)
	if os.IsNotExist(err) {
		logFileDir := filepath.Dir(logFile)
		err = fMkdirAll(logFileDir, 0755)
		if err != nil {
			return fmt.Errorf("os.MkdirAll(%s) failed", logFileDir)
		}
	}

	file, err := fOpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	hostname, err := fHostname()
	if err != nil {
		hostname = "unknown"
	}
	fh := FileLogHandle{}
	fh.logFile = file
	fh.hostname = &hostname
	fileLogHandle = &fh

	fmt.Println("Log:", logFile)
	return nil
}

func initSyslogHandle(syslogTag string) (err error) {
	if syslogDebugHandle == nil {
		syslogDebugHandle, err = fNewSyslogger(syslogFacility|syslog.LOG_DEBUG, syslogTag)
		if err != nil {
			syslogDebugHandle = nil
			return errors.New("creating syslog debug handle failed")
		}
	}

	if syslogInfoHandle == nil {
		syslogInfoHandle, err = fNewSyslogger(syslogFacility|syslog.LOG_INFO, syslogTag)
		if err != nil {
			syslogInfoHandle = nil
			return errors.New("creating syslog info handle failed")
		}
	}

	if syslogWarnHandle == nil {
		syslogWarnHandle, err = fNewSyslogger(syslogFacility|syslog.LOG_WARNING, syslogTag)
		if err != nil {
			syslogWarnHandle = nil
			return errors.New("creating syslog warning handle failed")
		}
	}
	if syslogErrorHandle == nil {
		syslogErrorHandle, err = fNewSyslogger(syslogFacility|syslog.LOG_ERR, syslogTag)
		if err != nil {
			syslogErrorHandle = nil
			return errors.New("creating syslog error handle failed")
		}
	}
	return nil
}

// InitializeLogger initializes customized file logger or syslog logger with given log config
func InitializeLogger(config Config) error {
	singleLogger.Lock()
	defer singleLogger.Unlock()
	var debugWriter, infoWriter, warnWriter, errWriter io.Writer
	if config.loggerType == FileLog {
		if err := initFileLogHandle(config.logfilePath); err != nil {
			return err
		}
		debugWriter = fileLogHandle
		infoWriter = fileLogHandle
		warnWriter = fileLogHandle
		errWriter = fileLogHandle
	} else {
		syslogTag := config.module
		if strings.Index(syslogTag, SyslogTagPrefix) != 0 {
			// all flex appliance syslog tag starts with same prefix, in case we could filter the logs
			syslogTag = SyslogTagPrefix + syslogTag
		}
		if err := initSyslogHandle(syslogTag); err != nil {
			return err
		}
		debugWriter = syslogDebugHandle
		infoWriter = syslogInfoHandle
		warnWriter = syslogWarnHandle
		errWriter = syslogErrorHandle
	}

	switch config.level {
	case "DEBUG":
		// Enabled logging levels: debug, info, warning, error
		initLogger(debugWriter, infoWriter, warnWriter, errWriter)
	case "INFO":
		// Enabled logging levels: info, warning, error
		initLogger(ioutil.Discard, infoWriter, warnWriter, errWriter)
	case "WARNING":
		// Enabled logging levels: warning, error
		initLogger(ioutil.Discard, ioutil.Discard, warnWriter, errWriter)
	default:
		// Enabled logging levels: error only
		initLogger(ioutil.Discard, ioutil.Discard, ioutil.Discard, errWriter)
	}
	Info.Println("CMD:", os.Args[:])
	return nil
}

// InitFileLogger initializes logger with given log level.
func InitFileLogger(logFile, logLevel string) error {
	return InitializeLogger(FileLogConfig(logLevel, logFile, ""))
}

// IsSysLogConfigPresent indicates whether syslog config is present.
func IsSysLogConfigPresent() bool {
	_, err := os.Stat(syslogConfig)
	if err != nil {
		return false
	}
	return true
}

// InitSysLogger initializes logger with given log level.
func InitSysLogger(module, logLevel string) error {
	// Make sure the config file exist
	if !IsSysLogConfigPresent() {
		return fmt.Errorf("syslog config file %v is not present", syslogConfig)
	}
	return InitializeLogger(SyslogConfig(logLevel, module))
}

// DeInitLogger closes the log file and syslog handler
func DeInitLogger() []error {
	singleLogger.Lock()
	defer singleLogger.Unlock()
	var errList []error
	closeErr := fCloseLogHandle(fileLogHandle)
	if closeErr != nil {
		errList = append(errList, closeErr)
	}
	fileLogHandle = nil

	closeErr = fCloseLogHandle(syslogDebugHandle)
	if closeErr != nil {
		errList = append(errList, closeErr)
	}
	syslogDebugHandle = nil

	closeErr = fCloseLogHandle(syslogInfoHandle)
	if closeErr != nil {
		errList = append(errList, closeErr)
	}
	syslogInfoHandle = nil

	closeErr = fCloseLogHandle(syslogWarnHandle)
	if closeErr != nil {
		errList = append(errList, closeErr)
	}
	syslogWarnHandle = nil

	closeErr = fCloseLogHandle(syslogErrorHandle)
	if closeErr != nil {
		errList = append(errList, closeErr)
	}
	syslogErrorHandle = nil

	return errList
}
