// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9
package pm

import (
	"os"
	"reflect"
	"sort"
	"testing"
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
			args: args{status: dStatusStart},
			want: "blue",
		},
		{
			name: "Ok/Pass",
			args: args{status: dStatusOk},
			want: "green",
		},
		{
			name: "Fail",
			args: args{status: dStatusFail},
			want: "red",
		},
		{
			name: "Skip",
			args: args{status: dStatusSkip},
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

func Test_updateGraph(t *testing.T) {
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
				status: dStatusOk,
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
			if err := updateGraph(getPluginType(tt.args.plugin), tt.args.plugin, tt.args.status, tt.args.url); (err != nil) != tt.wantErr {
				t.Errorf("updateGraph() error = %v, wantErr %v", err, tt.wantErr)
			}
			rowsInterface, _ := g.subgraph.Load(getPluginType(tt.args.plugin))
			rows := rowsInterface.([]string)
			if !reflect.DeepEqual(rows, tt.wants.rows) {
				t.Errorf("updateGraph() g.rows = %v, wants.rows %v", rows, tt.wants.rows)
			}
		})
	}
}

func Test_initGraph(t *testing.T) {
	type args struct {
		pluginType  string
		pluginsInfo Plugins
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
				pluginsInfo: Plugins{},
			},
			wantrows: []string{},
			wantErr:  false,
		},
		{
			name: "One plugin",
			args: args{
				pluginType: "test1",
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test1",
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
				pluginsInfo: Plugins{
					{

						Name:        "A/a.test2",
						Description: "A's description",
						Requires:    []string{},
						ExecStart:   "/bin/echo 'Running A...!'",
					},
					{

						Name:        "B/b.test2",
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
		{
			name: "Dependent plugin",
			args: args{
				pluginType: "test3",
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test3",
						Description: "A's description",
						Requires:    []string{},
						ExecStart:   "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.test3",
						Description: "B's description",
						Requires:    []string{"A/a.test3"},
						ExecStart:   "/bin/echo 'Running B...!'",
					},
				},
			},
			wantrows: []string{
				`"A/a.test3" [label="A's description",style=filled,fillcolor=lightgrey,URL="./A/a.test3"]`,
				`"A/a.test3"`,
				`"B/b.test3" [label="B's description",style=filled,fillcolor=lightgrey,URL="./B/b.test3"]`,
				`"B/b.test3"`,
				`"A/a.test3" -> "B/b.test3"`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initGraph(tt.args.pluginType, tt.args.pluginsInfo); (err != nil) != tt.wantErr {
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
