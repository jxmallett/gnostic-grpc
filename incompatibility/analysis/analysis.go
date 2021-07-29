// Copyright 2021 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/googleapis/gnostic-grpc/incompatibility"
	"github.com/googleapis/gnostic-grpc/utils"
)

func main() {
	if len(os.Args) != 2 {
		exitIfError(errors.New("argument should be a path to a directory"))
	}
	runAnalysis(os.Args[1])
}

// runs analysis on given directory
func runAnalysis(dirPath string) {
	analysisAggregation := incompatibility.NewAnalysis()
	readingDirectoryErr := filepath.WalkDir(os.Args[1], func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("walk error for file at %s", path)
		}
		newAnalysis, analysisErr := router(path, d, err)
		if analysisErr != nil {
			log.Printf("unable to produce analysis for file %s with error <%s>", path, analysisErr.Error())
		}
		analysisAggregation = incompatibility.AggregateAnalysis(analysisAggregation, newAnalysis)
		return nil
	})
	if readingDirectoryErr != nil {
		exitIfError(errors.New("unable to walk through directory"))
	}
}

// router directs logic for either a file or a directory to produce an analysis in either case
func router(path string, d fs.DirEntry, err error) (*incompatibility.ApiSetIncompatibility, error) {
	if err != nil {
		return nil, err
	}
	if d.IsDir() {
		return directoryHandler(path)
	}
	singleFileReport, reportErr := fileHandler(path)
	if reportErr != nil {
		return nil, reportErr
	}
	return incompatibility.FileReport2Analysis(singleFileReport), nil
}

// fileHander attempts to parse the file at path and to then create an incompatibility report
func fileHandler(path string) (*incompatibility.IncompatibilityReport, error) {
	openAPIDoc, err := utils.ParseOpenAPIDoc(path)
	if err != nil {
		return nil, err
	}
	incompatibilityReport := incompatibility.ScanIncompatibilities(openAPIDoc)
	log.Printf("created incompatibility report for file at %s\n", path)
	return incompatibilityReport, nil
}

// TODO
func directoryHandler(dirPath string) (*incompatibility.ApiSetIncompatibility, error) {
	return nil, nil
}

func exitIfError(e error) {
	if e == nil {
		return
	}
	log.Printf("Exiting with error: %s\n", e)
	os.Exit(1)
}