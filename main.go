// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// +build linux

package main

import (
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins"
	"github.com/sylabs/nomad-plugin-singularity/plugin"
)

func main() {
	plugins.Serve(factory)
}

// factory returns a new instance of the Singularity Driver plugin
func factory(log hclog.Logger) interface{} {
	return plugin.NewDriver(log)
}
