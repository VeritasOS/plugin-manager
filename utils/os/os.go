// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package os contains some utility functions as well as pointers to os package
// functions so as to help in unit testing (mocking).
package os

import (
	"os"
	"os/exec"
	"strings"
	"sync"

	logger "github.com/VeritasOS/plugin-manager/utils/log"
)

// ExecCommand a pointer to exec.Command
var ExecCommand = exec.Command

// os package related functions.
var (
	// OsOpenFile is a pointer to os.OpenFile
	OsOpenFile = os.OpenFile

	// OsRemove is a pointer to os.RemoveAll
	OsRemoveAll = os.RemoveAll

	// OsMkdirAll is a pointer to os.MkdirAll
	OsMkdirAll = os.MkdirAll

	// OsEnviron is a pointer to os.Environ
	OsEnviron = os.Environ

	// Map maintaining os environment variables.
	envMap       = map[string]string{}
	isEnvMapInit sync.Once
)

// EnvMap returns OS environment variable values in a map.
// Any ENV variables set after this program is started (or func is called first time) would not be part of the map.
func EnvMap() map[string]string {
	logger.Debug.Printf("In EnvMap()")
	defer logger.Debug.Printf("Exiting EnvMap()")
	// If already initialized once, return that map.
	isEnvMapInit.Do(initEnvMap)
	// Make copy as maps are passed/returned as reference.
	// INFO: maps.Copy requires 1.21 golang version on builder.
	// maps.Copy(em, envMap)
	em := make(map[string]string, len(envMap))
	for k, v := range envMap {
		em[k] = v
	}
	return em
}

func initEnvMap() {
	logger.Debug.Printf("In initEnvMap()")
	defer logger.Debug.Printf("Exiting initEnvMap()")

	for _, keyval := range OsEnviron() {
		// INFO: strings.Cut requires 1.18 golang version on builder.
		// key, val, found := strings.Cut(keyval, "=")
		fields := strings.SplitN(keyval, "=", 2)
		if len(fields) == 2 {
			key := fields[0]
			val := fields[1]
			// logger.Debug.Printf("%v (key) = %v (value)", key, val)
			envMap[key] = val
		}
	}
}
