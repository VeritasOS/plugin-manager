// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9
package config

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	logger "github.com/VeritasOS/plugin-manager/utils/log"
)

func initTestLogging() {
	myLogFile := "pm.log"
	if GetPMLogFile() != "" {
		myLogFile = GetPMLogFile()
	}
	if GetPMLogDir() != "" {
		myLogFile = GetPMLogDir() + myLogFile
	}
	errList := logger.DeInitLogger()
	if len(errList) > 0 {
		fmt.Printf("Failed to deinitialize logger, err=[%v]", errList)
		os.Exit(-1)
	}
	err := logger.InitFileLogger(myLogFile, "INFO")
	if err != nil {
		fmt.Printf("Failed to initialize logger, err=[%v]", err)
		os.Exit(-1)
	}
}

func init() {
	initTestLogging()
	// EnvConfFile is environment variable containing config file path.
	// NOTE: "PM_CONF_FILE" value would be set in Makefile.
	EnvConfFile = "PM_CONF_FILE"
}

func Test_Load(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}

	// Restore env value so that other tests won't be affected.
	restoreEnvConfFile := os.Getenv(EnvConfFile)
	defer os.Setenv(EnvConfFile, restoreEnvConfFile)
	defer t.Logf("Restoring EnvConfFile value from %s to %s\n", EnvConfFile, restoreEnvConfFile)

	type args struct {
		EnvConfFile string
	}
	type PluginManager struct {
		Library string `yaml:"library"`
		LogDir  string `yaml:"log dir"`
		LogFile string `yaml:"log file"`
	}
	tests := []struct {
		name string
		args args
		want Config
	}{
		{
			name: "Valid pm.config file",
			args: args{
				EnvConfFile: "../sample/pm.config.yaml",
			},
			want: Config{
				PluginManager: struct {
					Library string "yaml:\"library\""
					LogDir  string "yaml:\"log dir\""
					LogFile string "yaml:\"log file\""
				}{
					Library: "../sample/library",
					LogDir:  "./",
					LogFile: "pm",
				},
			},
		},
		{
			name: "Non existing pm.config file",
			args: args{
				EnvConfFile: "non-existing/pm.config.yaml",
			},
			want: Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Setenv(EnvConfFile, tt.args.EnvConfFile); nil != err {
				t.Errorf("Failed to set environment variable \"EnvConfFile\". Error: %s",
					err.Error())
			}
			t.Logf("EnvConfFile: '%s'\n", os.Getenv(EnvConfFile))
			Load()
			got := myConfig
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readConfigFile(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "RUNNING" {
		t.Skip("Not applicable while running integration tests.")
		return
	}

	type args struct {
		confFilePath string
	}
	type PluginManager struct {
		Library   string `yaml:"library"`
		LogDir    string `yaml:"log dir"`
		LogFile   string `yaml:"log file"`
		PluginDir string `yaml:"plugin dir"`
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{
			name: "Valid conf file",
			args: args{
				confFilePath: "../sample/pm.config.yaml",
			},
			want: Config{
				PluginManager: struct {
					Library string "yaml:\"library\""
					LogDir  string "yaml:\"log dir\""
					LogFile string "yaml:\"log file\""
				}{
					Library: "../sample/library",
					LogDir:  "./",
					LogFile: "pm",
				},
			},
			wantErr: false,
		},
		{
			name: "Non existing conf file",
			args: args{
				confFilePath: "non-existing-dir/pm.config.yaml",
			},
			want: Config{
				PluginManager: struct {
					Library string "yaml:\"library\""
					LogDir  string "yaml:\"log dir\""
					LogFile string "yaml:\"log file\""
				}{},
			},
			wantErr: true,
		},
		{
			name: "Invalid conf file",
			args: args{
				confFilePath: "../../sample/library/D/preupgrade.sh",
			},
			want: Config{
				PluginManager: struct {
					Library string "yaml:\"library\""
					LogDir  string "yaml:\"log dir\""
					LogFile string "yaml:\"log file\""
				}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readConfigFile(tt.args.confFilePath)
			if (err != nil) != tt.wantErr {
				t.Log("Error:", err.Error())
				t.Errorf("readConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
