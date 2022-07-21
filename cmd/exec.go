// Copyright (c) 2022, SailPoint Technologies, Inc. All rights reserved.

//go:build linux || darwin || dragonfly || freebsd || netbsd || openbsd
// +build linux darwin dragonfly freebsd netbsd openbsd

package cmd

import (
	"os/exec"
	"syscall"
)

// ExecCommand runs commands on non windows environment with Setpgid flag set to true
func ExecCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd.Start()
}
