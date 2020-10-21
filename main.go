// Copyright © 2020 Bohdan Mushkevych
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"flag"
	"fmt"
	"github.com/mushkevych/9ofm/commander/configuration"
	"io/ioutil"
	"os"

	"github.com/mushkevych/9ofm/commander"
	log "github.com/sirupsen/logrus"
)

var (
	flgVersion bool
	alphaRoot string
	betaRoot string

	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

func initLogging() {
	var logFileObj *os.File
	var err error

	if configuration.Config.GetBool("log.enabled") {
		logFileObj, err = os.OpenFile(configuration.Config.GetString("log.path"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		log.SetOutput(logFileObj)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	Formatter := new(log.TextFormatter)
	Formatter.DisableTimestamp = true
	log.SetFormatter(Formatter)

	level, err := log.ParseLevel(configuration.Config.GetString("log.level"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	log.SetLevel(level)
	log.Debug("Starting 9ofm...")
	//log.Debugf("config filepath: %s", config.Config.ConfigFileUsed())
	//for k, v := range config.Config.AllSettings() {
	//	log.Debug("config value: ", k, " : ", v)
	//}
}

func main() {
	flag.BoolVar(&flgVersion, "version", false, "Version of the 9ofm")
	flag.StringVar(&alphaRoot, "a", "/", "Starting path for panel A")
	flag.StringVar(&betaRoot, "b", "/", "Starting path for panel B")
	flag.Parse()

	if flgVersion {
		fmt.Printf("Build on %s from sha1 %s\n", buildTime, sha1ver)
		os.Exit(0)
	}

	initLogging()
	commander.Start()
}
