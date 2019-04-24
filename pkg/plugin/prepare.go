// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package singularity

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hashicorp/nomad/client/lib/fifo"
	"github.com/hashicorp/nomad/plugins/drivers"
)

// prepareContainer preloads the taskcnf into args to be apssed to a execCmd
func prepareContainer(d *Driver, cfg *drivers.TaskConfig, taskCfg TaskConfig) syexec {
	argv := make([]string, 0, 50)
	var se syexec
	se.taskConfig = taskCfg
	se.cfg = cfg
	se.env = cfg.EnvList()

	// global flags
	if taskCfg.Debug {
		argv = append(argv, "-d")
	}
	if taskCfg.Verbose {
		argv = append(argv, "-v")
	}
	// action can be run/exec
	argv = append(argv, taskCfg.Command)
	for _, bind := range taskCfg.Binds {
		argv = append(argv, "--bind", bind)
	}
	for _, sec := range taskCfg.Security {
		argv = append(argv, "--security", sec)
	}
	if taskCfg.KeepPrivs {
		argv = append(argv, "--keep-privs")
	}
	if taskCfg.DropCaps != "" {
		argv = append(argv, "--drop-caps", taskCfg.DropCaps)
	}
	if taskCfg.Contain {
		argv = append(argv, "--contain")
	}
	if taskCfg.NoHome {
		argv = append(argv, "--no-home")
	}
	if taskCfg.Home != "" {
		argv = append(argv, "--home", taskCfg.Home)
	}
	for _, fs := range taskCfg.Overlay {
		argv = append(argv, "--overlay", fs)
	}
	if taskCfg.Workdir != "" {
		argv = append(argv, "--workdir", taskCfg.Workdir)
	}
	if taskCfg.Pwd != "" {
		argv = append(argv, "--pwd", taskCfg.Pwd)
	}
	if taskCfg.App != "" {
		argv = append(argv, "--app", taskCfg.App)
	}
	argv = append(argv, taskCfg.Image)
	se.argv = append(argv, taskCfg.Args...)

	return se
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// Stdout returns a writer for the configured file descriptor
func (s *syexec) Stdout() (io.WriteCloser, error) {
	if s.stdout == nil {
		if s.cfg.StdoutPath != "" {
			f, err := fifo.OpenWriter(s.cfg.StdoutPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create stdout: %v", err)
			}
			s.stdout = f
		} else {
			s.stdout = nopCloser{ioutil.Discard}
		}
	}
	return s.stdout, nil
}

// Stderr returns a writer for the configured file descriptor
func (s *syexec) Stderr() (io.WriteCloser, error) {
	if s.stderr == nil {
		if s.cfg.StderrPath != "" {
			f, err := fifo.OpenWriter(s.cfg.StderrPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create stderr: %v", err)
			}
			s.stderr = f
		} else {
			s.stderr = nopCloser{ioutil.Discard}
		}
	}
	return s.stderr, nil
}

func (s *syexec) Close() {
	if s.stdout != nil {
		s.stdout.Close()
	}
	if s.stderr != nil {
		s.stderr.Close()
	}
}
