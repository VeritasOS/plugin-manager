// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package pm

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/VeritasOS/plugin-manager/config"
	logger "github.com/VeritasOS/plugin-manager/utils/log"
)

func initTestLogging( /*t *testing.T*/ ) {
	myLogFile := "pm.log"
	if config.GetPMLogFile() != "" {
		myLogFile = config.GetPMLogFile()
	}
	if config.GetPMLogDir() != "" {
		myLogFile = config.GetPMLogDir() + myLogFile
	}
	// t.Logf("Logging to specified log file: %s", myLogFile)
	errList := logger.DeInitLogger()
	if len(errList) > 0 {
		fmt.Printf("Failed to deinitialize logger, err=[%v]", errList)
		os.Exit(-1)
	}
	err := logger.InitFileLogger(myLogFile, "DEBUG")
	if err != nil {
		fmt.Printf("Failed to initialize logger, err=[%v]", err)
		os.Exit(-1)
	}
}

func init() {
	initTestLogging()
}

func Test_getPluginFiles(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}
	myConfigFile := os.Getenv(config.EnvConfFile)
	if myConfigFile == "" {
		// For case, where tests are run through IDE.
		myConfigFile = filepath.FromSlash("./sample/pm.config.yaml")
	}
	wd, _ := os.Getwd()
	t.Logf("PWD: %s;\nConfig file: %+v\n", wd, myConfigFile)
	config.SetPluginsLibrary(filepath.FromSlash(filepath.Dir(myConfigFile) + "/library"))
	// t.Logf("Config: %+v\n", myConfig)
	tests := []struct {
		name       string
		pluginType string
		output     struct {
			pluginFiles []string
			err         error
		}
	}{
		// TODO: Figure out why reflect.DeepEqual fails even
		// when expected is same as actual for this case!

		// {
		// 	name: "No postreboot-validate plugin file",
		// 	pluginType:  "postreboot-validate",
		// 	output: struct {
		// 		pluginFiles []string
		// 		err         error
		// 	}{
		// 		pluginFiles: []string{},
		// 		err:         nil,
		// 	},
		// },

		{
			name:       "1 postreboot plugin file",
			pluginType: "postreboot",
			output: struct {
				pluginFiles []string
				err         error
			}{
				pluginFiles: []string{filepath.FromSlash("A/a.postreboot")},
				err:         nil,
			},
		},
		{
			name:       "4 prereboot plugin files",
			pluginType: "prereboot",
			output: struct {
				pluginFiles []string
				err         error
			}{
				pluginFiles: []string{
					filepath.FromSlash("A/a.prereboot"),
					filepath.FromSlash("B/b.prereboot"),
					filepath.FromSlash("C/c.prereboot"),
					filepath.FromSlash("D/d.prereboot"),
				},
				err: nil,
			},
		},
		// {
		// 	name:       "Not a plugins library",
		// 	pluginType: "test",
		// 	output: struct {
		// 		pluginFiles []string
		// 		err         error
		// 	}{
		// 		pluginFiles: []string{},
		// 		err:         nil,
		// 	},
		// },
	}

	library := config.GetPluginsLibrary()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resFiles, resStatus := getPluginFiles(tt.pluginType, library)
			if resStatus != tt.output.err {
				t.Errorf("Status: got %+v, want %+v", resStatus, tt.output.err)
			}
			if reflect.DeepEqual(resFiles, tt.output.pluginFiles) == false {
				t.Errorf("File list: got %+v, want %+v", resFiles, tt.output.pluginFiles)
			}
		})
	}
}

func Test_validateDependencies(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}

	type args struct {
		pluginsInfo Plugins
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "No dependencies",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						Requires:    []string{},
						ExecStart:   "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.test",
						Description: "Applying \"B\" settings",
						Requires:    []string{},
						ExecStart:   "/bin/echo \"Running B...\"",
					},
				},
			},
			want: []string{
				"A/a.test",
				"B/b.test",
			},
			wantErr: false,
		},
		{
			name: "Single dependency",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						Requires:    []string{},
						ExecStart:   "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.test",
						Description: "Applying \"B\" settings",
						Requires: []string{
							"A/a.test",
						},
						ExecStart: "/bin/echo \"Running B...\"",
					},
				},
			},
			want: []string{
				"A/a.test",
				"B/b.test",
			},
			wantErr: false,
		},
		{
			name: "Multiple dependencies",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						Requires:    []string{},
						ExecStart:   "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.test",
						Description: "Applying \"B\" settings",
						Requires:    []string{},
						ExecStart:   "/bin/echo \"Running B...\"",
					},
					{
						Name:        "C/c.test",
						Description: "Applying \"C\" settings",
						Requires: []string{
							"A/a.test",
							"B/b.test",
						},
						ExecStart: "/bin/echo \"Running C...\"",
					},
				},
			},
			want: []string{
				"A/a.test",
				"B/b.test",
				"C/c.test",
			},
			wantErr: false,
		},
		{
			name: "Multi-level dependencies",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						Requires: []string{
							"D/d.test",
							"C/c.test",
						},
						ExecStart: "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.test",
						Description: "Applying \"B\" settings",
						Requires:    []string{},
						ExecStart:   "/bin/echo \"Running B...\"",
					},
					{
						Name:        "C/c.test",
						Description: "Applying \"C\" settings",
						Requires:    []string{},
						ExecStart:   "/bin/echo \"Running C...\"",
					},
					{
						Name:        "D/d.test",
						Description: "Applying \"D\" settings",
						Requires: []string{
							"B/b.test",
						},
						ExecStart: "/bin/echo 'Running D...!'",
					},
				},
			},
			want: []string{
				"B/b.test",
				"C/c.test",
				"D/d.test",
				"A/a.test",
			},
			wantErr: false,
		},
		{
			name: "Direct circular dependency",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						Requires: []string{
							"B/b.test",
						},
						ExecStart: "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.test",
						Description: "Applying \"B\" settings",
						Requires: []string{
							"A/a.test",
						},
						ExecStart: "/bin/echo \"Running B...\"",
					},
				},
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "Requires & Required-by circular dependency",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.circular",
						Description: "Applying \"A\" settings",
						Requires: []string{
							"B/b.circular",
						},
						RequiredBy: []string{
							"B/b.circular",
						},
						ExecStart: "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.circular",
						Description: "Applying \"B\" settings",
						ExecStart:   "/bin/echo \"Running B...\"",
					},
				},
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "Indirect circular dependency",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						Requires: []string{
							"B/b.test",
						},
						ExecStart: "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.test",
						Description: "Applying \"B\" settings",
						Requires: []string{
							"C/c.test",
						},
						ExecStart: "/bin/echo \"Running B...\"",
					},
					{
						Name:        "C/c.test",
						Description: "Applying \"C\" settings",
						Requires: []string{
							"A/a.test",
						},
						ExecStart: "/bin/echo \"Running C...\"",
					},
				},
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "Dependency not met",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						Requires: []string{
							"B/b.test",
						},
						ExecStart: "/bin/echo 'Running A...!'",
					},
				},
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "Dependencies not met",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						Requires: []string{
							"C/c.test",
						},
						ExecStart: "/bin/echo 'Running A...!'",
					},
					{
						Name:        "B/b.test",
						Description: "Applying \"B\" settings",
						Requires: []string{
							"C/c.test",
						},
						ExecStart: "/bin/echo \"Running B...\"",
					},
				},
			},
			want:    []string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// INFO: Validating dependencies requires pluginsInfo to be in
			// 	normalized form, so first call normalizePluginsInfo() before
			//  calling validateDependencies().
			nPInfo := normalizePluginsInfo(tt.args.pluginsInfo)
			t.Logf("Normalized plugins info: %v", nPInfo)
			got, err := validateDependencies(nPInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDependencies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateDependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseUnitFile(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}
	tests := []struct {
		name         string
		fileContents string
		pluginInfo   Plugin
	}{
		{
			name:         "Plugin file with no contents",
			fileContents: "",
			pluginInfo:   Plugin{},
		},
		{
			name: "Plugin file with desc & exec",
			fileContents: `
Description=Applying "A" settings
ExecStart=/bin/echo "Running A...!"
`,
			pluginInfo: Plugin{
				Description: "Applying \"A\" settings",
				ExecStart:   "/bin/echo \"Running A...!\"",
			},
		}, {
			name: "Plugin file with comments",
			fileContents: `
Description=Applying "A" settings
# Requires= 
ExecStart=/bin/echo "Running A...!"
`,
			pluginInfo: Plugin{
				Description: "Applying \"A\" settings",
				ExecStart:   "/bin/echo \"Running A...!\"",
			},
		},
		{
			name: "Plugin file with desc, single Requires & exec",
			fileContents: `
Description=Applying "D" settings
Requires=a.test
ExecStart=/bin/echo "Running D...!"
`,
			pluginInfo: Plugin{
				Description: "Applying \"D\" settings",
				Requires:    []string{"a.test"},
				ExecStart:   "/bin/echo \"Running D...!\"",
			},
		},
		{
			name: "Plugin file with desc, multiple Requires & exec",
			fileContents: `
Description=Applying "D" settings
Requires=a.test b.test c.test
ExecStart=/bin/echo "Running D...!"
`,
			pluginInfo: Plugin{
				Description: "Applying \"D\" settings",
				Requires: []string{
					"a.test",
					"b.test",
					"c.test",
				},
				ExecStart: "/bin/echo \"Running D...!\"",
			},
		},
		{
			name: "Plugin file with colon in desc",
			fileContents: `
Description=Applying "A:B" settings
ExecStart=/bin/echo "Running A & B...!"
`,
			pluginInfo: Plugin{
				Description: "Applying \"A:B\" settings",
				ExecStart:   "/bin/echo \"Running A & B...!\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, _ := parseUnitFile(tt.fileContents)
			if reflect.DeepEqual(res, tt.pluginInfo) == false {
				t.Errorf("got %+v, want %+v", res, tt.pluginInfo)
			}
		})
	}
}

func Test_executePlugins(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}
	type want struct {
		returnStatus bool
		psStatus     Plugins
	}

	tests := []struct {
		name       string
		pluginInfo Plugins
		sequential bool
		want       want
	}{
		{
			name:       "No plugins",
			pluginInfo: Plugins{},
			want: want{
				returnStatus: true,
				psStatus:     Plugins{},
			},
		},
		{
			name: "One plugin only with multiple arguments to ExecStart",
			pluginInfo: Plugins{
				{
					Name:        "A/a.test",
					Description: "Applying \"A\" settings",
					ExecStart:   "/usr/bin/top -b -n 1",
				},
			},
			want: want{
				returnStatus: true,
				psStatus: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						ExecStart:   "/usr/bin/top -b -n 1",
						RequiredBy:  []string{},
						Requires:    []string{},
						Status:      "Succeeded",
					},
				},
			},
		},
		{
			name: "One plugin without ExecStart value",
			pluginInfo: Plugins{
				{
					Name:        "A/a.test",
					Description: "Applying \"A\" settings",
					ExecStart:   "",
				},
			},
			want: want{
				returnStatus: true,
				psStatus: Plugins{
					{
						Description: "Applying \"A\" settings",
						Name:        "A/a.test",
						ExecStart:   "",
						RequiredBy:  []string{},
						Requires:    []string{},
						Status:      "Succeeded",
					},
				},
			},
		},
		{
			name: "Only one failing plugin",
			pluginInfo: Plugins{
				{
					Name:        "A/a.test",
					Description: "Applying \"A\" settings",
					ExecStart:   "exit 1",
				},
			},
			want: want{
				returnStatus: false,
				psStatus: Plugins{
					{
						Description: "Applying \"A\" settings",
						Name:        "A/a.test",
						ExecStart:   "exit 1",
						RequiredBy:  []string{},
						Requires:    []string{},
						Status:      "Failed",
					},
				},
			},
		},
		{
			name: "Plugin with dependency",
			pluginInfo: Plugins{
				{
					Name:        "A/a.test",
					Description: "Applying \"A\" settings",
					ExecStart:   `/bin/echo Running A...`,
				},
				{
					Name:        "D/d.test",
					Description: "Applying \"D\" settings",
					Requires:    []string{"A/a.test"},
					ExecStart:   `/bin/echo Running D...`,
				},
			},
			want: want{
				returnStatus: true,
				psStatus: Plugins{
					{
						Description: "Applying \"A\" settings",
						Name:        "A/a.test",
						ExecStart:   `/bin/echo "Running A..."`,
						RequiredBy:  []string{"D/d.test"},
						Requires:    []string{},
						Status:      "Succeeded",
						StdOutErr:   []string{"Running A..."},
					},
					{
						Description: "Applying \"D\" settings",
						Name:        "D/d.test",
						ExecStart:   `/bin/echo "Running D..."`,
						RequiredBy:  []string{},
						Requires:    []string{"A/a.test"},
						Status:      "Succeeded",
						StdOutErr:   []string{"Running D..."},
					},
				},
			},
		},
		{
			name: "Plugin with RequiredBy & Requires circular dependency",
			pluginInfo: Plugins{
				{
					Name:        "D/d.test",
					Description: "Applying \"D\" settings",
					Requires:    []string{"A/a.test"},
					RequiredBy:  []string{"A/a.test"},
					ExecStart:   "/bin/echo \"Running D...!\"",
				},
				{
					Name:        "A/a.test",
					Description: "Applying \"A\" settings",
					ExecStart:   "/bin/echo \"Running A...!\"",
				},
			},
			want: want{
				returnStatus: false,
				psStatus: Plugins{
					{
						Name:        "D/d.test",
						Description: "Applying \"D\" settings",
						Requires:    []string{"A/a.test"},
						RequiredBy:  []string{"A/a.test"},
						ExecStart:   "/bin/echo \"Running D...!\"",
					},
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						ExecStart:   "/bin/echo \"Running A...!\"",
					},
				},
			},
		},
		{
			name: "Plugin with RequiredBy & Requires dependency",
			pluginInfo: Plugins{
				{
					Name:        "D/d.test",
					Description: "Applying \"D\" settings",
					Requires:    []string{"A/a.test"},
					ExecStart:   "/bin/echo \"Running D...!\"",
				},
				{
					Name:        "A/a.test",
					Description: "Applying \"A\" settings",
					RequiredBy:  []string{"D/d.test"},
					ExecStart:   "/bin/echo \"Running A...!\"",
				},
			},
			want: want{
				returnStatus: true,
				psStatus: Plugins{
					{
						Description: "Applying \"A\" settings",
						Name:        "A/a.test",
						ExecStart:   "/bin/echo \"Running A...!\"",
						RequiredBy:  []string{"D/d.test"},
						Requires:    []string{},
						Status:      "Succeeded",
					},
					{
						Description: "Applying \"D\" settings",
						Name:        "D/d.test",
						ExecStart:   "/bin/echo \"Running D...!\"",
						RequiredBy:  []string{},
						Requires:    []string{"A/a.test"},
						Status:      "Succeeded",
					},
				},
			},
		},
		{
			name: "Plugin with RequiredBy dependency",
			pluginInfo: Plugins{
				{
					Name:        "D/d.test",
					Description: "Applying \"D\" settings",
					ExecStart:   "/bin/echo \"Running D...!\"",
				},
				{
					Name:        "A/a.test",
					Description: "Applying \"A\" settings",
					RequiredBy:  []string{"D/d.test"},
					ExecStart:   "/bin/echo \"Running A...!\"",
				},
			},
			want: want{
				returnStatus: true,
				psStatus: Plugins{
					{
						Description: "Applying \"A\" settings", Name: "A/a.test",
						ExecStart:  "/bin/echo \"Running A...!\"",
						RequiredBy: []string{"D/d.test"},
						Requires:   []string{},
						Status:     "Succeeded",
					},
					{
						Description: "Applying \"D\" settings",
						Name:        "D/d.test",
						ExecStart:   "/bin/echo \"Running D...!\"",
						RequiredBy:  []string{},
						Requires:    []string{"A/a.test"},
						Status:      "Succeeded",
					},
				},
			},
		},
		{
			name: "Skip when dependency fails and mark overall status as Failed",
			pluginInfo: Plugins{
				{
					Name:        "A/a.test",
					Description: "Applying \"A\" settings",
					ExecStart:   "exit 1",
				},
				{
					Name:        "D/d.test",
					Description: "Applying \"D\" settings",
					Requires:    []string{"A/a.test"},
					ExecStart:   "/bin/echo \"Running D...!\"",
				},
			},
			want: want{
				returnStatus: false,
				psStatus: Plugins{
					{
						Name:        "A/a.test",
						Description: "Applying \"A\" settings",
						ExecStart:   "exit 1",
						Status:      "Failed",
					},
					{
						Name:        "D/d.test",
						Description: "Applying \"D\" settings",
						ExecStart:   "/bin/echo \"Running D...!\"",
						Requires:    []string{"A/a.test"},
						RequiredBy:  []string{},
						Status:      "Skipped",
					},
				},
			},
		},
	}

	initGraphConfig(config.GetPMLogFile())
	for _, tt := range tests {
		// Test Sequential as well as sequential execution
		for _, tt.sequential = range []bool{false, true} {
			t.Run(tt.name+fmt.Sprintf("(sequential=%v)", tt.sequential),
				func(t *testing.T) {
					res := executePlugins(&tt.pluginInfo, tt.sequential, map[string]string{})
					// t.Logf("res: %+v, expected: %v", res, tt.want.returnStatus)
					if res != tt.want.returnStatus {
						t.Errorf("Return value: got %+v, want %+v",
							res, tt.want.returnStatus)
						return
					}
					// t.Logf("result of all plugins: %+v", tt.pluginInfo)
					for i := range tt.pluginInfo {
						if reflect.DeepEqual(tt.pluginInfo[i].Status,
							tt.want.psStatus[i].Status) == false {
							t.Errorf("Plugins %s Status: got %+v, want %+v",
								tt.pluginInfo[i].Name,
								tt.pluginInfo[i].Status, tt.want.psStatus[i].Status)
						}
						if len(tt.want.psStatus[i].StdOutErr) != 0 &&
							reflect.DeepEqual(tt.pluginInfo[i].StdOutErr,
								tt.want.psStatus[i].StdOutErr) == false {
							t.Errorf("Plugins StdOutErr: got %+v, want %+v",
								tt.pluginInfo[i].StdOutErr,
								tt.want.psStatus[i].StdOutErr)
						}
					}
				},
			)
		}
	}
}

func Test_getPluginsInfoFromJSONStrOrFile(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}

	type args struct {
		jsonStrOrFile string
	}
	tests := []struct {
		name    string
		args    args
		want    RunStatus
		wantErr bool
	}{
		{
			name: "Plugins in JSON String",
			args: args{jsonStrOrFile: `
				{
					"Plugins": [
						{
							"Name": "plugin1",
							"Description": "plugin 1 description...",
							"ExecStart": "echo command to run..."
						},
						{
							"Name": "plugin2",
							"Description": "plugin 2 description...",
							"ExecStart": "echo command to run..."
						},
						{
							"Name": "plugin3",
							"Description": "Plugin 3 depends on 1 and 2",
							"ExecStart": "echo Running plugin 3",
							"Requires": [
								"plugin1",
								"plugin2"
							]
						}
					]
				}`,
			},
			want: RunStatus{
				Plugins: Plugins{
					{
						Name:        "plugin1",
						Description: "plugin 1 description...",
						ExecStart:   "echo command to run...",
					},
					{
						Name:        "plugin2",
						Description: "plugin 2 description...",
						ExecStart:   "echo command to run...",
					},
					{
						Name:        "plugin3",
						Description: "Plugin 3 depends on 1 and 2",
						ExecStart:   "echo Running plugin 3",
						Requires:    []string{"plugin1", "plugin2"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Plugins in JSON file",
			args: args{jsonStrOrFile: "./sample/plugins-prereboot.json"},
			want: RunStatus{
				Plugins: Plugins{
					{
						Name:        "A/a.prereboot",
						Description: "Applying \"A\" settings",
						ExecStart:   "/usr/bin/ls -l -t",
						Requires: []string{
							"C/c.prereboot",
							"D/d.prereboot",
						},
					},
					{
						Name:        "B/b.prereboot",
						Description: "Applying \"B\" settings...",
						ExecStart:   "/bin/echo \"Running B...\"",
					},
					{
						Name:        "C/c.prereboot",
						Description: "Applying \"C\" settings...",
						ExecStart:   "/bin/echo \"Running C...\"",
					},
					{
						Name:        "D/d.prereboot",
						Description: "Applying \"D\" settings...",
						ExecStart:   "/bin/echo 'Running D...!'",
						Requires:    []string{"B/b.prereboot"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPluginsInfoFromJSONStrOrFile(tt.args.jsonStrOrFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPluginsInfoFromJSONStrOrFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPluginsInfoFromJSONStrOrFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_normalizePluginsInfo(t *testing.T) {
	type args struct {
		pluginsInfo Plugins
	}
	tests := []struct {
		name string
		args args
		want Plugins
	}{
		{
			name: "A requires B",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A",
						Description: "Plugin A",
						Requires:    []string{"B"},
					},
					{
						Name:        "B",
						Description: "Plugin B",
					},
				},
			},
			want: Plugins{
				{
					Name:        "A",
					Description: "Plugin A",
					Requires:    []string{"B"},
				},
				{
					Name:        "B",
					Description: "Plugin B",
					RequiredBy:  []string{"A"},
				},
			},
		},
		{
			name: "A requires B with A in later index",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "B",
						Description: "Plugin B",
					},
					{
						Name:        "A",
						Description: "Plugin A",
						Requires:    []string{"B"},
					},
				},
			},
			want: Plugins{
				{
					Name:        "B",
					Description: "Plugin B",
					RequiredBy:  []string{"A"},
				},
				{
					Name:        "A",
					Description: "Plugin A",
					Requires:    []string{"B"},
				},
			},
		},
		{
			name: "B requires A",
			args: args{
				pluginsInfo: Plugins{
					{
						Name:        "A",
						Description: "Plugin A",
					},
					{
						Name:        "B",
						Description: "Plugin B",
						Requires:    []string{"A"},
					},
				},
			},
			want: Plugins{
				{
					Name:        "A",
					Description: "Plugin A",
					RequiredBy:  []string{"B"},
				},
				{
					Name:        "B",
					Description: "Plugin B",
					Requires:    []string{"A"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizePluginsInfo(tt.args.pluginsInfo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizePluginsInfo() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
