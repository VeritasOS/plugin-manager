// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/VeritasOS/plugin-manager/v2/config"
	"github.com/VeritasOS/plugin-manager/v2/pluginmanager"

	yaml "gopkg.in/yaml.v3"
)

// Config is Plugin Manager's configuration information.
type Config struct {
	// PluginManager configuration information.
	PluginManager struct {
		// Library is the path where plugin directories containing plugin files are present.
		Library string `yaml:"library"`
		LogDir  string `yaml:"log dir"`
		LogFile string `yaml:"log file"`
		// PluginDir is deprecated. Use Library instead.
		PluginDir string `yaml:"plugin dir"`
	}
}

func saveConfig(newConfig Config, configFile string) error {
	log.Println("Entering saveConfig")
	defer log.Println("Exiting saveConfig")

	log.Printf("config file: %s\n", configFile)
	out, err := yaml.Marshal(newConfig)
	if err != nil {
		log.Printf("Failed to marshal plugin config: %+v\n", newConfig)
		return err
	}
	ioutil.WriteFile(configFile, out, os.FileMode(0644))
	if err != nil {
		log.Printf("Failed to write %s file\n", configFile)
		return err
	}
	return nil
}

func setIntegrationEnvironment(topPath string) string {
	log.Println("Entering setIntegrationEnvironment")
	defer log.Println("Exiting setIntegrationEnvironment")

	configFile := filepath.FromSlash(topPath + "/pm.config-integ.yaml")

	var newConfig Config
	newConfig.PluginManager.Library = filepath.FromSlash(topPath + "/sample/library")
	newConfig.PluginManager.LogDir = filepath.FromSlash(topPath)
	newConfig.PluginManager.LogFile = "pm-integ"

	saveConfig(newConfig, configFile)
	os.Setenv(config.EnvConfFile, configFile)

	return configFile
}

// setDeprecatedIntegrationEnvironment() sets the values that are now deprecated and tests to make sure that
// there are no regressions until they are removed.
// TODO: Delete this once PluginDir support is removed.
func setDeprecatedIntegrationEnvironment(topPath string) string {
	log.Println("Entering setIntegrationEnvironment")
	defer log.Println("Exiting setIntegrationEnvironment")

	configFile := filepath.FromSlash(topPath + "/pm.config-integ.yaml")

	var newConfig Config
	newConfig.PluginManager.LogDir = filepath.FromSlash(topPath)
	newConfig.PluginManager.LogFile = "pm-integ"
	newConfig.PluginManager.PluginDir = filepath.FromSlash(topPath + "/sample/library")

	saveConfig(newConfig, configFile)
	os.Setenv(config.EnvConfFile, configFile)

	return configFile
}

// REFERENCE: https://www.cyphar.com/blog/post/20170412-golang-integration-coverage
func TestIntegration(t *testing.T) {
	var (
		cmdArgs []string
		run     bool
	)

	for _, arg := range os.Args {
		switch {
		case arg == "__DEVEL--integration-tests":
			run = true
		case strings.HasPrefix(arg, "-test"):
		case strings.HasPrefix(arg, "__DEVEL"):
		default:
			cmdArgs = append(cmdArgs, arg)
		}
	}
	os.Args = cmdArgs

	if run {
		main()
		return
	}

	tDir := os.Getenv("INTEG_TEST_BIN")
	if tDir == "" {
		t.Error("The integration test should have \"INTEG_TEST_BIN\" " +
			"set to the plugin manager location.\n")
		return
	}
	t.Log("INTEG_TEST_BIN:", tDir)
	pmBinary := filepath.FromSlash(tDir + "/pm")

	if os.Getenv("INTEGRATION_TEST") == "START" {
		binCmd := exec.Command("go", "test", "-v", "-covermode=count", "-c", "-o", pmBinary)
		stdOutErr, err := binCmd.CombinedOutput()
		t.Log("Stdout & Stderr:", string(stdOutErr))
		t.Log("Error:", err)
		os.Setenv("INTEGRATION_TEST", "RUNNING")
	}

	oriConfigFile := os.Getenv(config.EnvConfFile)
	configFile := setIntegrationEnvironment(tDir)
	defer os.Remove(configFile)
	defer os.Setenv(config.EnvConfFile, oriConfigFile)

	integTest(t, pmBinary, tDir, "")

	// setIntegrationEnvironment() for deprecated  scenario
	setDeprecatedIntegrationEnvironment(tDir)
	integTest(t, pmBinary, tDir, " deprecated")

	os.Setenv("INTEGRATION_TEST", "DONE")
	os.Remove(pmBinary)
}

func integTest(t *testing.T, pmBinary, tDir, deprecated string) {
	type args struct {
		pluginType           string
		sequential           bool
		testPluginExitStatus int
	}

	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Precheck",
			args: args{
				pluginType: "preupgrade",
			},
			want: []string{
				"",
				"Checking for \"D\" settings...: " + pluginmanager.DStatusStart,
				"Checking for \"D\" settings...: " + pluginmanager.DStatusOk,
				"",
				"Checking for \"A\" settings: " + pluginmanager.DStatusStart,
				"Checking for \"A\" settings: " + pluginmanager.DStatusOk,
				"Running preupgrade plugins: " + pluginmanager.DStatusOk,
			},
			wantErr: false,
		},
		{
			name: "Skip when dependency fail",
			args: args{
				pluginType:           "preupgrade",
				testPluginExitStatus: 1,
			},
			want: []string{
				"",
				"Checking for \"D\" settings...: " + pluginmanager.DStatusStart,
				"Checking for \"D\" settings...: " + pluginmanager.DStatusFail,
				"",
				"Checking for \"A\" settings: " + pluginmanager.DStatusStart,
				"Checking for \"A\" settings: " + pluginmanager.DStatusSkip,
				"Running preupgrade plugins: " + pluginmanager.DStatusFail,
				"",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		for _, tc.args.sequential = range []bool{false, true} {
			t.Run(tc.name+fmt.Sprintf("(sequential=%v)", tc.args.sequential)+fmt.Sprintf("%v", deprecated), func(t *testing.T) {
				cmdStr := pmBinary
				tmpfile, err := ioutil.TempFile(tDir+"/cover/", "cover.out")
				if err != nil {
					t.Fatal(err)
				}
				var cmdParams []string
				cmdParams = append(cmdParams, "-test.coverprofile="+tmpfile.Name())
				cmdParams = append(cmdParams, "__DEVEL--integration-tests")
				if tc.args.pluginType != "" {
					cmdParams = append(cmdParams, "run")
					cmdParams = append(cmdParams, "-type")
					cmdParams = append(cmdParams, tc.args.pluginType)
				}
				// TODO: Update test cases & output to handle sequential execution.
				cmdParams = append(cmdParams, "-sequential="+strconv.FormatBool(tc.args.sequential))
				if tc.args.testPluginExitStatus == 1 {
					t.Logf("Setting testPluginExitStatus to %d\n", tc.args.testPluginExitStatus)
					os.Setenv("TEST_PLUGIN_EXIT_STATUS", "1")
				} else {
					os.Setenv("TEST_PLUGIN_EXIT_STATUS", "0")
				}

				t.Logf("EnvConfFile: %s", os.Getenv(config.EnvConfFile))
				t.Logf("Command: %+v; Params: %+v\n", cmdStr, cmdParams)
				cmd := exec.Command(cmdStr, cmdParams...)
				stdOutErr, err := cmd.CombinedOutput()
				t.Log("Stdout & Stderr:", string(stdOutErr))
				got := strings.Split(string(stdOutErr), "\n")
				for len(got) > 1 && strings.Contains(got[0], "Log: ") {
					got = got[1:]
				}
				if len(got) >= 3 && strings.Contains(got[len(got)-3], "PASS") {
					got = got[:len(got)-3]
					t.Logf("After PASS comparison: %+v\n", got)
				} else {
					t.Logf("Didn't match PASS in %d: %+v\n", len(got), got[len(got)-1])
				}
				for m := range got {
					t.Logf("%d ---> %s\n", m, got[m])
				}

				if (err != nil) != tc.wantErr {
					t.Logf("Error: %v", err)
					t.Errorf("%s %s: error = %v, wantErr %v",
						cmdStr, cmdParams, err, tc.wantErr)
				}
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("got %v, want %v", got, tc.want)
				}

			})
		}
	}
}
