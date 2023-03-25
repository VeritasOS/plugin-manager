// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

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

// SetLogging sets the logfile for this program.
// To handle log rotation, the specified myLogFile would be suffixed with
// the current date before the log file extension.
// Ex: If myLogFile := "/log/asum/pm", then the logfile used for
// logging would be "/log/asum/pm.20181212235500.0000.log".
func SetLogging(myLogFile string) error {
	ts := time.Now().Format(time.RFC3339Nano)
	myLogFile = filepath.Clean(myLogFile)
	tLogFile := myLogFile + "." + ts + ".log"

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
