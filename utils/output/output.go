// Copyright (c) 2023 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package output

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// cmdOptions contains commandline parameters/options for generating output in specified format.
var cmdOptions struct {
	// File indicates the file name to write the results.
	File string

	// Format indicates the output format to write the results.
	//  Supported formats are "json", "yaml".
	Format string
}

// GetFile returns the output file name that is currently set.
func GetFile() string {
	return cmdOptions.File
}

// GetFormat returns the output format type that is currently set.
func GetFormat() string {
	return cmdOptions.Format
}

// RegisterCommandOptions registers the command options related to the output options.
func RegisterCommandOptions(f *flag.FlagSet, defaultParams map[string]string) {
	defaultOutputFile, ok := defaultParams["output-file"]
	if !ok {
		defaultOutputFile = ""
	}
	defaultOutputFormat, ok := defaultParams["output-format"]
	if !ok {
		defaultOutputFormat = ""
	}
	f.StringVar(
		&cmdOptions.File,
		"output-file",
		defaultOutputFile,
		"Name of the file to write the results.",
	)
	f.StringVar(
		&cmdOptions.Format,
		"output-format",
		defaultOutputFormat,
		"The format of output to display the results. "+
			"Supported output formats are 'json', 'yaml'.",
	)
}

// Write writes the given data in the format {json|yaml} that was set in options.
//
//	into a specified file. If file is not specified, then it will print
//	on STDOUT.
func Write(data interface{}) error {
	log.Println("Entering Write")
	defer log.Println("Exiting Write")

	if cmdOptions.Format == "" {
		// log.Printf("Skipping the Write() as output format is not set.")
		return nil
	}
	log.Printf("Writing output in %s to file name: %s",
		cmdOptions.Format, cmdOptions.File)
	return writeToFile(data, cmdOptions.Format, cmdOptions.File)
}

// writeToFile writes the given data in the specified format {json|yaml}
//
//	into a specified file. If file is not specified, then it will print
//	on STDOUT.
func writeToFile(data interface{}, format string, filePath string) error {
	log.Println("Entering writeToFile")
	defer log.Println("Exiting writeToFile")

	var err error
	var out []byte

	if format == "json" {
		out, err = json.MarshalIndent(data, "", "  ")
	} else {
		out, err = yaml.Marshal(data)
	}
	if err != nil {
		log.Printf("Unable to marshal %s data into %s. Error: %v",
			data, format, err)
		return err
	}

	if filePath == "" {
		// fmt.Println("File name to write the status is not specified.",
		// 	"Outputting to console.")
		fmt.Println(string(out))
		return nil
	}

	filePath = filepath.FromSlash(filePath)
	log.Printf("Output file: %s\n", filePath)
	err = ioutil.WriteFile(filePath, out, 0764)
	if err != nil {
		log.Printf("Unable to write to specified file %s. Error: %v",
			filePath, err)
		return err
	}

	return nil
}
