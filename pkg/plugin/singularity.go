// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package singularity

import (
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
)

const (
	// defaultFailedCode for singularity runtime
	defaultFailedCode = 255
)

type syexec struct {
	argv         []string
	cmd          *exec.Cmd
	taskConfig   TaskConfig
	cfg          *drivers.TaskConfig
	stdout       io.WriteCloser
	stderr       io.WriteCloser
	env          []string
	TaskDir      string
	state        *psState
	containerPid int
	exitCode     int
	ExitError    error
	logger       hclog.Logger
}

type psState struct {
	Pid      int
	ExitCode int
	Signal   int
	Time     time.Time
}

func (s *syexec) startContainer(commandCfg *drivers.TaskConfig) error {
	s.logger.Debug("launching command", strings.Join(s.argv, " "))

	cmd := exec.Command(singularityCmd, s.argv...)

	// set the writers for stdout and stderr
	stdout, err := s.Stdout()
	if err != nil {
		return err
	}
	stderr, err := s.Stderr()
	if err != nil {
		return err
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// set the task dir as the working directory for the command
	cmd.Dir = commandCfg.TaskDir().Dir
	cmd.Path = singularityCmd
	cmd.Args = append([]string{cmd.Path}, s.argv...)
	cmd.Env = s.env

	// Start the process
	if err := cmd.Run(); err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			s.exitCode = ws.ExitStatus()
		} else {
			s.logger.Error("Could not get exit code for failed program: ", "singularity", s.argv)
			s.exitCode = defaultFailedCode
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		s.exitCode = ws.ExitStatus()
	}

	s.cmd = cmd

	s.state = &psState{Pid: s.cmd.Process.Pid, ExitCode: s.exitCode, Time: time.Now()}
	return nil
}

// waitTillStopped blocks and returns true when container exit;
// returns false with an error message if the container processes cannot be identified.
// func (s *syexec) waitTillStopped() (bool, error) {
// 	ps, err := os.FindProcess(s.containerPid)
// 	if err != nil {
// 		return false, err
// 	}

// 	for {
// 		if err := ps.Signal(syscall.Signal(0)); err != nil {
// 			return true, nil
// 		}
// 	}
// }
