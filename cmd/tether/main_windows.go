// Copyright 2016 VMware, Inc. All Rights Reserved.
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
	_ "net/http/pprof"
	"os"
	"runtime/debug"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/vmware/vic/pkg/vsphere/extraconfig"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("run time panic: %s : %s", r, debug.Stack())
		}
		halt()
	}()

	// where to look for the various devices and files related to tether
	pathPrefix = "com://"

	if strings.HasSuffix(os.Args[0], "-debug") {
		extraconfig.DecodeLogLevel = log.DebugLevel
		extraconfig.EncodeLogLevel = log.DebugLevel
		log.SetLevel(log.DebugLevel)
	}

	// the OS ops and utils to use
	win := &osopsWin{}
	ops = win
	utils = win

	// get the windows service logic running so that we can play well in that mode
	runService("VMware Tether", false)

	server = &attachServerSSH{}
	src, err := extraconfig.GuestInfoSource()
	if err != nil {
		log.Error(err)
		return
	}

	sink, err := extraconfig.GuestInfoSink()
	if err != nil {
		log.Error(err)
		return
	}

	err = run(src, sink)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Clean exit from tether")
}

// exit reports completion detail in persistent fashion then cleanly
// shuts down the system
func halt() {
	log.Infof("Powering off the system")
	// TODO: windows fast halt command
}
