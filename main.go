// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build linux

package main

import (
	log "github.com/hashicorp/go-hclog"

	"github.com/hashicorp/nomad/plugins"
	"github.com/sylabs/nomad-driver-singularity/plugin"

	useragent "github.com/sylabs/singularity/pkg/util/user-agent"
)

func main() {
	// Serve the plugin
	plugins.Serve(factory)
}

// factory returns a new instance of a nomad driver plugin
func factory(log log.Logger) interface{} {
	return singularity.NewSingularityDriver(log)
}

// Initialize user agent strings
useragent.InitValue("singularity", "3.0.0")
