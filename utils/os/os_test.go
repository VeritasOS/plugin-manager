// Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package os

import (
	"testing"
)

func TestExecCommand(t *testing.T) {
	if ExecCommand == nil {
		t.Error("ExecCommand", "doesn't wrap the expected func", "exec.Command")
	}
}

func TestOsMkdirAll(t *testing.T) {
	if OsMkdirAll == nil {
		t.Error("OsMkdirAll", "doesn't wrap the expected func", "os.MkdirAll")
	}
}

func TestOsOpenFile(t *testing.T) {
	if OsOpenFile == nil {
		t.Error("OsOpenFile", "doesn't wrap the expected func", "os.OpenFile")
	}
}

func TestOsRemoveAll(t *testing.T) {
	if OsRemoveAll == nil {
		t.Error("OsRemoveAll", "doesn't wrap the expected func", "os.RemoveAll")
	}
}
