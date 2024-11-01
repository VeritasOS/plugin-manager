// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

// Package types defines new plugin manager types.
package types

import "github.com/VeritasOS/plugin-manager/types/runtime"

// Plugin is plugin's info: name, description, cmd to run, status, stdouterr.
type Plugin struct {
	Name        string
	Description string
	RequiredBy  []string
	Requires    []string
	ExecStart   string
	Plugins     Plugins
	Library     string
	// TODO: Add Percentage to get no. of pending vs. completed run of plugins.
	RunTime   runtime.RunTime
	Status    string
	StdOutErr []string
}

// Plugins is a list of plugins' info.
type Plugins []*Plugin
