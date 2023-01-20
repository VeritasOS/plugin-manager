// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package os contains pointers to os package functions so as to help
// in unit testing (mocking).
package os

import (
	"os"
	"os/exec"
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
)
