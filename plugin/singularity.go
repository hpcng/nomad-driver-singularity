// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package singularity

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// imageExec preloads the taskcnf into args to be apssed to a execCMD
func imageExec(cfg TaskConfig) (argv []string) {
	argv = make([]string, 0, 50)
	if cfg.Debug {
		argv = append(argv, "-d")
	}
	// action can be run/exec
	argv = append(argv, cfg.Command)
	for _, bind := range cfg.Binds {
		argv = append(argv, "--bind", bind)
	}
	for _, sec := range cfg.Security {
		argv = append(argv, "--security", sec)
	}
	if cfg.KeepPrivs {
		argv = append(argv, "--keep-privs")
	}
	if cfg.DropCaps != "" {
		argv = append(argv, "--drop-caps", cfg.DropCaps)
	}
	if cfg.Contain {
		argv = append(argv, "--contain")
	}
	if cfg.NoHome {
		argv = append(argv, "--no-home")
	}
	if cfg.Home != "" {
		argv = append(argv, "--home", cfg.Home)
	}
	for _, fs := range cfg.Overlay {
		argv = append(argv, "--overlay", fs)
	}
	if cfg.Workdir != "" {
		argv = append(argv, "--workdir", cfg.Workdir)
	}
	if cfg.Pwd != "" {
		argv = append(argv, "--pwd", cfg.Pwd)
	}
	if cfg.App != "" {
		argv = append(argv, "--app", cfg.App)
	}
	argv = append(argv, cfg.Image)
	argv = append(argv, cfg.Args...)

	return
}

func getAbsolutePath(bin string) (string, error) {
	lp, err := exec.LookPath(bin)
	if err != nil {
		lp, err = exec.LookPath(CONFIG.singpath)
		if err != nill {
			return "", fmt.Errorf("failed to resolve path to %q executable: %v", bin, err)
		}
	}

	return filepath.EvalSymlinks(lp)
}
