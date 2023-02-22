// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9
package pg

import (
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/VeritasOS/plugin-manager/config"
	"github.com/VeritasOS/plugin-manager/pluginmanager"
)

func Test_getStatusColor(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}

	type args struct {
		status string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Start",
			args: args{status: pluginmanager.DStatusStart},
			want: "blue",
		},
		{
			name: "Ok/Pass",
			args: args{status: pluginmanager.DStatusOk},
			want: "green",
		},
		{
			name: "Fail",
			args: args{status: pluginmanager.DStatusFail},
			want: "red",
		},
		{
			name: "Skip",
			args: args{status: pluginmanager.DStatusSkip},
			want: "yellow",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStatusColor(tt.args.status); got != tt.want {
				t.Errorf("getStatusColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

// getPluginType returns the plugin type of the specified plugin file.
func getPluginType(file string) string {
	return strings.Replace(path.Ext(file), ".", ``, -1)
}

func Test_UpdateGraph(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}
	type args struct {
		plugin string
		status string
		url    string
	}
	type wants struct {
		rows []string
		url  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   wants
	}{
		{
			name: "Append a row",
			args: args{
				plugin: "A/a.test",
				status: pluginmanager.DStatusOk,
				url:    "url/A/a.test",
			},
			wantErr: false,
			wants: wants{
				rows: []string{"\"A/a.test\" [style=filled,fillcolor=green" + ",URL=\"url/A/a.test\"]"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateGraph(getPluginType(tt.args.plugin), tt.args.plugin, tt.args.status, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("UpdateGraph() error = %v, wantErr %v", err, tt.wantErr)
			}
			rowsInterface, _ := g.subgraph.Load(getPluginType(tt.args.plugin))
			rows := rowsInterface.([]string)
			if !reflect.DeepEqual(rows, tt.wants.rows) {
				t.Errorf("UpdateGraph() g.rows = %v, wants.rows %v", rows, tt.wants.rows)
			}
		})
	}
}

func Test_initGraph(t *testing.T) {
	type args struct {
		pluginType  string
		pluginsInfo pluginmanager.Plugins
	}
	tests := []struct {
		name     string
		args     args
		wantrows []string
		wantErr  bool
	}{
		{
			name: "No plugins",
			args: args{
				pluginType:  "prereboot",
				pluginsInfo: pluginmanager.Plugins{},
			},
			wantrows: []string{},
			wantErr:  false,
		},
		{
			name: "One plugin",
			args: args{
				pluginType: "test1",
				pluginsInfo: pluginmanager.Plugins{
					"A/a.test1": {
						Description: "A's description",
						Requires:    []string{},
						ExecStart:   "/bin/echo 'Running A...!'",
					},
				},
			},
			wantrows: []string{
				`"A/a.test1" [label="A's description",style=filled,fillcolor=lightgrey,URL="./A/a.test1"]`,
				`"A/a.test1"`,
			},
			wantErr: false,
		},
		{
			name: "Two independent plugins",
			args: args{
				pluginType: "test2",
				pluginsInfo: pluginmanager.Plugins{
					"A/a.test2": {
						Description: "A's description",
						Requires:    []string{},
						ExecStart:   "/bin/echo 'Running A...!'",
					},
					"B/b.test2": {
						Description: "B's description",
						Requires:    []string{},
						ExecStart:   "/bin/echo 'Running B...!'",
					},
				},
			},
			wantrows: []string{
				`"A/a.test2" [label="A's description",style=filled,fillcolor=lightgrey,URL="./A/a.test2"]`,
				`"A/a.test2"`,
				`"B/b.test2" [label="B's description",style=filled,fillcolor=lightgrey,URL="./B/b.test2"]`,
				`"B/b.test2"`,
			},
			wantErr: false,
		},
		// TODO: Requires normalize plugins
		// {
		// 	name: "Dependent plugin",
		// 	args: args{
		// 		pluginType: "test3",
		// 		pluginsInfo: pluginmanager.Plugins{
		// 			"A/a.test3": {
		// 				Description: "A's description",
		// 				Requires:    []string{},
		// 				ExecStart:   "/bin/echo 'Running A...!'",
		// 			},
		// 			"B/b.test3": {
		// 				Description: "B's description",
		// 				Requires:    []string{"A/a.test3"},
		// 				ExecStart:   "/bin/echo 'Running B...!'",
		// 			},
		// 		},
		// 	},
		// 	wantrows: []string{
		// 		`"A/a.test3" [label="A's description",style=filled,fillcolor=lightgrey,URL="./A/a.test3"]`,
		// 		`"A/a.test3"`,
		// 		`"B/b.test3" [label="B's description",style=filled,fillcolor=lightgrey,URL="./B/b.test3"]`,
		// 		`"B/b.test3"`,
		// 		`"A/a.test3" -> "B/b.test3"`,
		// 		`"A/a.test3" -> "B/b.test3"`,
		// 	},
		// 	wantErr: false,
		// },
	}
	// Set log file name to "test", so that cleaning becomes easier.
	config.SetPMLogFile("test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nPInfo := tt.args.pluginsInfo
			// nPInfo := normalizePluginsInfo(tt.args.pluginsInfo)
			if err := InitGraph(tt.args.pluginType, nPInfo); (err != nil) != tt.wantErr {
				t.Errorf("initGraph() error = %v, wantErr %v", err, tt.wantErr)
			}
			rowsI, _ := g.subgraph.Load(tt.args.pluginType)
			if rowsI == nil {
				if 0 != len(tt.wantrows) {
					t.Errorf("initGraph() got = %+v, want %+v", rowsI, tt.wantrows)
				}
				return
			}
			sort.Strings(rowsI.([]string))
			sort.Strings(tt.wantrows)
			if !reflect.DeepEqual(tt.wantrows, rowsI.([]string)) {
				t.Errorf("initGraph() got = %+v (%d), want %+v (%d)",
					rowsI, len(rowsI.([]string)), tt.wantrows, len(tt.wantrows))
			}
		})
	}
}
